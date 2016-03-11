package main

import (
	"log"

	"github.com/glerchundi/journald-forwarder/core"
)

func main() {
	p, err := NewStdoutProvider(StdoutProviderConfig{})
	if err != nil {
		log.Fatalf("error creating provider: %v", err)
	}

	f, err := core.NewForwarder(core.ForwarderConfig{
		Provider: p,
	})
	if err != nil {
		log.Fatalf("error creating forwarder: %v", err)
	}

	f.Run()
}

