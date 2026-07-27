package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all feature flags and their current states",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := workspace.FindRoot(cwd)
		if err != nil {
			return err
		}

		wsCfg, err := workspace.LoadWorkspace(root)
		if err != nil {
			return err
		}

		cfgPath := filepath.Join(root, wsCfg.ConfigPath)
		cfg, err := config.LoadConfig(cfgPath, wsCfg.Format)
		if err != nil {
			return err
		}

		if len(cfg.Flags) == 0 {
			fmt.Printf("No flags found in %s. Use 'flagctl create' to add a flag.\n", wsCfg.ConfigPath)
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "KEY\tSTATE\tDEFAULT\tVARIANTS\tSTATUS")
		fmt.Fprintln(w, "---\t-----\t-------\t--------\t------")

		for key, flag := range cfg.Flags {
			varNames := ""
			for vk := range flag.Variants {
				if varNames != "" {
					varNames += ","
				}
				varNames += vk
			}

			status := "Active"
			if cfg.IsDeprecated(flag) {
				status = "⚠️ DEPRECATED (Frozen)"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", key, flag.State, flag.DefaultVariant, varNames, status)
		}

		w.Flush()
		return nil
	},
}
