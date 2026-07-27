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
	initFormat    string
	initOutput    string
	initFlagSetID string
	initLanguage  string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a flagd workspace and configuration file (idempotent)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if initFlagSetID == "" {
			initFlagSetID = filepath.Base(cwd)
		}

		initFormat = strings.ToLower(initFormat)
		if initFormat != "json" && initFormat != "yaml" && initFormat != "yml" {
			return fmt.Errorf("invalid format '%s'. Must be 'json' or 'yaml'", initFormat)
		}

		if initOutput == "" {
			if initFormat == "yaml" || initFormat == "yml" {
				initOutput = "flags.yaml"
			} else {
				initOutput = "flags.json"
			}
		}

		targetPath := filepath.Join(cwd, initOutput)
		isReinit := false

		if _, err := os.Stat(targetPath); err == nil {
			isReinit = true
		}

		if !isReinit {
			cfg := config.NewFlagConfig(initFlagSetID)
			if err := config.SaveConfig(targetPath, initFormat, cfg); err != nil {
				return fmt.Errorf("failed to write %s: %w", initOutput, err)
			}
		}

		// Save .flagctl.json workspace config
		wsCfg := &workspace.WorkspaceConfig{
			Version:       "1",
			ConfigPath:    initOutput,
			Format:        initFormat,
			SchemaURL:     "https://flagd.dev/schema/v0/flags.json",
			SchemaVersion: "v0",
		}

		if initLanguage != "" {
			outExt := "ts"
			if initLanguage == "go" {
				outExt = "go"
			}
			wsCfg.Codegen = &workspace.CodegenConfig{
				Enabled:    true,
				Language:   initLanguage,
				OutputPath: fmt.Sprintf("src/flags.gen.%s", outExt),
			}
		}

		if err := workspace.SaveWorkspace(cwd, wsCfg); err != nil {
			return fmt.Errorf("failed to save workspace config: %w", err)
		}

		if isReinit {
			fmt.Printf("✔ Reinitialized existing flagd configuration at %s\n", targetPath)
		} else {
			fmt.Printf("✔ Initialized empty flagd configuration at %s\n", targetPath)
			fmt.Println("💡 Run 'flagctl create' to add your first feature flag!")
		}

		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initFormat, "format", "f", "json", "Output format (json or yaml)")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "", "Custom output path for flag config")
	initCmd.Flags().StringVar(&initFlagSetID, "flag-set-id", "", "ID of the flag set (defaults to directory name)")
	initCmd.Flags().StringVarP(&initLanguage, "language", "l", "", "Language for code generation (typescript or go)")
}
