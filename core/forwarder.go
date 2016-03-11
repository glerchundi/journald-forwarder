package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/glerchundi/go-systemd/sdjournal"
	"github.com/glerchundi/journald-forwarder/core/ring"
)

type ForwarderConfig struct {
	RingSize     int
	Path         string
	ForwardFlush time.Duration
	CursorPath   string
	CursorFlush  time.Duration
}

func NewForwarderConfig(ringSize int) ForwarderConfig {
	return ForwarderConfig{
		RingSize:     ringSize,
		Path:         "/var/log/journal",
		ForwardFlush: 5 * time.Second,
		CursorPath:   "/var/run/journald-forwarder/cursor",
		CursorFlush:  1 * time.Second,
	}
}

type Forwarder struct {
	follower     *JournalFollower
	forwardFlush time.Duration

	ring         *ring.Ring

	cursorc      chan string
	cursorPath   string
	cursorFlush  time.Duration

	recvc        chan *sdjournal.JournalEntry
	stopc        chan time.Time
	donec        chan bool
	errc         chan error
}

func NewForwarder(config ForwarderConfig) (*Forwarder, error) {
	// Create cursor file
	cursor := ""
	if _, err := os.Stat(config.CursorPath); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(config.CursorPath)
		if err != nil {
			// TODO: fail here?!
			return nil, err
		}
		cursor = string(data)
	} else {
		err = os.MkdirAll(filepath.Dir(config.CursorPath), 0755)
		if err != nil {
			return nil, err
		}
	}

	// Open journal
	jf, err := NewJournalFollower(JournalFollowerConfig{
		Cursor: cursor,
		Path: config.Path,
	})
	if err != nil {
		return nil, err
	}

	// Create forwarder
	return &Forwarder{
		follower: jf,
		forwardFlush: config.ForwardFlush,

		ring: ring.NewRing(config.RingSize),

		cursorc: make(chan string),
		cursorPath: config.CursorPath,
		cursorFlush: config.CursorFlush,

		recvc: make(chan *sdjournal.JournalEntry, 1),
		stopc: make(chan time.Time),
		donec: make(chan bool),
		errc:  make(chan error),
	}, nil
}

func (f *Forwarder) forward(provider Provider) {
	defer close(f.donec)

	tduration := 10 * time.Second
	timer := time.NewTimer(tduration)
	for {
		select {
		case <- timer.C:
			f.publish(provider, true)
		case e := <-f.recvc:
			f.ring.Enqueue(e)
			f.publish(provider, false)
		case <-f.stopc:
			return
		}

		if !timer.Reset(f.forwardFlush) {
			timer = time.NewTimer(f.forwardFlush)
		}
	}
}

func (f *Forwarder) publish(provider Provider, force bool) {
	if f.ring.Len() == f.ring.Capacity() || force {
		errorOccurred := true
		for errorOccurred {
			n, err := provider.Publish(f.ring.Iterator())
			if err != nil {
				f.errc <- err
				time.Sleep(1 * time.Second)
			}
			for i := 0; i < n; i++ {
				e := f.ring.Dequeue()
				if i+1 == n {
					f.cursorc <- e.Cursor
				}
			}
			errorOccurred = false
		}
	}
}

func (f *Forwarder) cursorPersist(flushFreq time.Duration) {
	defer close(f.donec)

	currentCursor := ""
	ticker := time.NewTicker(flushFreq)
	for {
		select {
		case <- ticker.C:
			if currentCursor == "" {
				break
			}
			if err := f.writeCursor(currentCursor); err != nil {
				f.errc <- err
				time.Sleep(1 * time.Second)
			}
		case c := <-f.cursorc:
			currentCursor = c
		case <-f.stopc:
			return
		}
	}
}

func (f *Forwarder) writeCursor(cursor string) error {
	tempFile, err := ioutil.TempFile(filepath.Dir(f.cursorPath), "." + filepath.Base(f.cursorPath))
	if err != nil {
		return err
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	_, err = tempFile.WriteString(cursor)
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), f.cursorPath)
	if err != nil {
		return err
	}

	return nil
}

func (f *Forwarder) Run(provider Provider) {
	// 1.- Start following
	go f.follower.Follow(f.recvc, f.stopc, f.donec, f.errc)

	// 2.- Start forwarding
	go f.forward(provider)

	// 3.- Persist cursor
	go f.cursorPersist(f.cursorFlush)
}