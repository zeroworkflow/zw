package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zw",
	Short: "Zero Workflow - AI-powered developer tools",
	Long: `Zero Workflow is a collection of AI-powered developer tools
that help automate common development tasks like generating commits,
documentation, and more.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
