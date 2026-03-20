// Package cfg provides configuration loading for the core Pulumi program.
// It merges config.yml, an optional .env file, and environment variables
// (in that order, later sources taking precedence) into a single koanf instance.
package cfg

import (
	"fmt"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// K is the global koanf instance. Call [Init] before reading any values.
var K = koanf.New(".")

// Init loads configuration from config.yml, an optional .env file, and the
// current process environment. Environment variables override file values.
// It must be called once at program start before any cfg.K lookups.
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

// Load unmarshals the koanf subtree at path into target using koanf struct tags.
// target must be a pointer to a struct with koanf field tags.
func Load(path string, target any) error {
	if err := K.Unmarshal(path, target); err != nil {
		return fmt.Errorf("error unmarshalling '%s': %w", path, err)
	}
	return nil
}
