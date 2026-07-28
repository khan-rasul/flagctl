package cmd

import (
	"github.com/spf13/cobra"
)

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage rule-based targeting and access control (Allowlists, Denylists, SemVer)",
	Long:  `flagctl target provides subcommands (add, list, remove) to configure ordered rule targeting for flags.`,
}

func init() {
	targetCmd.AddCommand(targetAddCmd)
	targetCmd.AddCommand(targetListCmd)
	targetCmd.AddCommand(targetRemoveCmd)
}
