package core

import "github.com/glerchundi/go-systemd/sdjournal"

type JournalEntryIterator interface {
	Next() bool
	Value() (int, *sdjournal.JournalEntry)
}