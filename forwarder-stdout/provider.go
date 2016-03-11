package main

import (
	"os"
	"github.com/glerchundi/journald-forwarder/core"
)

type StdoutProviderConfig struct {
}

func NewStdoutProvider(config StdoutProviderConfig) (*StdoutProvider, error) {
	return &StdoutProvider{core.JournalEntryMarshaller{}}, nil
}

type StdoutProvider struct {
	entryMarshaller core.JournalEntryMarshaller
}

func (sp *StdoutProvider) GetBulkSize() int {
	return 1
}

func (sp *StdoutProvider) Publish(iterator core.JournalEntryIterator) (int, error) {
	index := 0
	for iterator.Next() {
		i, e := iterator.Value()
		sp.entryMarshaller.Marshal(e)
		os.Stdout.WriteString(sp.entryMarshaller.String())
		os.Stdout.Write([]byte{'\n','\n'})
		index = i
	}

	return index, nil
}
