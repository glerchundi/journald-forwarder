package main

import (
	flag "github.com/spf13/pflag"
	"github.com/glerchundi/journald-forwarder/core"
)

func main() {
	// main delegate
	core.Main(core.MainConfig{
		ProviderConfig: NewLogglyProviderConfig(),
		Flags: func(pc core.ProviderConfig, fs *flag.FlagSet) {
			lc := pc.(*LogglyProviderConfig)
			fs.StringVar(&lc.Token, "loggly-token", lc.Token, "loggly token")
			fs.StringVar(&lc.Tags, "loggly-tags", lc.Tags, "loggly tags")
		},
		Provider: func(pc core.ProviderConfig) (core.Provider, error) {
			return NewLogglyProvider(pc.(*LogglyProviderConfig))
		},
	})
}