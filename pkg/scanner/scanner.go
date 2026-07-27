package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/khan-rasul/flagctl/pkg/config"
)

type Match struct {
	FlagKey    string `json:"flagKey"`
	FilePath   string `json:"filePath"`
	LineNumber int    `json:"lineNumber"`
	Snippet    string `json:"snippet"`
}

type AuditReport struct {
	MissingFlags     []Match  `json:"missingFlags"`
	DeprecatedInCode []Match  `json:"deprecatedInCode"`
	OrphanedFlags    []string `json:"orphanedFlags"`
}

type Scanner struct {
	Patterns      []*regexp.Regexp
	AccessorPattern *regexp.Regexp
	Includes      []string
	Excludes      []string
}

// pascalToKebab converts "NewCheckoutFlow" or "GetNewCheckoutFlow" to "new-checkout-flow"
func pascalToKebab(s string) string {
	s = strings.TrimPrefix(s, "get")
	s = strings.TrimPrefix(s, "is")
	var res []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				res = append(res, '-')
			}
			res = append(res, r+('a'-'A'))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

func DefaultSDKPatterns() []*regexp.Regexp {
	raw := []string{
		// JS/TS/Java/Python/Rust: getBooleanValue("flag-key", ...)
		`\b(?:getBooleanValue|getStringValue|getNumberValue|getObjectValue|getEvaluationDetails|get_boolean_value|get_string_value|get_number_value|get_object_value|get_bool_value|useFeatureFlag|useFlag)\s*\(\s*["']([^"']+)["']`,
		// Go: BooleanValue(ctx, "flag-key", ...)
		`\b(?:BooleanValue|StringValue|NumberValue|ObjectValue|Boolean|String|Float|Int)\s*\(\s*[^,]+,\s*["']([^"']+)["']`,
		// React JSX: <FeatureFlag name="flag-key">
		`\bname=["']([^"']+)["']`,
	}

	compiled := make([]*regexp.Regexp, 0, len(raw))
	for _, p := range raw {
		if re, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}

func NewScanner(customPatterns []string) *Scanner {
	patterns := DefaultSDKPatterns()
	for _, cp := range customPatterns {
		if re, err := regexp.Compile(cp); err == nil {
			patterns = append(patterns, re)
		}
	}

	return &Scanner{
		Patterns:        patterns,
		AccessorPattern: regexp.MustCompile(`flags\.((?:get|is)?[A-Z][a-zA-Z0-9_]+)`),
		Includes:        []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py", ".java", ".cs", ".php", ".rs"},
		Excludes:        []string{"node_modules", "vendor", "dist", ".git", "coverage"},
	}
}

func (s *Scanner) isIncluded(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, inc := range s.Includes {
		if ext == inc {
			return true
		}
	}
	return false
}

func (s *Scanner) isExcluded(path string) bool {
	parts := strings.Split(path, string(os.PathSeparator))
	for _, part := range parts {
		for _, exc := range s.Excludes {
			if part == exc {
				return true
			}
		}
	}
	return false
}

func (s *Scanner) ScanDirectory(root string) ([]Match, error) {
	var matches []Match

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if s.isExcluded(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !s.isIncluded(path) || s.isExcluded(path) {
			return nil
		}

		fileMatches, err := s.ScanFile(path)
		if err == nil {
			matches = append(matches, fileMatches...)
		}

		return nil
	})

	return matches, err
}

func (s *Scanner) ScanFile(path string) ([]Match, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []Match
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Standard SDK string patterns
		for _, re := range s.Patterns {
			submatches := re.FindAllStringSubmatch(line, -1)
			for _, sub := range submatches {
				if len(sub) > 1 && sub[1] != "" {
					matches = append(matches, Match{
						FlagKey:    sub[1],
						FilePath:   path,
						LineNumber: lineNum,
						Snippet:    strings.TrimSpace(line),
					})
				}
			}
		}

		// Typed accessor pattern (flags.getNewCheckoutFlow)
		if s.AccessorPattern != nil {
			accMatches := s.AccessorPattern.FindAllStringSubmatch(line, -1)
			for _, sub := range accMatches {
				if len(sub) > 1 && sub[1] != "" {
					kebabKey := pascalToKebab(sub[1])
					matches = append(matches, Match{
						FlagKey:    kebabKey,
						FilePath:   path,
						LineNumber: lineNum,
						Snippet:    strings.TrimSpace(line),
					})
				}
			}
		}
	}

	return matches, scanner.Err()
}

func (s *Scanner) FindFlagReferences(root string, flagKey string) ([]Match, error) {
	allMatches, err := s.ScanDirectory(root)
	if err != nil {
		return nil, err
	}

	var filtered []Match
	for _, m := range allMatches {
		if strings.EqualFold(m.FlagKey, flagKey) || pascalToKebab(m.FlagKey) == flagKey {
			filtered = append(filtered, m)
		}
	}

	return filtered, nil
}

func (s *Scanner) Audit(root string, cfg *config.FlagConfig) (*AuditReport, error) {
	matches, err := s.ScanDirectory(root)
	if err != nil {
		return nil, err
	}

	codeFlags := make(map[string][]Match)
	for _, m := range matches {
		codeFlags[m.FlagKey] = append(codeFlags[m.FlagKey], m)
	}

	report := &AuditReport{
		MissingFlags:     []Match{},
		DeprecatedInCode: []Match{},
		OrphanedFlags:    []string{},
	}

	for key, mList := range codeFlags {
		flag, exists := cfg.GetFlag(key)
		if !exists {
			report.MissingFlags = append(report.MissingFlags, mList...)
		} else if cfg.IsDeprecated(flag) {
			report.DeprecatedInCode = append(report.DeprecatedInCode, mList...)
		}
	}

	for key := range cfg.Flags {
		if _, called := codeFlags[key]; !called {
			report.OrphanedFlags = append(report.OrphanedFlags, key)
		}
	}

	return report, nil
}
