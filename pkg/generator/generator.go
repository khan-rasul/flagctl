package generator

import (
	"fmt"

	"github.com/khan-rasul/flagctl/pkg/config"
)

type Generator interface {
	Language() string
	Generate(cfg *config.FlagConfig) (string, error)
}

var registry = make(map[string]Generator)

func Register(g Generator) {
	registry[g.Language()] = g
}

func GetGenerator(lang string) (Generator, error) {
	g, ok := registry[lang]
	if !ok {
		return nil, fmt.Errorf("unsupported generator language '%s'. Supported: typescript, go", lang)
	}
	return g, nil
}

func init() {
	Register(&TypeScriptGenerator{})
	Register(&GoGenerator{})
}
