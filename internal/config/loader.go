// CRC: crc-ConfigLoader.md, Spec: main.md
package config

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const ConfigFileName = "p2p-webapp.toml"

// LoadFromDir loads configuration from a filesystem directory
// Returns default config if file doesn't exist
// CRC: crc-ConfigLoader.md
func LoadFromDir(baseDir string) (*Config, error) {
	configPath := filepath.Join(baseDir, "config", ConfigFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, return defaults
		return DefaultConfig(), nil
	}

	// Load and parse TOML file
	cfg := DefaultConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// LoadFromZIP loads configuration from a ZIP archive
// Returns default config if file doesn't exist in archive
// CRC: crc-ConfigLoader.md
func LoadFromZIP(reader *zip.Reader) (*Config, error) {
	// Look for config file in config/ subdirectory
	configPathInZip := "config/" + ConfigFileName
	var configFile *zip.File
	for _, f := range reader.File {
		if f.Name == configPathInZip {
			configFile = f
			break
		}
	}

	// No config file in ZIP, return defaults
	if configFile == nil {
		return DefaultConfig(), nil
	}

	// Open and read the config file
	rc, err := configFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open config file from ZIP: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from ZIP: %w", err)
	}

	// Parse TOML
	cfg := DefaultConfig()
	if err := toml.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Merge merges command-line flags into configuration
// Flags take precedence over config file values
// CRC: crc-ConfigLoader.md
func (c *Config) Merge(port int, noOpen bool, linger bool, verbosity int) {
	// Only override if flag was explicitly set
	if port != 0 {
		c.Server.Port = port
	}

	// noOpen flag overrides autoOpenBrowser
	if noOpen {
		c.Behavior.AutoOpenBrowser = false
	}

	// linger flag overrides config
	if linger {
		c.Behavior.Linger = true
	}

	// verbosity flag overrides config
	if verbosity > 0 {
		c.Behavior.Verbosity = verbosity
	}
}

// Validate checks if configuration values are valid
// CRC: crc-ConfigLoader.md
func (c *Config) Validate() error {
	// Validate port
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 0-65535)", c.Server.Port)
	}

	// Validate port range
	if c.Server.PortRange < 1 {
		return fmt.Errorf("invalid port range: %d (must be >= 1)", c.Server.PortRange)
	}

	// Validate timeouts (should be positive)
	if c.Server.Timeouts.Read.Duration < 0 {
		return fmt.Errorf("invalid read timeout: %v (must be positive)", c.Server.Timeouts.Read)
	}
	if c.Server.Timeouts.Write.Duration < 0 {
		return fmt.Errorf("invalid write timeout: %v (must be positive)", c.Server.Timeouts.Write)
	}

	// Validate index file
	if c.Files.IndexFile == "" {
		return fmt.Errorf("index file cannot be empty")
	}

	return nil
}
