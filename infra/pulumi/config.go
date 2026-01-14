package main

import (
	"fmt"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Define k globally for package main
var k = koanf.New(".")

// InitConfig loads .env and config.yml. Call this once in main.go.
func InitConfig() error {
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		fmt.Printf("Info: .env file not loaded: %v\n", err)
	}

	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading config.yml: %w", err)
	}

	// 3. Load System Environment Variables
	// This allows using generic os env vars if needed, mapped to config structure
	// k.Load(env.Provider("", ".", func(s string) string {
	// 	return strings.Replace(strings.ToLower(s), "_", ".", -1)
	// }), nil)

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
