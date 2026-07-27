package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/generator"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	generateLanguage string
	generateOutput   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate strongly-typed flag accessor code for your project",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		lang := generateLanguage
		if lang == "" && wsCfg.Codegen != nil {
			lang = wsCfg.Codegen.Language
		}
		if lang == "" {
			lang = "typescript"
		}

		outPath := generateOutput
		if outPath == "" && wsCfg.Codegen != nil {
			outPath = wsCfg.Codegen.OutputPath
		}
		if outPath == "" {
			if lang == "go" {
				outPath = "pkg/flags/flags.gen.go"
			} else {
				outPath = "src/flags.gen.ts"
			}
		}

		cfgPath := filepath.Join(root, wsCfg.ConfigPath)
		cfg, err := config.LoadConfig(cfgPath, wsCfg.Format)
		if err != nil {
			return err
		}

		gen, err := generator.GetGenerator(lang)
		if err != nil {
			return err
		}

		code, err := gen.Generate(cfg)
		if err != nil {
			return fmt.Errorf("failed to generate code for %s: %w", lang, err)
		}

		fullOutPath := filepath.Join(root, outPath)
		if err := os.MkdirAll(filepath.Dir(fullOutPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(fullOutPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write generated code to %s: %w", fullOutPath, err)
		}

		fmt.Printf("✔ Successfully generated %s strongly-typed accessors at %s\n", lang, outPath)

		// Save codegen preferences to .flagctl.json
		wsCfg.Codegen = &workspace.CodegenConfig{
			Enabled:    true,
			Language:   lang,
			OutputPath: outPath,
		}
		_ = workspace.SaveWorkspace(root, wsCfg)

		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateLanguage, "language", "l", "", "Language generator (typescript or go)")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path for generated accessors")
}
