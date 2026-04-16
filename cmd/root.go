package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "irw",
	Short: "Spaced repetition file tracker",
	Long:  `A spaced repetition system for managing reading and writing queues.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func Init(path string) error {
	return nil
}
