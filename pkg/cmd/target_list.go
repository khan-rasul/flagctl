package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var targetListKey string

var targetListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display ordered targeting rules and overlap analysis",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetListKey == "" {
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

		flag, ok := cfg.GetFlag(targetListKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", targetListKey)
		}

		rules := cfg.GetTargetRules(flag)
		fmt.Printf("ORDERED TARGETING RULES FOR '%s' (Default: %s):\n", targetListKey, flag.DefaultVariant)

		if len(rules) == 0 {
			fmt.Println("  (No targeting rules configured. Flag serves defaultVariant 100%)")
			return nil
		}

		for _, r := range rules {
			if r.Type == config.RuleTypeLaunch {
				fmt.Printf("  [%d] LAUNCH    : fractional rollout (%s)\n", r.Index, r.Variant)
			} else {
				fmt.Printf("  [%d] %-9s: %s %s %v -> %s\n", r.Index, r.Type, r.Attribute, r.Operator, r.Value, r.Variant)
			}
		}

		return nil
	},
}

func init() {
	targetListCmd.Flags().StringVarP(&targetListKey, "key", "k", "", "Flag key (Required)")
}
