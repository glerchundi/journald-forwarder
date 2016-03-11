package core

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	flag "github.com/spf13/pflag"
)

const (
	cliName = "journald-forwarder"
)

type MainConfig struct {
	ProviderConfig ProviderConfig
	Provider       func(ProviderConfig)(Provider,error)
	Flags          func(ProviderConfig, *flag.FlagSet)
}

func Main(mainConfig MainConfig) {
	// Create forwarder config
	fc := NewForwarderConfig(mainConfig.ProviderConfig.BulkSize())

	// Define flag sets
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&fc.Path, "journal-path", fc.Path, "journald path.")
	fs.StringVar(&fc.CursorPath, "cursor-path", fc.CursorPath, "cursor path.")
	fs.DurationVar(&fc.CursorFlush, "cursor-flush", fc.CursorFlush, "cursor flush frequency.")
	fs.DurationVar(&fc.ForwardFlush, "forward-flush", fc.ForwardFlush, "forward flush frequency.")

	// If provider has custom flags, append them
	if mainConfig.Flags != nil {
		mainConfig.Flags(mainConfig.ProviderConfig, fs)
	}

	// Set normalization func
	fs.SetNormalizeFunc(
		func(f *flag.FlagSet, name string) flag.NormalizedName {
			if strings.Contains(name, "_") {
				return flag.NormalizedName(strings.Replace(name, "_", "-", -1))
			}
			return flag.NormalizedName(name)
		},
	)

	// Parse
	fs.Parse(os.Args[1:])

	// Set from env (if present)
	fs.VisitAll(func(f *flag.Flag) {
		if !f.Changed {
			key := strings.ToUpper(strings.Join(
				[]string{
					strings.Replace(cliName, "-", "_", -1),
					strings.Replace(f.Name, "-", "_", -1),
				},
				"_",
			))
			val := os.Getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}
	})

	// Create forwarder
	f, err := NewForwarder(fc)
	if err != nil {
		log.Fatalf("error creating forwarder: %v", err)
	}

	// Create provider
	p, err := mainConfig.Provider(mainConfig.ProviderConfig)
	if err != nil {
		log.Fatalf("error creating provider: %v", err)
	}

	// Run forwarder
	f.Run(p)

	// Wait for signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-f.errc:
			os.Stderr.Write([]byte(err.Error()))
		case s := <-signalChan:
			log.Print(fmt.Sprintf("Captured %v. Exiting...", s))
			close(f.stopc)
		case <-f.donec:
			os.Exit(0)
		}
	}
}