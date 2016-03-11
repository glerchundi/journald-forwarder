package main

import (
	"github.com/glerchundi/journald-forwarder/core"
)

func main() {
	// main delegate
	core.Main(core.MainConfig{
		ProviderConfig: NewStdoutProviderConfig(),
		Provider: func(pc core.ProviderConfig) (core.Provider, error) {
			return NewStdoutProvider(pc.(*StdoutProviderConfig))
		},
	})
}

