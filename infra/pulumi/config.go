package main

import (
	"fmt"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Define k globally for package main
var k = koanf.New(".")

// InitConfig loads configuration in priority order (last load wins):
//  1. config.yml  — non-sensitive base config, checked into git
//  2. .env file   — optional local override, gitignored (legacy / dev convenience)
//  3. env vars    — highest priority; injected by: sops exec-env infra/secrets/bootstrap.env.sops -- pulumi ...
func InitConfig() error {
	// 1. Base config (non-sensitive, required)
	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading config.yml: %w", err)
	}

	// 2. Optional .env file (gitignored; for local dev without SOPS)
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		// .env is optional — not an error when absent
	}

	// 3. Process environment variables (keys preserved as-is, e.g. PROXMOX_PASSWORD, HCLOUD_TOKEN)
	// These are injected by `sops exec-env` and take precedence over all file-based config.
	if err := k.Load(env.Provider("", ".", func(s string) string { return s }), nil); err != nil {
		return fmt.Errorf("error loading env vars: %w", err)
	}

	return nil
}

// path: The YAML key (e.g. "proxmox")
// target: Pointer to the struct (e.g. &clusterConfig)
func LoadConfig(path string, target any) error {
	if err := k.Unmarshal(path, target); err != nil {
		return fmt.Errorf("error unmarshalling config section '%s': %w", path, err)
	}
	return nil
}
