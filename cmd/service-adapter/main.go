package main

import (
	"log"
	"os"

	"github.com/gurbuzali/hazelcast-service-adapter/adapter"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

func main() {
	stderrLogger := log.New(os.Stderr, "[hazelcast-service-adapter] ", log.LstdFlags)
	manifestGenerator := &adapter.ManifestGenerator{
		StderrLogger: stderrLogger,
	}
	binder := &adapter.Binder{
		CommandRunner:       adapter.ExternalCommandRunner{},
		StderrLogger:        stderrLogger,
	}
	serviceadapter.HandleCommandLineInvocation(os.Args, manifestGenerator, binder, &adapter.DashboardUrlGenerator{})
}
