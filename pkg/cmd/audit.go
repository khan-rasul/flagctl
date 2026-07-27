package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/khan-rasul/flagctl/pkg/scanner"
	"github.com/khan-rasul/flagctl/pkg/workspace"
	"github.com/spf13/cobra"
)

var auditStrict bool

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit codebase against flag config for missing, orphaned, and deprecated flags",
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

		cfgPath := filepath.Join(root, wsCfg.ConfigPath)
		cfg, err := config.LoadConfig(cfgPath, wsCfg.Format)
		if err != nil {
			return err
		}

		codeScanner := scanner.NewScanner(nil)
		report, err := codeScanner.Audit(root, cfg)
		if err != nil {
			return err
		}

		fmt.Printf("🔍 Auditing codebase in %s against %s...\n\n", root, wsCfg.ConfigPath)

		hasIssues := false

		// 1. Missing Flags
		if len(report.MissingFlags) > 0 {
			hasIssues = true
			fmt.Printf("❌ Missing Flags (%d called in code, missing in config):\n", len(report.MissingFlags))
			for _, m := range report.MissingFlags {
				rel, _ := filepath.Rel(root, m.FilePath)
				fmt.Printf("   • '%s' at %s:L%d\n", m.FlagKey, rel, m.LineNumber)
			}
			fmt.Println()
		} else {
			fmt.Println("✔ 0 Missing flags (All code calls exist in config)")
		}

		// 2. Deprecated Flags in Code
		if len(report.DeprecatedInCode) > 0 {
			hasIssues = true
			fmt.Printf("⚠️ Deprecated Flags Still in Code (%d calls remaining):\n", len(report.DeprecatedInCode))
			for _, m := range report.DeprecatedInCode {
				rel, _ := filepath.Rel(root, m.FilePath)
				fmt.Printf("   • '%s' at %s:L%d\n", m.FlagKey, rel, m.LineNumber)
			}
			fmt.Println()
		} else {
			fmt.Println("✔ 0 Deprecated flags called in code")
		}

		// 3. Orphaned Flags
		if len(report.OrphanedFlags) > 0 {
			fmt.Printf("ℹ️ Orphaned Flags (%d defined in config, 0 code references):\n", len(report.OrphanedFlags))
			for _, key := range report.OrphanedFlags {
				fmt.Printf("   • '%s' (Run 'flagctl delete --key %s' to clean up)\n", key, key)
			}
			fmt.Println()
		} else {
			fmt.Println("✔ 0 Orphaned flags")
		}

		if hasIssues && auditStrict {
			return fmt.Errorf("audit failed in --strict mode: missing or deprecated flag issues detected")
		}

		return nil
	},
}

func init() {
	auditCmd.Flags().BoolVar(&auditStrict, "strict", false, "Exit with status code 1 if missing or deprecated flag issues are found (for CI/CD)")
}
