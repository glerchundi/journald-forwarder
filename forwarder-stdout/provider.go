package main

import (
	"os"
	"github.com/glerchundi/journald-forwarder/core"
)

type StdoutProviderConfig struct {
}

func NewStdoutProviderConfig() *StdoutProviderConfig {
	return &StdoutProviderConfig{}
}

func (*StdoutProviderConfig) Name() string {
	return "stdout"
}

func (*StdoutProviderConfig) BulkSize() int {
	return 1
}

type StdoutProvider struct {
	marshaller core.JournalEntryMarshaller
}

func NewStdoutProvider(config *StdoutProviderConfig) (*StdoutProvider, error) {
	return &StdoutProvider{core.JournalEntryMarshaller{}}, nil
}

func (sp *StdoutProvider) Publish(iterator core.JournalEntryIterator) (int, error) {
	index := 0
	for iterator.Next() {
		i, e := iterator.Value()
		os.Stdout.WriteString(string(sp.marshaller.MarshalOne(e)))
		os.Stdout.Write([]byte{'\n','\n'})
		index = i
	}

	return index, nil
}
