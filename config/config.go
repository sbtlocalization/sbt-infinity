// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package config

import (
	"fmt"
	"os"
	"sort"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

const ConfigFileName = ".sbt-inf.toml"

// KeyConfig represents the structure of the .sbt-inf.toml configuration file
type KeyConfig struct {
	Games map[string]string `toml:"Games"`
}

// LoadKeyConfig loads the configuration file from the specified path
// If the file doesn't exist, it returns a config with an empty Games map
func LoadKeyConfig(configPath string) (*KeyConfig, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return empty config
		return &KeyConfig{Games: make(map[string]string)}, nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse TOML
	var config KeyConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Initialize Games map if it's nil
	if config.Games == nil {
		config.Games = make(map[string]string)
	}

	return &config, nil
}

// GetGameKeyPath returns the path to the key file for a specific game and whether it exists
// Returns (path, true) if the game is found, ("", false) if not found
func (c *KeyConfig) GetGameKeyPath(gameName string) (string, bool) {
	path, exists := c.Games[gameName]
	return path, exists
}

// ListGames returns a slice of all configured game names
func (c *KeyConfig) ListGames() []string {
	games := make([]string, 0, len(c.Games))
	for gameName := range c.Games {
		games = append(games, gameName)
	}
	return games
}

// GetFirstGame returns the key path of the first game from the configured games and whether any game exists
// Returns (path, true) if games are configured, ("", false) if no games are configured
// Games are returned in alphabetical order for consistency
func (c *KeyConfig) GetFirstGame() (string, bool) {
	if len(c.Games) == 0 {
		return "", false
	}

	games := c.ListGames()
	sort.Strings(games)
	firstName := games[0]
	return c.Games[firstName], true
}

// ResolveKeyPath resolves the key file path using the following priority:
// 1. If keyPath is provided directly, use it
// 2. If gameName is provided, use that specific game from config
// 3. If no gameName but config has games, use the first game
// Returns (resolvedPath, error)
func ResolveKeyPath(cmd *cobra.Command) (string, error) {
	configPath, _ := cmd.Flags().GetString("config")
	gameName, _ := cmd.Flags().GetString("game")
	keyPath, _ := cmd.Flags().GetString("key")

	// If key path is provided directly, use it
	if keyPath != "" {
		return keyPath, nil
	}

	// Load config file
	config, err := LoadKeyConfig(configPath)
	if err != nil {
		// Config file exists but has errors - return the error
		return "", fmt.Errorf("error loading config: %v", err)
	}

	// If specific game requested
	if gameName != "" {
		if path, exists := config.GetGameKeyPath(gameName); exists {
			return path, nil
		}
		return "", fmt.Errorf("game '%s' not found in config", gameName)
	}

	// Try to get first available game from config
	if path, exists := config.GetFirstGame(); exists {
		return path, nil
	}

	// No key path, no games configured
	return "", fmt.Errorf("no games configured and no key file path provided")
}
