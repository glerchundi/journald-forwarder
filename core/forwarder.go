package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/glerchundi/go-systemd/sdjournal"
	"github.com/glerchundi/journald-forwarder/core/ring"
)

type ForwarderConfig struct {
	Provider Provider
	CursorPath string
}

type Forwarder struct {
	jf *JournalFollower
	p Provider
	publishAtLeastFreq time.Duration

	ring *ring.Ring

	cursorc chan string
	cursorPath string
	cursorFlushFreq time.Duration

	recvc chan *sdjournal.JournalEntry
	stopc <-chan time.Time
	donec chan bool
	errc chan error
}

func NewForwarder(config ForwarderConfig) (*Forwarder, error) {
	if config.Provider == nil {
		return nil, fmt.Errorf("provide a forwarding publisher")
	}

	cursorPath := "./cursor"
	if config.CursorPath != "" {
		cursorPath = config.CursorPath
	}

	cursor := ""
	if _, err := os.Stat(cursorPath); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(cursorPath)
		if err != nil {
			// TODO: fail here?!
			return nil, err
		}
		cursor = string(data)
	} else {
		err = os.MkdirAll(filepath.Dir(cursorPath), 0755)
		if err != nil {
			return nil, err
		}
	}

	jf, err := NewJournalFollower(JournalFollowerConfig{
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}

	return &Forwarder{
		jf: jf,
		p: config.Provider,
		publishAtLeastFreq: 5 * time.Second,

		ring: ring.NewRing(config.Provider.GetBulkSize()),

		cursorc: make(chan string),
		cursorPath: cursorPath,
		cursorFlushFreq: 1 * time.Second,

		recvc: make(chan *sdjournal.JournalEntry, 1),
		stopc: make(<-chan time.Time),
		donec: make(chan bool),
		errc: make(chan error),
	}, nil
}

func (f *Forwarder) forward() {
	defer close(f.donec)

	tduration := 10 * time.Second
	timer := time.NewTimer(tduration)
	for {
		select {
		case <- timer.C:
			f.publish(true)
		case e := <-f.recvc:
			f.ring.Enqueue(e)
			f.publish(false)
		case <-f.stopc:
			return
		}

		if !timer.Reset(f.publishAtLeastFreq) {
			timer = time.NewTimer(f.publishAtLeastFreq)
		}
	}
}

func (f *Forwarder) publish(force bool) {
	if f.ring.Len() == f.p.GetBulkSize() || force {
		errorOccurred := true
		for errorOccurred {
			n, err := f.p.Publish(f.ring.Iterator())
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

func (f *Forwarder) Run() {
	// 1.- start following
	go f.jf.Follow(f.recvc, f.stopc, f.donec, f.errc)

	// 2.- start forwarding
	go f.forward()

	// 3.- persist cursor
	go f.cursorPersist(f.cursorFlushFreq)

	// wait for signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-f.errc:
			os.Stderr.Write([]byte(err.Error()))
		case s := <-signalChan:
			log.Print("Captured %v. Exiting...", s)
			close(f.donec)
		case <-f.donec:
			os.Exit(0)
		}
	}
}