package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/subcommands"
)

// Set at linking time
var (
	Commit string
	Date   string
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(new(versionCmd), "")
	subcommands.Register(new(serveCmd), "")
	subcommands.Register(new(purgeCmd), "")

	flag.Parse()

	// Listen for ctrl+c and kill signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		got := <-sig
		slog.Debug("Received signal, shutting down", slog.String("signal", got.String()))
		cancel()
	}()

	os.Exit(int(subcommands.Execute(ctx)))
}
