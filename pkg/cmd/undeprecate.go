package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var undeprecateKey string

var undeprecateCmd = &cobra.Command{
	Use:   "undeprecate",
	Short: "Unfreeze and remove deprecation tag from a feature flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		if undeprecateKey == "" {
			return fmt.Errorf("--key is required")
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

		if err := cfg.UndeprecateFlag(undeprecateKey); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully un-deprecated '%s' (flag unfrozen)\n", undeprecateKey)

		return nil
	},
}

func init() {
	undeprecateCmd.Flags().StringVarP(&undeprecateKey, "key", "k", "", "Flag key (Required)")
}
