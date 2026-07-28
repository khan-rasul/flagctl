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
	targetRemoveKey   string
	targetRemoveIndex int
)

var targetRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a targeting rule by index",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetRemoveKey == "" {
			return fmt.Errorf("--key is required")
		}

		if targetRemoveIndex <= 0 {
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

		flag, ok := cfg.GetFlag(targetRemoveKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", targetRemoveKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen", targetRemoveKey)
		}

		if err := cfg.RemoveTargetRule(flag, targetRemoveIndex); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully removed targeting rule [%d] from '%s'\n", targetRemoveIndex, targetRemoveKey)
		return nil
	},
}

func init() {
	targetRemoveCmd.Flags().StringVarP(&targetRemoveKey, "key", "k", "", "Flag key (Required)")
	targetRemoveCmd.Flags().IntVarP(&targetRemoveIndex, "index", "i", 0, "Rule index (Required)")
}
