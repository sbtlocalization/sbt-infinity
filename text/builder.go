// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/dialog"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

type TextEntry struct {
	Id             int
	HasText        bool
	HasSound       bool
	HasToken       bool
	Text           string
	Sound          string
	VolumeVariance uint32
	PitchVariance  uint32
	Labels         map[string]struct{}
	Context        map[ContextType]map[string][]string
}

type ContextType int

const (
	ContextDialog ContextType = iota
	ContextSound
	ContextUI
	ContextCreature
	ContextCreatureSound
	ContextWorldMap
)

type TextCollection struct {
	Entries map[int]*TextEntry
}

func NewTextCollection(tlk *p.Tlk) *TextCollection {
	collection := &TextCollection{
		Entries: make(map[int]*TextEntry),
	}

	for id, entry := range tlk.Entries {
		text, err := entry.Text()
		if err != nil {
			fmt.Printf("Warning: unable to decode text for ID %d: %v\n", id, err)
		}

		tEntry := &TextEntry{
			Id:             id,
			HasText:        entry.Flags.TextExists,
			HasSound:       entry.Flags.SoundExists,
			HasToken:       entry.Flags.TokenExists,
			Text:           text,
			Sound:          entry.AudioName,
			VolumeVariance: entry.VolumeVariance,
			PitchVariance:  entry.PitchVariance,
			Labels:         make(map[string]struct{}),
			Context:        make(map[ContextType]map[string][]string),
		}
		collection.Entries[id] = tEntry

		if entry.Flags.SoundExists {
			collection.AddLabel(id, "with sound")
			collection.AddContext(id, ContextSound, entry.AudioName, "")
		}

		if entry.Flags.TokenExists {
			collection.AddLabel(id, "with token")
		}

		if !entry.Flags.TextExists && text == "" {
			collection.AddLabel(id, "no text")
		}

		collection.Entries[id] = tEntry
	}

	return collection
}

func (c *TextCollection) AddContext(id int, contextType ContextType, key, value string) {
	if id == 0 || id == 0xFFFFFFFF {
		return
	}

	if entry, ok := c.Entries[id]; ok {
		if entry.Context[contextType] == nil {
			entry.Context[contextType] = make(map[string][]string)
		}
		entry.Context[contextType][key] = append(entry.Context[contextType][key], value)
	}
}

func (c *TextCollection) AddCreatureSoundContext(id int, soundType string, file string) {
	c.AddContext(id, ContextCreatureSound, soundType, file)
}

func (c *TextCollection) AddLabel(id int, label string) {
	if id == 0 || id == 0xFFFFFFFF {
		return
	}

	if entry, ok := c.Entries[id]; ok {
		entry.Labels[label] = struct{}{}
	}
}

func (c *TextCollection) LoadContextFromDialogs(baseUrl string, dlg *dialog.DialogCollection) {
	for _, d := range dlg.Dialogs {
		for _, node := range d.All() {
			switch node.Type {
			case dialog.StateNodeType:
				ref := int(node.State.TextRef)
				c.AddContext(ref, ContextDialog, d.Id.String(), node.ToUrl(baseUrl))

				c.AddLabel(ref, "dialog")
				c.AddLabel(ref, "question")
				c.AddLabel(ref, d.Id.DlgName)
				c.AddLabel(ref, fmt.Sprintf("dialog %d @ %s", d.Id.Index, d.Id.DlgName))
			case dialog.TransitionNodeType:
				url := node.ToUrl(baseUrl)

				if node.Transition.HasText {
					ref := int(node.Transition.TextRef)
					c.AddContext(ref, ContextDialog, d.Id.String(), url)
					c.AddLabel(ref, "dialog")
					c.AddLabel(ref, "answer")
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf("dialog %d @ %s", d.Id.Index, d.Id.DlgName))
				}

				if node.Transition.HasJournalText {
					ref := int(node.Transition.JournalTextRef)
					c.AddContext(ref, ContextDialog, d.Id.String(), url)
					c.AddLabel(ref, "dialog")
					c.AddLabel(ref, "journal")
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf("dialog %d @ %s", d.Id.Index, d.Id.DlgName))
				}
			default:
				// ignore errors and loops
			}
		}
	}
}

func (c *TextCollection) LoadContextFromUiScreens(uiFilename string, chu *p.Chu) error {
	windows, err := chu.Windows()
	if err != nil {
		return fmt.Errorf("unable to get windows from CHU: %v", err)
	}

	for _, window := range windows {
		controls, err := window.Controls()
		if err != nil {
			return fmt.Errorf("unable to get controls from window: %v", err)
		}

		for _, control := range controls {
			data, err := control.Data()
			if err != nil {
				return fmt.Errorf("unable to get control struct: %v", err)
			}

			switch data.Type {
			case p.Chu_Control_ControlStruct_StructType__Label:
				label := data.Properties.(*p.Chu_Control_ControlStruct_Label)
				if label.InitialTextRef != 0 && label.InitialTextRef != 0xFFFFFFFF {
					ref := int(label.InitialTextRef)
					c.AddLabel(ref, "UI label")
					c.AddContext(ref, ContextUI, uiFilename, fmt.Sprintf("window %d → control %d", window.WinId, int16(data.ControlId)))
				}
			default:
				continue
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromCreature(creFilename string, cre *p.Cre, ids *p.Ids) error {
	longName := cre.LongNameRef
	if longName != 0 && longName != 0xFFFFFFFF {
		ref := int(longName)
		c.AddLabel(ref, "creature")
		c.AddContext(ref, ContextCreature, "Long name", strings.ToLower(creFilename))
	}

	shortName := cre.ShortNameRef
	if shortName != 0 && shortName != 0xFFFFFFFF {
		ref := int(shortName)
		c.AddLabel(ref, "creature")
		c.AddContext(ref, ContextCreature, "Short name (tooltip)", strings.ToLower(creFilename))
	}

	sound_refs := cre.Body.Header.StrRefs

	for val, identifier := range ids.Entries {
		if val < 0 || val >= int32(len(sound_refs)) {
			continue
		}
		ref := int(sound_refs[val])
		if ref == 0 || ref == 0xFFFFFFFF {
			continue
		}
		if dialog := cre.Body.Header.Dialog; dialog != "" && dialog != "0" && dialog != "None" {
			c.AddLabel(ref, strings.ToUpper(dialog))
		}
		c.AddCreatureSoundContext(ref, identifier, strings.ToLower(creFilename))
	}

	return nil
}

func (c *TextCollection) LoadContextFromWorldMaps(wmpFilename string, wmp *p.Wmp) error {
	wmEntries, err := wmp.WorldmapEntries()
	if err != nil {
		return fmt.Errorf("unable to get world map entries: %v", err)
	}

	for _, wmEntry := range wmEntries {
		nameRef := int(wmEntry.AreaNameRef)

		if nameRef != 0 && nameRef != 0xFFFFFFFF {
			c.AddLabel(nameRef, "world map")
			c.AddLabel(nameRef, fmt.Sprintf("map %s[%d]", wmpFilename, wmEntry.MapId))
			c.AddContext(nameRef, ContextWorldMap, "World map name", fmt.Sprintf("%s → map %d", wmpFilename, wmEntry.MapId))
		}

		areas, err := wmEntry.AreaEntries()
		if err != nil {
			return fmt.Errorf("unable to get area entries for world map entry %d: %v", wmEntry.MapId, err)
		}

		for i, area := range areas {
			if captionRef := int(area.CaptionRef); captionRef != 0 && captionRef != 0xFFFFFFFF {
				c.AddLabel(captionRef, fmt.Sprintf("area %d @ %s[%d]", i, wmpFilename, wmEntry.MapId))
				c.AddContext(captionRef, ContextWorldMap, "Area caption", fmt.Sprintf("%s → map %d → area %d", wmpFilename, wmEntry.MapId, i))
			}
			if tooltipRef := int(area.TooltipRef); tooltipRef != 0 && tooltipRef != 0xFFFFFFFF {
				c.AddLabel(tooltipRef, fmt.Sprintf("area %d @ %s[%d]", i, wmpFilename, wmEntry.MapId))
				c.AddContext(tooltipRef, ContextWorldMap, "Area tooltip", fmt.Sprintf("%s → map %d → area %d", wmpFilename, wmEntry.MapId, i))
			}
		}
	}

	return nil
}
