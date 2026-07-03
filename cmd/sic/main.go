package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wfrs-dev/sic/internal/form"
)

const logo = `┌─────────────────────────┐
│ ⬣ Spring Initializr CLI │
└─────────────────────────┘`

func DefaultLogger() *os.File {
	f, err := os.OpenFile("/tmp/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	logger := slog.New(
		slog.NewTextHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	slog.SetDefault(logger)

	return f
}

var Version = "v.0.0.0-dev"

func main() {
	/*
		f := DefaultLogger()
		defer f.Close()
		//*/

	fmt.Println(logo)
	fmt.Println("  Version", Version)
	fmt.Println()
	form, err := form.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = form.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
