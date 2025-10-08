// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
)

func (c *TextCollection) ExportToXlsx(outputPath string) error {
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Sheet1")
	if err != nil {
		return fmt.Errorf("failed to add sheet: %w", err)
	}

	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "key"
	headerRow.AddCell().Value = "source or translation"
	headerRow.AddCell().Value = "sound file"
	headerRow.AddCell().Value = "labels"
	headerRow.AddCell().Value = "context"
	headerRow.AddCell().Value = "has text"
	headerRow.AddCell().Value = "has sound"
	headerRow.AddCell().Value = "has token"

	ids := slices.Sorted(maps.Keys(c.Entries))

	for _, id := range ids {
		entry := c.Entries[id]

		row := sheet.AddRow()

		idCell := row.AddCell()
		idCell.SetInt(id)

		textCell := row.AddCell()
		textCell.SetString(entry.Text)

		soundCell := row.AddCell()
		soundCell.SetString(entry.Sound)

		labelsCell := row.AddCell()
		labelsCell.SetString(strings.Join(slices.Sorted(maps.Keys(entry.Labels)), ","))

		contextCell := row.AddCell()
		contextCell.SetString(joinContext(entry.Context))

		hasTextCell := row.AddCell()
		hasTextCell.SetBool(entry.HasText)

		hasSoundCell := row.AddCell()
		hasSoundCell.SetBool(entry.HasSound)

		hasTokenCell := row.AddCell()
		hasTokenCell.SetBool(entry.HasToken)
	}

	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create output directory %s: %v", outputDir, err)
	}

	err = xlsxFile.Save(outputPath)
	if err != nil {
		return fmt.Errorf("failed to save xlsx file: %w", err)
	}

	return nil
}

func joinContext(contexts map[ContextType][]string) string {
	var parts []string
	if sndContexts, exists := contexts[ContextSound]; exists && len(sndContexts) > 0 {
		parts = append(parts, "Sound: "+strings.Join(sndContexts, "\n"))
	}
	if dlgContexts, exists := contexts[ContextDialog]; exists && len(dlgContexts) > 0 {
		parts = append(parts, "Dialogs:\n"+strings.Join(dlgContexts, "\n"))
	}
	if uiContexts, exists := contexts[ContextUI]; exists && len(uiContexts) > 0 {
		parts = append(parts, "UI:\n"+strings.Join(uiContexts, "\n"))
	}
	return strings.Join(parts, "\n\n")
}
