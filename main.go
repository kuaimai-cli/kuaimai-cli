package main

import (
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
