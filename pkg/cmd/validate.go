package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/schema"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var validatePath string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate flag configuration file against the versioned flagd schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		targetFile := validatePath
		schemaVersion := "v0"

		if targetFile == "" {
			root, err := workspace.FindRoot(cwd)
			if err != nil {
				return err
			}
			wsCfg, err := workspace.LoadWorkspace(root)
			if err != nil {
				return err
			}
			targetFile = filepath.Join(root, wsCfg.ConfigPath)
			schemaVersion = wsCfg.SchemaVersion
		}

		cfg, err := config.LoadConfig(targetFile, "json")
		if err != nil {
			return err
		}

		// Convert to JSON bytes for schema validation
		jsonData, err := json.Marshal(cfg)
		if err != nil {
			return err
		}

		reg, err := schema.NewRegistry()
		if err != nil {
			return fmt.Errorf("failed to load schema registry: %w", err)
		}

		if err := reg.ValidateJSON(schemaVersion, jsonData); err != nil {
			fmt.Printf("❌ Validation failed for %s against schema %s:\n%v\n", targetFile, schemaVersion, err)
			return err
		}

		fmt.Printf("✔ %s is valid according to flagd schema %s!\n", targetFile, schemaVersion)
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validatePath, "file", "f", "", "Path to flag config file")
}
