// Taken from:
// https://raw.githubusercontent.com/coreos/go-systemd/dde89c25321b9584309f75e2a42b3473624ed8c7/sdjournal/read.go

// Copyright 2015 RedHat, Inc.
// Copyright 2015 CoreOS, Inc.
// + Copyright 2016 Gorka Lerchundi Osa.

package core

import (
	"io"
	"log"
	"time"

	"github.com/glerchundi/go-systemd/sdjournal"
)

// JournalFollowerConfig represents options to drive the behavior of a JournalFollower.
type JournalFollowerConfig struct {
	// Start relative to the cursor
	Cursor  string

	// Show only journal entries whose fields match the supplied values. If
	// the array is empty, entries will not be filtered.
	Matches []sdjournal.Match

	// If not empty, the journal instance will point to a journal residing
	// in this directory. The supplied path may be relative or absolute.
	Path string
}

// JournalFollower is an io.ReadCloser which provides a simple interface for iterating through the
// systemd journal.
type JournalFollower struct {
	journal *sdjournal.Journal
}

// NewJournalFollower creates a new JournalFollower with configuration options that are similar to the
// systemd journalctl tool's iteration and filtering features.
func NewJournalFollower(config JournalFollowerConfig) (*JournalFollower, error) {
	r := &JournalFollower{}

	// Open the journal
	var err error
	if config.Path != "" {
		r.journal, err = sdjournal.NewJournalFromDir(config.Path)
	} else {
		r.journal, err = sdjournal.NewJournal()
	}
	if err != nil {
		return nil, err
	}

	// Add any supplied matches
	for _, m := range config.Matches {
		r.journal.AddMatch(m.String())
	}

	if config.Cursor != "" {
		// Start based on a custom cursor
		if err := r.journal.SeekCursor(config.Cursor); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *JournalFollower) Close() error {
	return r.journal.Close()
}

func (r *JournalFollower) Follow(recvc chan<- *sdjournal.JournalEntry,
                               stopc <-chan time.Time,
                               donec chan bool,
                               errc chan<- error) {
	defer close(donec)

	// Process journal entries and events. Entries are flushed until the tail or
	// timeout is reached, and then we wait for new events or the timeout.
	process:
	for {
		e, err := r.readEntry()
		if err != nil && err != io.EOF {
			break process
		}

		select {
		case <-stopc:
			return
		default:
			if e != nil {
				recvc <- e
				continue process
			}
		}

		// We're at the tail, so wait for new events or time out.
		// Holds journal events to process. Tightly bounded for now unless there's a
		// reason to unblock the journal watch routine more quickly.
		events := make(chan int, 1)
		pollDone := make(chan bool, 1)
		go func() {
			for {
				select {
				case <-pollDone:
					return
				default:
					events <- r.journal.Wait(time.Duration(1) * time.Second)
				}
			}
		}()

		select {
		case <-stopc:
			pollDone <- true
			return
		case e := <-events:
			pollDone <- true
			switch e {
			case sdjournal.SD_JOURNAL_NOP, sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
				// TODO: need to account for any of these?
			default:
				log.Printf("Received unknown event: %d\n", e)
			}
			continue process
		}
	}

	return
}

func (r *JournalFollower) readEntry() (*sdjournal.JournalEntry, error) {
	// Advance the journal cursor
	c, err := r.journal.Next()

	// An unexpected error
	if err != nil {
		return nil, err
	}

	// EOF detection
	if c == 0 {
		return nil, io.EOF
	}

	return r.journal.GetEntry()
}