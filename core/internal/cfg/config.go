package cfg

import (
	"fmt"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var K = koanf.New(".")

func Init() error {
	if err := K.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading config.yml: %w", err)
	}
	_ = K.Load(file.Provider(".env"), dotenv.Parser())
	if err := K.Load(env.Provider("", ".", func(s string) string { return s }), nil); err != nil {
		return fmt.Errorf("error loading env vars: %w", err)
	}
	return nil
}

func Load(path string, target any) error {
	if err := K.Unmarshal(path, target); err != nil {
		return fmt.Errorf("error unmarshalling '%s': %w", path, err)
	}
	return nil
}
