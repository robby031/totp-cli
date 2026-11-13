package main

import (
	"fmt"
	"os"

	"github.com/robby031/totp-cli/internal/command"
	"github.com/robby031/totp-cli/internal/storage"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("totp-cli", version)
			return
		}
	}

	store, err := storage.NewDefaultStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	runner := command.NewRunner(store, os.Stdin, os.Stdout, os.Stderr)
	store.SetPasswordProvider(runner.Password)

	if err := runner.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
