// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"

	"github.com/sbtlocalization/sbt-infinity/dialog"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

type TextEntry struct {
	Id       int
	HasText  bool
	HasSound bool
	HasToken bool
	Text     string
	Sound    string
	Labels   []string
	Context  map[ContextType][]string
}

type ContextType int

const (
	ContextDialog ContextType = iota
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
			Id:       id,
			HasText:  entry.Flags.TextExists,
			HasSound: entry.Flags.SoundExists,
			HasToken: entry.Flags.TokenExists,
			Text:     text,
			Sound:    entry.AudioName,
			Labels:   make([]string, 0),
			Context:  make(map[ContextType][]string),
		}

		if entry.Flags.SoundExists {
			tEntry.Labels = append(tEntry.Labels, "with sound")
		}

		if entry.Flags.TokenExists {
			tEntry.Labels = append(tEntry.Labels, "with token")
		}

		if !entry.Flags.TextExists && text == "" {
			tEntry.Labels = append(tEntry.Labels, "no text")
		}

		collection.Entries[id] = tEntry
	}

	return collection
}

func (c *TextCollection) AddContext(id int, contextType ContextType, context string) {
	if id == 0xFFFFFFFF {
		// Invalid text reference, skip
		return
	}

	if entry, ok := c.Entries[id]; ok {
		if entry.Context == nil {
			entry.Context = make(map[ContextType][]string)
		}
		ctx := entry.Context[contextType]
		if ctx == nil {
			ctx = make([]string, 0)
		}
		entry.Context[contextType] = append(ctx, context)
	}
}

func (c *TextCollection) AddLabel(id int, label string) {
	if id == 0xFFFFFFFF {
		// Invalid text reference, skip
		return
	}

	if entry, ok := c.Entries[id]; ok {
		entry.Labels = append(entry.Labels, label)
	}
}

func (c *TextCollection) LoadContextFromDialogs(baseUrl string, dlg *dialog.DialogCollection) {
	for _, d := range dlg.Dialogs {
		for _, node := range d.All() {
			switch node.Type {
			case dialog.StateNodeType:
				ref := int(node.State.TextRef)
				c.AddContext(ref, ContextDialog, node.ToUrl(baseUrl))

				c.AddLabel(ref, "dialog")
				c.AddLabel(ref, "NPC's line")
				c.AddLabel(ref, d.Id.DlgName)
				c.AddLabel(ref, fmt.Sprintf("dialog %s", d.Id))
			case dialog.TransitionNodeType:
				url := node.ToUrl(baseUrl)

				if node.Transition.HasText {
					ref := int(node.Transition.TextRef)
					c.AddContext(ref, ContextDialog, url)
					c.AddLabel(ref, "dialog")
					c.AddLabel(ref, "player's line")
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf("dialog %s", d.Id))
				}

				if node.Transition.HasJournalText {
					ref := int(node.Transition.JournalTextRef)
					c.AddContext(ref, ContextDialog, url)
					c.AddLabel(ref, "dialog")
					c.AddLabel(ref, "journal")
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf("dialog %s", d.Id))
				}
			default:
				// ignore errors and loops
			}
		}
	}
}
