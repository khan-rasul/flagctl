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
	targetAddKey       string
	targetAddAttribute string
	targetAddOperator  string
	targetAddValue     string
	targetAddVariant   string
	targetAddTop       bool
)

var targetAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a targeting rule into the ordered rule chain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetAddKey == "" || targetAddAttribute == "" || targetAddValue == "" || targetAddVariant == "" {
			return fmt.Errorf("--key, --attribute, --value, and --variant are required")
		}

		if targetAddOperator == "" {
			targetAddOperator = "=="
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

		flag, ok := cfg.GetFlag(targetAddKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", targetAddKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen", targetAddKey)
		}

		if _, ok := flag.Variants[targetAddVariant]; !ok {
			return fmt.Errorf("variant '%s' does not exist", targetAddVariant)
		}

		if err := cfg.AddTargetRule(flag, targetAddAttribute, targetAddOperator, targetAddValue, targetAddVariant, targetAddTop); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Added targeting rule for '%s' (%s %s %s -> %s)\n", targetAddKey, targetAddAttribute, targetAddOperator, targetAddValue, targetAddVariant)
		return nil
	},
}

func init() {
	targetAddCmd.Flags().StringVarP(&targetAddKey, "key", "k", "", "Flag key (Required)")
	targetAddCmd.Flags().StringVarP(&targetAddAttribute, "attribute", "a", "", "Context attribute name (Required)")
	targetAddCmd.Flags().StringVarP(&targetAddOperator, "operator", "o", "==", "Operator (==, !=, endsWith, startsWith, sem_ver, in)")
	targetAddCmd.Flags().StringVarP(&targetAddValue, "value", "v", "", "Attribute match value (Required)")
	targetAddCmd.Flags().StringVar(&targetAddVariant, "variant", "", "Variant to serve (Required)")
	targetAddCmd.Flags().BoolVar(&targetAddTop, "top", false, "Insert at top of rule evaluation chain (Denylist priority)")
}
