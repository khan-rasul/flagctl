package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	launchRemoveKey   string
	launchRemoveIndex int
)

var launchRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a launch ramp by index",
	RunE: func(cmd *cobra.Command, args []string) error {
		if launchRemoveKey == "" {
			return fmt.Errorf("--key is required")
		}

		if launchRemoveIndex <= 0 {
			return fmt.Errorf("--index must be >= 1")
		}

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

		flag, ok := cfg.GetFlag(launchRemoveKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", launchRemoveKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen", launchRemoveKey)
		}

		if err := cfg.RemoveTargetRule(flag, launchRemoveIndex); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully removed launch ramp [%d] from '%s'\n", launchRemoveIndex, launchRemoveKey)
		return nil
	},
}

func init() {
	launchRemoveCmd.Flags().StringVarP(&launchRemoveKey, "key", "k", "", "Flag key (Required)")
	launchRemoveCmd.Flags().IntVarP(&launchRemoveIndex, "index", "i", 0, "Launch ramp index (Required)")
}
