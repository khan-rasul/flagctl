package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	launchAddKey       string
	launchAddPercent   int
	launchAddVariant   string
	launchAddSplits    string
	launchAddBucketBy  string
	launchAddAttribute string
	launchAddValue     string
)

var launchAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a progressive launch ramp (global or cohort-specific)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if launchAddKey == "" {
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

		flag, ok := cfg.GetFlag(launchAddKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", launchAddKey)
		}

		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen. No launches allowed", launchAddKey)
		}

		splits := make(map[string]int)

		if launchAddSplits != "" {
			total := 0
			for _, pair := range strings.Split(launchAddSplits, ",") {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v, err := strconv.Atoi(strings.TrimSpace(kv[1]))
					if err != nil {
						return fmt.Errorf("invalid percentage value '%s'", kv[1])
					}
					if _, ok := flag.Variants[k]; !ok {
						return fmt.Errorf("variant '%s' in splits does not exist", k)
					}
					splits[k] = v
					total += v
				}
			}
			if total != 100 {
				return fmt.Errorf("splits total must equal 100%% (got %d%%)", total)
			}
		} else {
			if launchAddVariant == "" {
				// Pick non-default variant or first variant
				for k := range flag.Variants {
					if k != flag.DefaultVariant {
						launchAddVariant = k
						break
					}
				}
				if launchAddVariant == "" {
					launchAddVariant = flag.DefaultVariant
				}
			}

			if _, ok := flag.Variants[launchAddVariant]; !ok {
				return fmt.Errorf("variant '%s' does not exist in flag variants", launchAddVariant)
			}

			splits[launchAddVariant] = launchAddPercent
			splits[flag.DefaultVariant] = 100 - launchAddPercent
		}

		if launchAddBucketBy == "" {
			launchAddBucketBy = "userId"
		}

		if err := cfg.AddLaunchRamp(flag, launchAddAttribute, launchAddValue, splits, launchAddBucketBy); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		if launchAddAttribute != "" {
			fmt.Printf("✔ Added cohort launch ramp for '%s' (Cohort: %s == %s, Variant: %s @ %d%%)\n", launchAddKey, launchAddAttribute, launchAddValue, launchAddVariant, launchAddPercent)
		} else {
			fmt.Printf("✔ Added global launch ramp for '%s' (Variant: %s @ %d%%)\n", launchAddKey, launchAddVariant, launchAddPercent)
		}

		return nil
	},
}

func init() {
	launchAddCmd.Flags().StringVarP(&launchAddKey, "key", "k", "", "Flag key (Required)")
	launchAddCmd.Flags().IntVarP(&launchAddPercent, "percent", "p", 0, "Progressive launch percentage (0 to 100)")
	launchAddCmd.Flags().StringVarP(&launchAddVariant, "variant", "v", "", "Launch variant")
	launchAddCmd.Flags().StringVarP(&launchAddSplits, "splits", "s", "", "Multi-variant splits (e.g. 'v1=30,v2=70')")
	launchAddCmd.Flags().StringVarP(&launchAddBucketBy, "bucket-by", "b", "userId", "Bucketing attribute ID")
	launchAddCmd.Flags().StringVarP(&launchAddAttribute, "attribute", "a", "", "Cohort attribute for targeted launch")
	launchAddCmd.Flags().StringVar(&launchAddValue, "value", "", "Cohort attribute value")
}
