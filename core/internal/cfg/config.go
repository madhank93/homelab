// Package cfg provides configuration loading for the core Pulumi program.
// It merges an optional .env file and environment variables (later sources take
// precedence) into a single koanf instance. Non-secret values live in the shared
// config package (config/config.yml).
package cfg

import (
	"fmt"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// K is the global koanf instance. Call [Init] before reading any values.
var K = koanf.New(".")

// Init loads an optional .env file and the current process environment.
// Environment variables override .env values.
// It must be called once at program start before any cfg.K lookups.
func Init() error {
	_ = K.Load(file.Provider(".env"), dotenv.Parser())
	if err := K.Load(env.Provider("", ".", func(s string) string { return s }), nil); err != nil {
		return fmt.Errorf("error loading env vars: %w", err)
	}
	return nil
}
