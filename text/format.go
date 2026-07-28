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
	"github.com/samber/lo"
)

const (
	autoListInlineLimit = 3
	labelFamilyLimit    = 10
)

func (c *TextCollection) ExportToXlsx(outputPath string, timestamps map[uint32]int64) error {
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Sheet1")
	if err != nil {
		return fmt.Errorf("failed to add sheet: %w", err)
	}

	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "key"
	headerRow.AddCell().Value = "source or translation"
	headerRow.AddCell().Value = "labels"
	headerRow.AddCell().Value = "context"
	headerRow.AddCell().Value = "has text"
	headerRow.AddCell().Value = "has token"
	headerRow.AddCell().Value = "has sound"
	headerRow.AddCell().Value = "sound file"
	headerRow.AddCell().Value = "volume variance"
	headerRow.AddCell().Value = "pitch variance"
	headerRow.AddCell().Value = "timestamp"

	ids := slices.Sorted(maps.Keys(c.Entries))

	for _, id := range ids {
		entry := c.Entries[id]

		row := sheet.AddRow()

		idCell := row.AddCell()
		idCell.SetInt(int(id))

		textCell := row.AddCell()
		textCell.SetString(entry.Text)

		labelsCell := row.AddCell()
		labelsCell.SetString(strings.Join(renderLabels(entry.Labels), ","))

		contextCell := row.AddCell()
		contextCell.SetString(joinContext(entry, c.Long))

		hasTextCell := row.AddCell()
		hasTextCell.SetBool(entry.HasText)

		hasTokenCell := row.AddCell()
		hasTokenCell.SetBool(entry.HasToken)

		hasSoundCell := row.AddCell()
		hasSoundCell.SetBool(entry.HasSound)

		soundCell := row.AddCell()
		soundCell.SetString(entry.Sound)

		volumeVariance := row.AddCell()
		volumeVariance.SetInt64(int64(entry.VolumeVariance))

		pitchVariance := row.AddCell()
		pitchVariance.SetInt64(int64(entry.PitchVariance))

		timestampCell := row.AddCell()
		if timestamps != nil {
			if ts, ok := timestamps[id]; ok {
				timestampCell.SetInt64(ts)
			}
		}
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

func renderLabels(labels map[string]LabelFamily) []string {
	familyCounts := make(map[LabelFamily]int)
	for _, family := range labels {
		if family != "" {
			familyCounts[family]++
		}
	}

	rendered := make(map[string]struct{}, len(labels))
	for label, family := range labels {
		if familyCounts[family] <= labelFamilyLimit {
			rendered[label] = struct{}{}
			continue
		}
		marker := "*." + string(family)
		if family == labelFamilySpeaker {
			marker = "many speakers"
		}
		rendered[marker] = struct{}{}
	}
	return slices.Sorted(maps.Keys(rendered))
}

func toListItem(item string, _ int) string {
	return fmt.Sprintf("- %s", item)
}

func toList(key string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	lines := lo.UniqMap(values, toListItem)
	return fmt.Sprintf("%s:\n%s", key, strings.Join(lines, "\n"))
}

func uniqueSorted(values []string) []string {
	unique := lo.Uniq(values)
	slices.Sort(unique)
	return unique
}

func toAutoListLong(key string, values []string) string {
	unique := uniqueSorted(values)
	if len(unique) == 0 {
		return ""
	}
	var text string
	if len(unique) <= autoListInlineLimit {
		text = strings.Join(unique, ", ")
	} else {
		text = "\n" + strings.Join(lo.Map(unique, toListItem), "\n")
	}
	return fmt.Sprintf("%s: %s", key, text)
}

func toAutoListShort(key string, values []string) string {
	unique := uniqueSorted(values)
	if len(unique) == 0 {
		return ""
	}
	text := strings.Join(unique, ", ")
	if len(unique) > autoListInlineLimit {
		text = strings.Join(unique[:autoListInlineLimit], ", ") + fmt.Sprintf(" (+%d more)", len(unique)-autoListInlineLimit)
	}
	return fmt.Sprintf("%s: %s", key, text)
}

func dialogActionName(call string) string {
	name, _, ok := strings.Cut(call, "(")
	if !ok {
		return ""
	}
	return strings.TrimSpace(name)
}

func joinDialogScriptContext(contexts map[string][]string, autoList func(string, []string) string) string {
	var journalEntries []string
	var otherActions []string

	for call, values := range contexts {
		if dialogActionName(call) == "AddJournalEntry" {
			journalEntries = append(journalEntries, toList(call, values))
			continue
		}
		otherActions = append(otherActions, autoList(call, values))
	}

	slices.Sort(journalEntries)
	slices.Sort(otherActions)
	return strings.Join(append(journalEntries, otherActions...), "\n")
}

func joinContext(entry *TextEntry, long bool) string {
	contexts := entry.Context
	var parts []string

	autoList := toAutoListShort
	if long {
		autoList = toAutoListLong
	}

	if sndContexts, ok := contexts[ContextTlkSound]; ok && len(sndContexts) > 0 {
		files := lo.MapToSlice(sndContexts, func(file string, _ []string) string { return file })
		parts = append(parts, "SOUND: "+strings.Join(files, "\n"))
	}

	if dlgContexts, ok := contexts[ContextDialog]; ok && len(dlgContexts) > 0 {
		dialogs := lo.MapToSlice(dlgContexts, toList)
		slices.Sort(dialogs)
		parts = append(parts, "DIALOGS: ----------\n"+strings.Join(dialogs, "\n"))
	}

	if scriptContexts, ok := contexts[ContextScript]; ok && len(scriptContexts) > 0 {
		compiledScripts := lo.MapToSlice(scriptContexts, toList)
		slices.Sort(compiledScripts)
		parts = append(parts, "SCRIPTS: ----------\n"+strings.Join(compiledScripts, "\n"))
	}

	if dlgScriptContexts, ok := contexts[ContextDialogScript]; ok && len(dlgScriptContexts) > 0 {
		parts = append(parts, "DIALOG SCRIPTS: ----------\n"+joinDialogScriptContext(dlgScriptContexts, autoList))
	}

	if uiContexts, ok := contexts[ContextUI]; ok && len(uiContexts) > 0 {
		screens := lo.MapToSlice(uiContexts, toList)
		slices.Sort(screens)
		parts = append(parts, "UI: ----------\n"+strings.Join(screens, "\n"))
	}

	if creContexts, ok := contexts[ContextCreature]; ok && len(creContexts) > 0 {
		creatures := lo.MapToSlice(creContexts, autoList)
		slices.Sort(creatures)
		parts = append(parts, "CREATURES: ----------\n"+strings.Join(creatures, "\n"))
	}

	if soundContexts, ok := contexts[ContextCreatureSound]; ok && len(soundContexts) > 0 {
		groups := lo.MapToSlice(soundContexts, func(soundType string, files []string) string {
			return fmt.Sprintf("- %s ← %s", soundType, strings.Join(files, ", "))
		})
		slices.Sort(groups)
		parts = append(parts, "USED BY CREATURE: ----------\n"+strings.Join(groups, "\n"))
	}

	if wmContexts, ok := contexts[ContextWorldMap]; ok && len(wmContexts) > 0 {
		maps := lo.MapToSlice(wmContexts, autoList)
		slices.Sort(maps)
		parts = append(parts, "WORLD MAPS: ----------\n"+strings.Join(maps, "\n"))
	}

	if areContexts, ok := contexts[ContextArea]; ok && len(areContexts) > 0 {
		areas := lo.MapToSlice(areContexts, autoList)
		slices.Sort(areas)
		parts = append(parts, "AREAS: ----------\n"+strings.Join(areas, "\n"))
	}

	if itemContexts, ok := contexts[ContextItem]; ok && len(itemContexts) > 0 {
		items := lo.MapToSlice(itemContexts, autoList)
		slices.Sort(items)
		parts = append(parts, "ITEMS: ----------\n"+strings.Join(items, "\n"))
	}

	if projContexts, ok := contexts[ContextProjectile]; ok && len(projContexts) > 0 {
		projectiles := lo.MapToSlice(projContexts, autoList)
		slices.Sort(projectiles)
		parts = append(parts, "PROJECTILES: ----------\n"+strings.Join(projectiles, "\n"))
	}

	if spellContexts, ok := contexts[ContextSpell]; ok && len(spellContexts) > 0 {
		spells := lo.MapToSlice(spellContexts, autoList)
		slices.Sort(spells)
		parts = append(parts, "SPELLS: ----------\n"+strings.Join(spells, "\n"))
	}

	if storeContexts, ok := contexts[ContextStore]; ok && len(storeContexts) > 0 {
		stores := lo.MapToSlice(storeContexts, autoList)
		slices.Sort(stores)
		parts = append(parts, "STORES: ----------\n"+strings.Join(stores, "\n"))
	}

	if effectContexts, ok := contexts[ContextEffect]; ok && len(effectContexts) > 0 {
		effects := lo.MapToSlice(effectContexts, autoList)
		slices.Sort(effects)
		parts = append(parts, "EFFECTS: ----------\n"+strings.Join(effects, "\n"))
	}

	if tracking2DAContexts, ok := contexts[ContextTracking2DA]; ok && len(tracking2DAContexts) > 0 {
		trackings := lo.MapToSlice(tracking2DAContexts, autoList)
		slices.Sort(trackings)
		parts = append(parts, "TRACKING.2DA: ----------\n"+strings.Join(trackings, "\n"))
	}

	if subtitleContexts, ok := contexts[ContextSubtitles]; ok && len(subtitleContexts) > 0 {
		sounds := lo.MapToSlice(subtitleContexts, toList)
		slices.Sort(sounds)
		parts = append(parts, "SUBTITLES: ----------\n"+strings.Join(sounds, "\n"))
	}

	return strings.Join(parts, "\n\n")
}
