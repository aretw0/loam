package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aretw0/lifecycle"
)

func main() {
	if err := lifecycle.Run(lifecycle.Job(func(ctx context.Context) error {
		return Execute(ctx)
	})); err != nil {
		fatal("application error", err)
	}
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
