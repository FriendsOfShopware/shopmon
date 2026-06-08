package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "shopmon",
		Short: "Shopmon - Shopware monitoring tool",
	}

	root.AddCommand(serverCmd())
	root.AddCommand(workerCmd())
	root.AddCommand(migrateCmd())
	root.AddCommand(fixturesCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
