package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flagctl",
	Short: "flagctl is a CLI tool for managing flagd feature flag configurations in Git",
	Long:  `flagctl provides an ergonomic command line interface to create, update, launch, target, deprecate, delete, validate, audit, and generate feature flag code accessors for flagd GitOps workflows.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(launchCmd)
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
