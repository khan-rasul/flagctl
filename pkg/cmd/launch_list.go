package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var launchListKey string

var launchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active progressive launches for a flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		if launchListKey == "" {
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

		flag, ok := cfg.GetFlag(launchListKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", launchListKey)
		}

		rules := cfg.GetTargetRules(flag)
		fmt.Printf("ACTIVE LAUNCH RAMPS FOR '%s' (Default: %s):\n", launchListKey, flag.DefaultVariant)

		found := false
		for _, r := range rules {
			if r.Type == config.RuleTypeLaunch {
				found = true
				if r.Attribute != "" {
					fmt.Printf("  [%d] Cohort Launch (%s == %v): fractional rollout (%s)\n", r.Index, r.Attribute, r.Value, r.Variant)
				} else {
					fmt.Printf("  [%d] Global Launch: fractional rollout (%s)\n", r.Index, r.Variant)
				}
			}
		}

		if !found {
			fmt.Println("  (No active progressive launches. Flag serves defaultVariant 100%)")
		}

		return nil
	},
}

func init() {
	launchListCmd.Flags().StringVarP(&launchListKey, "key", "k", "", "Flag key (Required)")
}
