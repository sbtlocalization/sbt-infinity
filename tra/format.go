// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package tra

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/parser"
)

func WrapWithDelimiters(text string) string {
	switch {
	case !strings.Contains(text, "~"):
		return "~" + text + "~"
	case !strings.Contains(text, "%"):
		return "%" + text + "%"
	case !strings.Contains(text, "\""):
		return "\"" + text + "\""
	default:
		return "~~~~~" + text + "~~~~~"
	}
}

func FormatEntry(entry parser.TraEntry, maxIdWidth int) string {
	var textPart string
	if entry.FemaleText != "" {
		textPart = WrapWithDelimiters(entry.MaleText) + " " + WrapWithDelimiters(entry.FemaleText)
	} else {
		textPart = WrapWithDelimiters(entry.MaleText)
	}

	result := fmt.Sprintf("@%-*d = %s", maxIdWidth, entry.ID, textPart)

	if entry.SoundFile != "" {
		result += fmt.Sprintf(" [%s]", entry.SoundFile)
	}

	return result
}

func Write(w io.Writer, entries []parser.TraEntry) error {
	maxIdWidth := calculateMaxIdWidth(entries)

	for _, entry := range entries {
		line := FormatEntry(entry, maxIdWidth)
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			return fmt.Errorf("error writing entry %d: %w", entry.ID, err)
		}
	}

	return nil
}

func WriteFile(path string, entries []parser.TraEntry) error {
	outputDir := filepath.Dir(path)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("unable to create output directory: %w", err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}
	defer file.Close()

	return Write(file, entries)
}

func calculateMaxIdWidth(entries []parser.TraEntry) int {
	maxId := uint32(0)
	for _, entry := range entries {
		if entry.ID > maxId {
			maxId = entry.ID
		}
	}
	return len(fmt.Sprintf("%d", maxId))
}
