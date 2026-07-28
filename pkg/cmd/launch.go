package cmd

import (
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:     "launch",
	Aliases: []string{"rollout"},
	Short:   "Manage progressive feature launches and percentage ramps",
	Long:    `flagctl launch provides subcommands (add, list, ramp, remove) to control global and cohort-specific progressive feature launches.`,
}

func init() {
	launchCmd.AddCommand(launchAddCmd)
	launchCmd.AddCommand(launchListCmd)
	launchCmd.AddCommand(launchRampCmd)
	launchCmd.AddCommand(launchRemoveCmd)
}
