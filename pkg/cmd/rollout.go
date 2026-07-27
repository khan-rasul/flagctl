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
	rolloutKey         string
	rolloutSplits      string
	rolloutBucketBy    string
	rolloutAutoBalance bool
	rolloutComplete    string
)

var rolloutCmd = &cobra.Command{
	Use:   "rollout",
	Short: "Manage percentage-based progressive rollouts for a feature flag",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rolloutKey == "" {
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

		flag, ok := cfg.GetFlag(rolloutKey)
		if !ok {
			return fmt.Errorf("flag '%s' not found", rolloutKey)
		}

		// Frozen Deprecated State Check
		if cfg.IsDeprecated(flag) {
			return fmt.Errorf("flag '%s' is DEPRECATED and frozen. No rollout updates allowed. Use 'flagctl delete' to remove or 'flagctl undeprecate' to unfreeze", rolloutKey)
		}

		// Handling --complete
		if rolloutComplete != "" {
			if _, ok := flag.Variants[rolloutComplete]; !ok {
				return fmt.Errorf("variant '%s' does not exist in flag variants", rolloutComplete)
			}
			flag.DefaultVariant = rolloutComplete
			flag.Targeting = nil // Strip targeting rules on rollout completion
			if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
				return err
			}
			fmt.Printf("✔ Completed rollout for '%s': defaultVariant set to '%s' (targeting rules cleared)\n", rolloutKey, rolloutComplete)
			return nil
		}

		if rolloutSplits == "" {
			return fmt.Errorf("--splits or --complete is required")
		}

		// Parse splits e.g. "on=20,off=80"
		splitMap := make(map[string]int)
		total := 0
		pairs := strings.Split(rolloutSplits, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v, err := strconv.Atoi(strings.TrimSpace(kv[1]))
				if err != nil {
					return fmt.Errorf("invalid percentage value '%s'", kv[1])
				}
				if _, ok := flag.Variants[k]; !ok {
					return fmt.Errorf("variant '%s' in splits does not exist in flag variants", k)
				}
				splitMap[k] = v
				total += v
			}
		}

		// Single variant auto-fill
		if len(splitMap) == 1 && len(flag.Variants) == 2 {
			for singleVar, weight := range splitMap {
				for varKey := range flag.Variants {
					if varKey != singleVar {
						splitMap[varKey] = 100 - weight
						total = 100
						break
					}
				}
			}
		}

		// Auto-balance if requested
		if total != 100 && rolloutAutoBalance {
			diff := 100 - total
			if flag.DefaultVariant != "" {
				splitMap[flag.DefaultVariant] += diff
				total = 100
			}
		}

		if total != 100 {
			return fmt.Errorf("invalid rollout splits: total sum must equal 100%% (got %d%%). Use --auto-balance or adjust splits", total)
		}

		// Construct fractional targeting rule
		var weights []interface{}
		for k, w := range splitMap {
			weights = append(weights, []interface{}{k, w})
		}

		if rolloutBucketBy == "" {
			rolloutBucketBy = "userId"
		}

		fractionalRule := map[string]interface{}{
			"fractional": append([]interface{}{map[string]interface{}{"var": rolloutBucketBy}}, weights...),
		}

		flag.Targeting = fractionalRule

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully updated rollout for '%s' (splits: %s, bucketBy: %s)\n", rolloutKey, rolloutSplits, rolloutBucketBy)

		return nil
	},
}

func init() {
	rolloutCmd.Flags().StringVarP(&rolloutKey, "key", "k", "", "Flag key (Required)")
	rolloutCmd.Flags().StringVarP(&rolloutSplits, "splits", "s", "", "Variant splits percentage (e.g. 'on=20,off=80')")
	rolloutCmd.Flags().StringVarP(&rolloutBucketBy, "bucket-by", "b", "userId", "Bucketing attribute for pseudorandom allocation")
	rolloutCmd.Flags().BoolVar(&rolloutAutoBalance, "auto-balance", false, "Auto-balance remaining percentage to default variant")
	rolloutCmd.Flags().StringVar(&rolloutComplete, "complete", "", "Complete rollout and set default variant (e.g. --complete on)")
}
