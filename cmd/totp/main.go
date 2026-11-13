package main

import (
	"fmt"
	"os"

	"totp-cli/internal/command"
	"totp-cli/internal/storage"
)

func main() {
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
