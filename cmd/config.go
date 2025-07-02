// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

// resolveRelativePath converts a relative path to absolute path relative to configDir
func resolveRelativePath(configDir, path string) string {
	if !filepath.IsAbs(path) {
		return filepath.Join(configDir, path)
	} else {
		return path
	}
}

// Config represents the configuration structure
type Config struct {
	BamPath     string
	InputMos    map[uint32]string
	NewFrames   map[uint32]string
	ExtractDir  string
	UpdateDir   string
}

// LoadConfig loads configuration from a TOML file
func LoadConfig(filePath string) (*Config, error) {
	// Create a new Viper instance for this configuration
	v := viper.New()
	v.SetConfigFile(filePath)

	// Read the configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Get the directory of the TOML file for relative path resolution
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path of config file: %v", err)
	}
	configDir := filepath.Dir(absFilePath)

	config := &Config{
		InputMos:  make(map[uint32]string),
		NewFrames: make(map[uint32]string),
	}

	// Read the BAM file path from the Input section and convert to absolute path
	bamPath := v.GetString("Input.bam")
	if bamPath == "" {
		return nil, fmt.Errorf("BAM path not found in configuration file")
	}

	config.BamPath = resolveRelativePath(configDir, bamPath)

	// Read output directories from the Output section
	extractDir := v.GetString("Output.extract")
	if extractDir != "" {
		config.ExtractDir = resolveRelativePath(configDir, extractDir)
	}

	updateDir := v.GetString("Output.update")
	if updateDir != "" {
		config.UpdateDir = resolveRelativePath(configDir, updateDir)
	}

	// Read the InputMos section into a map with uint keys and convert paths
	inputMosSection := v.GetStringMapString("InputMos")
	for key, value := range inputMosSection {
		if uintKey, err := strconv.ParseUint(key, 10, 32); err == nil {
			config.InputMos[uint32(uintKey)] = resolveRelativePath(configDir, value)
		} else {
			fmt.Printf("Warning: Invalid numeric key in InputMos: %s\n", key)
		}
	}

	// Read the NewFrames section into a map with uint keys and convert paths
	newFramesSection := v.GetStringMapString("NewFrames")
	for key, value := range newFramesSection {
		if uintKey, err := strconv.ParseUint(key, 10, 32); err == nil {
			config.NewFrames[uint32(uintKey)] = resolveRelativePath(configDir, value)
		} else {
			fmt.Printf("Warning: Invalid numeric key in NewFrames: %s\n", key)
		}
	}

	return config, nil
}

// PrintConfig prints the configuration details to console
func (c *Config) PrintConfig() {
	fmt.Printf("BAM file path: %s\n", c.BamPath)

	fmt.Printf("InputMos entries: %d\n", len(c.InputMos))
	for key, value := range c.InputMos {
		fmt.Printf("  %d: %s\n", key, value)
	}

	fmt.Printf("NewFrames entries: %d\n", len(c.NewFrames))
	for key, value := range c.NewFrames {
		fmt.Printf("  %d: %s\n", key, value)
	}

	if c.ExtractDir != "" {
		fmt.Printf("Extract directory: %s\n", c.ExtractDir)
	}

	if c.UpdateDir != "" {
		fmt.Printf("Update directory: %s\n", c.UpdateDir)
	}
}
