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
	deprecateKey    string
	deprecateReason string
)

var deprecateCmd = &cobra.Command{
	Use:   "deprecate",
	Short: "Soft-deprecate a feature flag (keeps state ENABLED, freezes flag, tags metadata)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if deprecateKey == "" {
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

		flag, ok := cfg.GetFlag(deprecateKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", deprecateKey)
		}

		if cfg.IsDeprecated(flag) {
			fmt.Printf("ℹ Flag '%s' is already deprecated and frozen.\n", deprecateKey)
			return nil
		}

		if err := cfg.DeprecateFlag(deprecateKey, deprecateReason); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Marked flag '%s' as DEPRECATED (state remains ENABLED for safety; flag is now frozen)\n", deprecateKey)
		fmt.Println("💡 Developers should now remove code calls to this flag before running 'flagctl delete'.")

		return nil
	},
}

func init() {
	deprecateCmd.Flags().StringVarP(&deprecateKey, "key", "k", "", "Flag key (Required)")
	deprecateCmd.Flags().StringVarP(&deprecateReason, "reason", "r", "", "Reason for deprecation")
}
