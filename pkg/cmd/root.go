package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flagctl",
	Short: "flagctl is a CLI tool to manage flagd feature flag configurations in Git",
	Long: `flagctl manages the lifecycle of flagd feature flags (create, update, rollout, target, deprecate, delete, validate, audit, generate)
directly in your repository's flags.json / flags.yaml configuration files adhering strictly to the flagd schema.`,
}

// Execute runs the root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(rolloutCmd)
	rootCmd.AddCommand(targetCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deprecateCmd)
	rootCmd.AddCommand(undeprecateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}
