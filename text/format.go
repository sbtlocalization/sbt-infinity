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
		textCell.Value = entry.Text

		soundCell := row.AddCell()
		soundCell.Value = entry.Sound

		labelsCell := row.AddCell()
		labelsCell.Value = strings.Join(entry.Labels, ",")

		contextCell := row.AddCell()
		contextCell.Value = strings.Join(entry.Context[ContextDialog], "\n\n")

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
