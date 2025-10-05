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

// Config represents the full configuration structure with game paths and per-game settings
type Config struct {
	Games       map[string]string     `toml:"Games"`
	GameConfigs map[string]GameConfig `toml:"-"`
}

// GameConfig represents per-game configuration options
type GameConfig struct {
	DialogSiteBaseUrl string `toml:"dialog_site_base_url"`
}

// LoadKeyConfig loads the configuration file from the specified path
// If the file doesn't exist, it returns a config with an empty Games map
func LoadKeyConfig(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return empty config
		return &Config{
			Games:       make(map[string]string),
			GameConfigs: make(map[string]GameConfig),
		}, nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// First, unmarshal into a raw map to get all game-specific sections
	var rawData map[string]interface{}
	if err := toml.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	// Parse main structure
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Initialize Games map if it's nil
	if config.Games == nil {
		config.Games = make(map[string]string)
	}

	// Initialize GameConfigs map
	config.GameConfigs = make(map[string]GameConfig)

	// Load per-game configurations
	for gameName := range config.Games {
		if gameSection, exists := rawData[gameName]; exists {
			// Convert the game section to TOML bytes and unmarshal into GameConfig
			gameBytes, err := toml.Marshal(gameSection)
			if err != nil {
				continue // Skip this game config if marshaling fails
			}

			var gameConfig GameConfig
			if err := toml.Unmarshal(gameBytes, &gameConfig); err != nil {
				continue // Skip this game config if unmarshaling fails
			}

			config.GameConfigs[gameName] = gameConfig
		}
	}

	return &config, nil
}

// GetGameKeyPath returns the path to the key file for a specific game and whether it exists
// Returns (path, true) if the game is found, ("", false) if not found
func (c *Config) GetGameKeyPath(gameName string) (string, bool) {
	path, exists := c.Games[gameName]
	return path, exists
}

// GetGameConfig returns the configuration for a specific game and whether it exists
// Returns (config, true) if the game has configuration, (empty config, false) if not found
func (c *Config) GetGameConfig(gameName string) (GameConfig, bool) {
	config, exists := c.GameConfigs[gameName]
	return config, exists
}

// ListGames returns a slice of all configured game names
func (c *Config) ListGames() []string {
	games := make([]string, 0, len(c.Games))
	for gameName := range c.Games {
		games = append(games, gameName)
	}
	return games
}

// GetFirstGame returns the key path of the first game from the configured games and whether any game exists
// Returns (path, true) if games are configured, ("", false) if no games are configured
// Games are returned in alphabetical order for consistency
func (c *Config) GetFirstGame() (string, bool) {
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

// ResolveDialogBaseUrl resolves the dialog base URL using the following priority:
// 1. If dlgBaseUrl is provided directly, use it
// 2. If gameName is provided, use that specific game's config
// 3. If no gameName but config has games, use the first game's config
// Returns (resolvedUrl, error)
func ResolveDialogBaseUrl(cmd *cobra.Command) (string, error) {
	configPath, _ := cmd.Flags().GetString("config")
	gameName, _ := cmd.Flags().GetString("game")
	dlgBaseUrl, _ := cmd.Flags().GetString("dlg-base-url")

	// If dialog base URL is provided directly, use it
	if dlgBaseUrl != "" {
		return dlgBaseUrl, nil
	}

	// Load config file
	config, err := LoadKeyConfig(configPath)
	if err != nil {
		// Config file exists but has errors - return the error
		return "", fmt.Errorf("error loading config: %v", err)
	}

	// If specific game requested
	if gameName != "" {
		if gameConfig, exists := config.GetGameConfig(gameName); exists {
			if gameConfig.DialogSiteBaseUrl != "" {
				return gameConfig.DialogSiteBaseUrl, nil
			}
			return "", fmt.Errorf("dialog base URL not configured for game '%s'", gameName)
		}
		return "", fmt.Errorf("game '%s' not found in config", gameName)
	}

	// Try to get first available game's dialog base URL from config
	if len(config.Games) > 0 {
		games := config.ListGames()
		sort.Strings(games)
		firstName := games[0]

		if gameConfig, exists := config.GetGameConfig(firstName); exists {
			if gameConfig.DialogSiteBaseUrl != "" {
				return gameConfig.DialogSiteBaseUrl, nil
			}
		}
		return "", fmt.Errorf("dialog base URL not configured for game '%s'", firstName)
	}

	// No games configured
	return "", fmt.Errorf("no games configured and no dialog base URL provided")
}
