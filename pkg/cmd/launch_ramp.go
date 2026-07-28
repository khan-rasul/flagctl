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
	launchRampKey     string
	launchRampPercent int
	launchRampVariant string
	launchRampIndex   int
)

var launchRampCmd = &cobra.Command{
	Use:   "ramp",
	Short: "Ramp launch percentage up or down (0% to 100%)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if launchRampKey == "" {
			return fmt.Errorf("--key is required")
		}

		if launchRampPercent < 0 || launchRampPercent > 100 {
			return fmt.Errorf("percentage must be between 0 and 100")
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

		flag, ok := cfg.GetFlag(launchRampKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", launchRampKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen", launchRampKey)
		}

		if launchRampPercent == 100 && launchRampVariant != "" {
			flag.DefaultVariant = launchRampVariant
			flag.Targeting = nil
			if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
				return err
			}
			fmt.Printf("✔ Completed launch for '%s': defaultVariant updated to '%s' (100%%)\n", launchRampKey, launchRampVariant)
			return nil
		}

		splits := make(map[string]int)
		targetVar := launchRampVariant
		if targetVar == "" {
			for k := range flag.Variants {
				if k != flag.DefaultVariant {
					targetVar = k
					break
				}
			}
		}

		splits[targetVar] = launchRampPercent
		splits[flag.DefaultVariant] = 100 - launchRampPercent

		if err := cfg.AddLaunchRamp(flag, "", nil, splits, "userId"); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Ramped launch percentage for '%s' to %d%% (%s=%d%%, %s=%d%%)\n", launchRampKey, launchRampPercent, targetVar, launchRampPercent, flag.DefaultVariant, 100-launchRampPercent)
		return nil
	},
}

func init() {
	launchRampCmd.Flags().StringVarP(&launchRampKey, "key", "k", "", "Flag key (Required)")
	launchRampCmd.Flags().IntVarP(&launchRampPercent, "percent", "p", 0, "Percentage ramp (0 to 100)")
	launchRampCmd.Flags().StringVarP(&launchRampVariant, "variant", "v", "", "Target launch variant")
	launchRampCmd.Flags().IntVarP(&launchRampIndex, "index", "i", 0, "Launch ramp index")
}
