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
	targetKey       string
	targetAttribute string
	targetOperator  string
	targetValue     string
	targetVariant   string
)

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Add attribute-based targeting rules to a feature flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetKey == "" || targetAttribute == "" || targetVariant == "" {
			return fmt.Errorf("--key, --attribute, and --variant are required")
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

		flag, ok := cfg.GetFlag(targetKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", targetKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen. No targeting updates allowed", targetKey)
		}

		if _, ok := flag.Variants[targetVariant]; !ok {
			return fmt.Errorf("variant '%s' does not exist in flag variants", targetVariant)
		}

		if targetOperator == "" {
			targetOperator = "endsWith"
		}

		// Construct JsonLogic if statement
		// { "if": [ { "endsWith": [ { "var": attribute }, value ] }, variant, defaultVariant ] }
		rule := map[string]interface{}{
			"if": []interface{}{
				map[string]interface{}{
					targetOperator: []interface{}{
						map[string]interface{}{"var": targetAttribute},
						targetValue,
					},
				},
				targetVariant,
				flag.DefaultVariant,
			},
		}

		flag.Targeting = rule

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully added targeting rule for '%s' (if %s %s %s -> %s)\n", targetKey, targetAttribute, targetOperator, targetValue, targetVariant)

		return nil
	},
}

func init() {
	targetCmd.Flags().StringVarP(&targetKey, "key", "k", "", "Flag key (Required)")
	targetCmd.Flags().StringVarP(&targetAttribute, "attribute", "a", "", "Context attribute name (e.g. 'email', 'userId')")
	targetCmd.Flags().StringVarP(&targetOperator, "operator", "o", "endsWith", "Comparison operator (endsWith, startsWith, ==, in)")
	targetCmd.Flags().StringVarP(&targetValue, "value", "v", "", "Targeting value")
	targetCmd.Flags().StringVar(&targetVariant, "variant", "", "Variant to serve if rule matches")
}
