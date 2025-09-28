// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

const ConfigFileName = ".sbt-inf.toml"

// KeyConfig represents the structure of the .sbt-inf.toml configuration file
type KeyConfig struct {
	Games map[string]string `toml:"Games"`
}

// LoadKeyConfig automatically loads the .sbt-inf.toml file from the current working directory
// If the file doesn't exist, it returns a config with an empty Games map
func LoadKeyConfig() (*KeyConfig, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &KeyConfig{Games: make(map[string]string)}, nil
	}

	// Construct path to config file
	configPath := filepath.Join(cwd, ConfigFileName)

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
// 1. If gameName is provided, use that specific game from config
// 2. If no gameName but config has games, use the first game
// 3. If no config or no games, use the provided keyPath argument
// Returns (resolvedPath, error)
func ResolveKeyPath(keyPath string, gameName string) (string, error) {
	config, err := LoadKeyConfig()
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

	// No config or no games - require keyPath argument
	if keyPath == "" {
		return "", fmt.Errorf("no games configured and no key file path provided")
	}

	return keyPath, nil
}

// ParseArgsWithKeyPath parses command arguments to separate key file path from other files
// Returns (keyPath, otherFiles) where keyPath is extracted if the first argument has .key extension
func ParseArgsWithKeyPath(args []string) (string, []string) {
	if len(args) == 0 {
		return "", []string{}
	}

	// If first argument has .key extension, treat it as keyPath
	if filepath.Ext(args[0]) == ".key" {
		return args[0], args[1:]
	}

	// Otherwise, all arguments are other files
	return "", args
}

// ResolveKeyPathFromArgs combines argument parsing and key path resolution
// This is the main function that commands should use to handle key path resolution
func ResolveKeyPathFromArgs(args []string, gameName string) (string, []string, error) {
	keyPath, otherFiles := ParseArgsWithKeyPath(args)

	resolvedKeyPath, err := ResolveKeyPath(keyPath, gameName)
	if err != nil {
		return "", nil, err
	}

	return resolvedKeyPath, otherFiles, nil
}

// AddGameFlag adds the standard --game flag to a cobra command
// This is a helper function to ensure consistent flag naming across commands
func AddGameFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("game", "g", "", "Game name from config to use")
}
