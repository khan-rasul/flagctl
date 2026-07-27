package cmd

import (
	"fmt"

	"github.com/khan-rasul/flagctl/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of flagctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("flagctl version %s\n", version.Version)
	},
}
