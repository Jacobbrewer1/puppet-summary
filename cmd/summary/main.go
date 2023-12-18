package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
)

// Set at linking time
var (
	Commit string
	Date   string
)

var versionFlag = flag.Bool("version", false, "Print version information and exit")

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf(
			"Commit: %s\nRuntime: %s %s/%s\nDate: %s\n",
			Commit,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
			Date,
		)
		os.Exit(0)
	}

	a, err := initializeApp()
	if err != nil {
		log.Fatalln(err)
	}
	if err := generateConfig(); err != nil {
		slog.Error("Error generating config", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}
	slog.Debug("Starting application")
	if err := a.run(); err != nil {
		slog.Error("Error running application", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}
}
