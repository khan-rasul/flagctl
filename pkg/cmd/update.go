package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	updateKey         string
	updateState       string
	updateDefault     string
	updateDescription string
	updateRenameTo    string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update core attributes of an existing feature flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateKey == "" {
			return fmt.Errorf("--key is required")
		}

		// Enforce Immutable Keys Invariant
		if updateRenameTo != "" {
			return fmt.Errorf("flag keys are IMMUTABLE and cannot be renamed. To rename a feature flag, create a new flag and deprecate the old one")
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

		flag, ok := cfg.GetFlag(updateKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", updateKey)
		}

		// Enforce Frozen Deprecated State Check
		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen. No updates allowed", updateKey)
		}

		updated := false

		if updateState != "" {
			st := strings.ToUpper(updateState)
			if st != "ENABLED" && st != "DISABLED" {
				return fmt.Errorf("invalid state '%s'. Must be ENABLED or DISABLED", updateState)
			}
			flag.State = st
			updated = true
		}

		if updateDefault != "" {
			if _, ok := flag.Variants[updateDefault]; !ok {
				return fmt.Errorf("default variant '%s' does not exist in flag variants", updateDefault)
			}
			flag.DefaultVariant = updateDefault
			updated = true
		}

		if updateDescription != "" {
			if flag.Metadata == nil {
				flag.Metadata = make(map[string]interface{})
			}
			flag.Metadata["description"] = updateDescription
			updated = true
		}

		if !updated {
			return fmt.Errorf("no update flags specified. Provide --state, --default, or --description")
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully updated flag '%s' in %s\n", updateKey, wsCfg.ConfigPath)

		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateKey, "key", "k", "", "Flag key (Required)")
	updateCmd.Flags().StringVarP(&updateState, "state", "s", "", "Change state (ENABLED or DISABLED)")
	updateCmd.Flags().StringVarP(&updateDefault, "default", "d", "", "Change default variant name")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "Update description")
	updateCmd.Flags().StringVar(&updateRenameTo, "rename-to", "", "Forbidden key rename parameter")
}
