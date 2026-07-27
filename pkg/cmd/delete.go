package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/generator"
	"github.com/khan-rasul/flagctl/pkg/scanner"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	deleteKey   string
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm"},
	Short:   "Delete a feature flag definition (code-aware guard)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteKey == "" {
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

		if _, ok := cfg.GetFlag(deleteKey); !ok {
			return fmt.Errorf("flag '%s' not found in configuration", deleteKey)
		}

		// Code-Aware Guard: Scan codebase for references
		if !deleteForce {
			codeScanner := scanner.NewScanner(nil)
			matches, err := codeScanner.FindFlagReferences(root, deleteKey)
			if err == nil && len(matches) > 0 {
				fmt.Printf("❌ Cannot delete flag '%s'. It is still actively called in %d source location(s):\n\n", deleteKey, len(matches))
				for _, m := range matches {
					rel, _ := filepath.Rel(root, m.FilePath)
					fmt.Printf("   • %s:L%d -> %s\n", rel, m.LineNumber, m.Snippet)
				}
				fmt.Println("\n💡 Remove code references before deleting, or use --force to override.")
				return fmt.Errorf("deletion blocked by code-aware guard (%d active references)", len(matches))
			}
		}

		if err := cfg.DeleteFlag(deleteKey); err != nil {
			return err
		}

		if err := config.SaveConfig(cfgPath, wsCfg.Format, cfg); err != nil {
			return err
		}

		fmt.Printf("✔ Successfully deleted flag '%s' from %s\n", deleteKey, wsCfg.ConfigPath)

		// Regenerate code accessors if configured
		if wsCfg.Codegen != nil && wsCfg.Codegen.Enabled {
			if gen, err := generator.GetGenerator(wsCfg.Codegen.Language); err == nil {
				if code, err := gen.Generate(cfg); err == nil {
					outPath := filepath.Join(root, wsCfg.Codegen.OutputPath)
					_ = os.WriteFile(outPath, []byte(code), 0644)
					fmt.Printf("✔ Regenerated typed accessors at %s\n", wsCfg.Codegen.OutputPath)
				}
			}
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteKey, "key", "k", "", "Flag key (Required)")
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Force deletion without checking active code references")
}
