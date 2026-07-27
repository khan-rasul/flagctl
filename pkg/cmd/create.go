package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/generator"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	createKey         string
	createType        string
	createDefault     string
	createVariants    string
	createDescription string
	createDisabled    bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new feature flag definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createKey == "" {
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

		createType = strings.ToLower(createType)
		variantsMap := make(map[string]interface{})

		if createVariants != "" {
			// Parse key=val,key2=val2
			pairs := strings.Split(createVariants, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])
					if createType == "boolean" {
						variantsMap[k] = (strings.ToLower(v) == "true")
					} else if createType == "number" {
						var num float64
						fmt.Sscanf(v, "%f", &num)
						variantsMap[k] = num
					} else if createType == "object" {
						var obj interface{}
						if err := json.Unmarshal([]byte(v), &obj); err == nil {
							variantsMap[k] = obj
						} else {
							variantsMap[k] = v
						}
					} else {
						variantsMap[k] = v
					}
				}
			}
		} else {
			// Default variants based on type
			if createType == "boolean" {
				variantsMap["on"] = true
				variantsMap["off"] = false
			} else {
				variantsMap["default"] = "value"
			}
		}

		if createDefault == "" {
			if _, ok := variantsMap["on"]; ok {
				createDefault = "on"
			} else if _, ok := variantsMap["default"]; ok {
				createDefault = "default"
			} else {
				for k := range variantsMap {
					createDefault = k
					break
				}
			}
		} else {
			if _, ok := variantsMap[createDefault]; !ok {
				return fmt.Errorf("default variant '%s' must be one of the declared variants", createDefault)
			}
		}

		state := "ENABLED"
		if createDisabled {
			state = "DISABLED"
		}

		meta := make(map[string]interface{})
		if createDescription != "" {
			meta["description"] = createDescription
		}

		flag := &config.Flag{
			State:          state,
			DefaultVariant: createDefault,
			Variants:       variantsMap,
			Metadata:       meta,
		}

		if err := cfg.AddFlag(createKey, flag); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully created flag '%s' in %s\n", createKey, wsCfg.ConfigPath)

		// Trigger codegen if configured
		if wsCfg.Codegen != nil && wsCfg.Codegen.Enabled {
			if gen, err := generator.GetGenerator(wsCfg.Codegen.Language); err == nil {
				if code, err := gen.Generate(cfg); err == nil {
					outPath := filepath.Join(root, wsCfg.Codegen.OutputPath)
					_ = os.MkdirAll(filepath.Dir(outPath), 0755)
					_ = os.WriteFile(outPath, []byte(code), 0644)
					fmt.Printf("✔ Regenerated typed accessors at %s\n", wsCfg.Codegen.OutputPath)
				}
			}
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createKey, "key", "k", "", "Flag key (Required)")
	createCmd.Flags().StringVarP(&createType, "type", "t", "boolean", "Flag type (boolean, string, number, object)")
	createCmd.Flags().StringVarP(&createDefault, "default", "d", "", "Default variant name")
	createCmd.Flags().StringVarP(&createVariants, "variants", "v", "", "Variants key=val pairs (comma separated)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Human readable description")
	createCmd.Flags().BoolVar(&createDisabled, "disabled", false, "Create flag in DISABLED state")
}
