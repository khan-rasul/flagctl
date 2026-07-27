package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/khan-rasul/flagctl/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SupportedVersions holds the list of known schema versions.
var SupportedVersions = []string{"v0"}

// Registry manages schema compilers for different schema versions.
type Registry struct {
	compilers map[string]*jsonschema.Compiler
	schemas   map[string]*jsonschema.Schema
}

// NewRegistry initializes a Registry with embedded flagd schemas.
func NewRegistry() (*Registry, error) {
	r := &Registry{
		compilers: make(map[string]*jsonschema.Compiler),
		schemas:   make(map[string]*jsonschema.Schema),
	}

	for _, ver := range SupportedVersions {
		compiler := jsonschema.NewCompiler()
		compiler.Draft = jsonschema.Draft7

		// Custom loader for embedded filesystem
		compiler.LoadURL = func(url string) (io.ReadCloser, error) {
			cleanPath := strings.TrimPrefix(url, "https://flagd.dev/schema/")
			cleanPath = strings.TrimPrefix(cleanPath, "file:///")
			if !strings.HasPrefix(cleanPath, ver) {
				cleanPath = ver + "/" + cleanPath
			}
			f, err := schema.FS.Open(cleanPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open embedded schema '%s': %w", cleanPath, err)
			}
			return f, nil
		}

		flagsPath := fmt.Sprintf("%s/flags.json", ver)
		flagsData, err := schema.FS.ReadFile(flagsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded schema %s: %w", flagsPath, err)
		}

		schemaURL := fmt.Sprintf("https://flagd.dev/schema/%s/flags.json", ver)
		if err := compiler.AddResource(schemaURL, strings.NewReader(string(flagsData))); err != nil {
			return nil, fmt.Errorf("failed to add schema resource for %s: %w", ver, err)
		}

		sch, err := compiler.Compile(schemaURL)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema for %s: %w", ver, err)
		}

		r.compilers[ver] = compiler
		r.schemas[ver] = sch
	}

	return r, nil
}

// DetectVersion extracts schema version from $schema or $id URL or defaults to v0.
func DetectVersion(content string) string {
	if strings.Contains(content, "schema/v1") {
		return "v1"
	}
	return "v0"
}

// ValidateJSON validates raw JSON bytes against the specified schema version.
func (r *Registry) ValidateJSON(version string, jsonData []byte) error {
	sch, ok := r.schemas[version]
	if !ok {
		return fmt.Errorf("unsupported schema version '%s'. Supported versions: %v", version, SupportedVersions)
	}

	var v interface{}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	if err := sch.Validate(v); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}
