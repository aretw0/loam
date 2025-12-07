package main

import (
	"fmt"
	"os"
)

func main() {
	Execute()
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
