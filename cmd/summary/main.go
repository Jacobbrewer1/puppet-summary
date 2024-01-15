package main

import (
	"context"
	"flag"
	"os"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/google/subcommands"
)

// Set at linking time
var (
	Commit string
	Date   string
)

func main() {
	_, err := logging.CommonLogger(logging.NewConfig(appName))
	if err != nil {
		panic(err)
	}

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(new(versionCmd), "")
	subcommands.Register(new(serveCmd), "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
