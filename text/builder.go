// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/dialog"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

type TextEntry struct {
	Id             uint32
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
	ContextArea ContextType = iota
	ContextCreature
	ContextCreatureSound
	ContextDialog
	ContextEffect
	ContextItem
	ContextProjectile
	ContextSound
	ContextSpell
	ContextStore
	ContextTracking2DA
	ContextUI
	ContextWorldMap
)

const (
	lb_ability             = "ability"
	lb_area                = "area"
	lb_area_automap_note   = "automap note"
	lb_area_container      = "container"
	lb_area_door           = "door"
	lb_area_rest_encounter = "rest encounter"
	lb_area_speaker        = "speaker"
	lb_area_trigger        = "area trigger"
	lb_creature            = "creature"
	lb_cynicismQuote       = "cynicism quote"
	lb_dialog              = "dialog"
	lb_dialog_answer       = "answer"
	lb_dialog_journal      = "journal"
	lb_dialog_question     = "question"
	lb_effect              = "used in effect"
	lb_item                = "item"
	lb_projectile          = "projectile"
	lb_ranger_tracking     = "tracking.2da"
	lb_spell               = "spell"
	lb_store               = "store"
	lb_store_drink         = "store drink"
	lb_tlk_no_text         = "no text"
	lb_tlk_with_sound      = "with sound"
	lb_tlk_with_token      = "with token"
	lb_ui                  = "UI label"
	lb_world_map           = "world map"
)

type TextCollection struct {
	Entries map[uint32]*TextEntry
}

func NewTextCollection(tlk *p.Tlk) *TextCollection {
	collection := &TextCollection{
		Entries: make(map[uint32]*TextEntry),
	}

	for i, entry := range tlk.Entries {
		id := uint32(i)
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
			collection.AddLabel(id, lb_tlk_with_sound)
			collection.AddContext(id, ContextSound, entry.AudioName, "")
		}

		if entry.Flags.TokenExists {
			collection.AddLabel(id, lb_tlk_with_token)
		}

		if !entry.Flags.TextExists && text == "" {
			collection.AddLabel(id, lb_tlk_no_text)
		}

		collection.Entries[id] = tEntry
	}

	return collection
}

func (c *TextCollection) AddContext(id uint32, contextType ContextType, key, value string) {
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

func (c *TextCollection) AddCreatureSoundContext(id uint32, soundType string, file string) {
	c.AddContext(id, ContextCreatureSound, soundType, file)
}

func (c *TextCollection) AddLabel(id uint32, label string) {
	if id == 0 || id == 0xFFFFFFFF {
		return
	}

	if entry, ok := c.Entries[id]; ok {
		entry.Labels[label] = struct{}{}
	}
}

func (c *TextCollection) LoadContextFromDialogs(baseUrl string, dlg *dialog.DialogCollection) {
	const labelFormat = "dialog %d @ %s"

	for _, d := range dlg.Dialogs {
		for _, node := range d.All() {
			switch node.Type {
			case dialog.StateNodeType:
				ref := node.State.TextRef
				c.AddContext(ref, ContextDialog, d.Id.String(), node.ToUrl(baseUrl))

				c.AddLabel(ref, lb_dialog)
				c.AddLabel(ref, lb_dialog_question)
				c.AddLabel(ref, node.Origin.DlgName)
				c.AddLabel(ref, fmt.Sprintf(labelFormat, d.Id.Index, d.Id.DlgName))
			case dialog.TransitionNodeType:
				url := node.ToUrl(baseUrl)

				if node.Transition.HasText {
					ref := node.Transition.TextRef
					c.AddContext(ref, ContextDialog, d.Id.String(), url)
					c.AddLabel(ref, lb_dialog)
					c.AddLabel(ref, lb_dialog_answer)
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf(labelFormat, d.Id.Index, d.Id.DlgName))
				}

				if node.Transition.HasJournalText {
					ref := node.Transition.JournalTextRef
					c.AddContext(ref, ContextDialog, d.Id.String(), url)
					c.AddLabel(ref, lb_dialog)
					c.AddLabel(ref, lb_dialog_journal)
					c.AddLabel(ref, d.Id.DlgName)
					c.AddLabel(ref, fmt.Sprintf(labelFormat, d.Id.Index, d.Id.DlgName))
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
					ref := label.InitialTextRef
					c.AddLabel(ref, lb_ui)
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
		c.AddLabel(longName, lb_creature)
		c.AddContext(longName, ContextCreature, "Long name", strings.ToLower(creFilename))
	}

	shortName := cre.ShortNameRef
	if shortName != 0 && shortName != 0xFFFFFFFF {
		c.AddLabel(shortName, lb_creature)
		c.AddContext(shortName, ContextCreature, "Short name (tooltip)", strings.ToLower(creFilename))
	}

	soundRefs := cre.Body.Header.StrRefs

	for val, identifier := range ids.Entries {
		if val < 0 || val >= int32(len(soundRefs)) {
			continue
		}
		ref := soundRefs[val]
		if ref == 0 || ref == 0xFFFFFFFF {
			continue
		}
		if dialog := cre.Body.Header.Dialog; dialog != "" && dialog != "0" && dialog != "None" {
			c.AddLabel(ref, strings.ToUpper(dialog))
		}
		c.AddCreatureSoundContext(ref, identifier, strings.ToLower(creFilename))
	}

	if effects, err := cre.Body.Effects(); err == nil {
		const creatureEffectPattern = "Creature %s → effect %d"

		for i, effect := range effects {
			switch v := effect.(type) {
			case *p.Eff_BodyV1:
				if strref, context := getStrrefFromEffect(uint32(v.Opcode), v.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
					c.AddLabel(strref, lb_creature)
					c.AddLabel(strref, lb_effect)
					c.AddContext(strref, ContextEffect, context, fmt.Sprintf(creatureEffectPattern, creFilename, i))
				}
			case *p.Eff_BodyV2:
				// Special handling for opcode 0x14A (Show floating text)
				opcode := v.Opcode
				param1 := v.Parameter1
				if opcode == 0x14A && param1 == 0 && v.Parameter2 == 1 {
					fromStrref := v.Parameter3
					count := v.Special

					for i := fromStrref; i < fromStrref+count; i++ {
						c.AddLabel(i, lb_effect)
						c.AddLabel(i, lb_cynicismQuote)
						c.AddContext(i, ContextEffect, "Show floating text", fmt.Sprintf(creatureEffectPattern, creFilename, i))
					}
				} else if strref, context := getStrrefFromEffect(uint32(v.Opcode), v.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
					c.AddLabel(strref, lb_creature)
					c.AddLabel(strref, lb_effect)
					c.AddContext(strref, ContextEffect, context, fmt.Sprintf(creatureEffectPattern, creFilename, i))
				}
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromWorldMaps(wmpFilename string, wmp *p.Wmp) error {
	wmEntries, err := wmp.WorldmapEntries()
	if err != nil {
		return fmt.Errorf("unable to get world map entries: %v", err)
	}

	for _, wmEntry := range wmEntries {
		nameRef := wmEntry.AreaNameRef

		if nameRef != 0 && nameRef != 0xFFFFFFFF {
			c.AddLabel(nameRef, lb_world_map)
			c.AddContext(nameRef, ContextWorldMap, "World map name", fmt.Sprintf("%s → map %d", wmpFilename, wmEntry.MapId))
		}

		areas, err := wmEntry.AreaEntries()
		if err != nil {
			return fmt.Errorf("unable to get area entries for world map entry %d: %v", wmEntry.MapId, err)
		}

		for i, area := range areas {
			if captionRef := area.CaptionRef; captionRef != 0 && captionRef != 0xFFFFFFFF {
				c.AddLabel(captionRef, lb_area)
				c.AddContext(captionRef, ContextWorldMap, "Area caption", fmt.Sprintf("%s → map %d → area %d", wmpFilename, wmEntry.MapId, i))
			}
			if tooltipRef := area.TooltipRef; tooltipRef != 0 && tooltipRef != 0xFFFFFFFF {
				c.AddLabel(tooltipRef, lb_area)
				c.AddContext(tooltipRef, ContextWorldMap, "Area tooltip", fmt.Sprintf("%s → map %d → area %d", wmpFilename, wmEntry.MapId, i))
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromArea(areFilename string, are *p.Are) error {
	if regions, err := are.Regions(); err == nil {
		for i, region := range regions {
			if infoRef := region.InfoRef; infoRef != 0 && infoRef != 0xFFFFFFFF {
				c.AddLabel(infoRef, lb_area_trigger)
				c.AddContext(infoRef, ContextArea, "Info point on the map", fmt.Sprintf("%s → trigger %d", areFilename, i))
			}

			if speakerRef := region.PstSpeakerNameRef; speakerRef != 0 && speakerRef != 0xFFFFFFFF {
				c.AddLabel(speakerRef, lb_area_trigger)
				c.AddLabel(speakerRef, lb_area_speaker)
				c.AddContext(speakerRef, ContextArea, "Trigger's speaker name (PST only)", fmt.Sprintf("%s → trigger %d", areFilename, i))
			}
		}
	}

	if containers, err := are.Containers(); err == nil {
		for i, container := range containers {
			if lockpickRef := container.LockpickRef; lockpickRef != 0 && lockpickRef != 0xFFFFFFFF {
				c.AddLabel(lockpickRef, lb_area_container)
				c.AddContext(lockpickRef, ContextArea, "Container's lockpicking message", fmt.Sprintf("%s → container %d", areFilename, i))
			}
		}
	}

	if doors, err := are.Doors(); err == nil {
		for i, door := range doors {
			if unlockMessageRef := door.UnlockMessageRef; unlockMessageRef != 0 && unlockMessageRef != 0xFFFFFFFF {
				c.AddLabel(unlockMessageRef, lb_area_door)
				c.AddContext(unlockMessageRef, ContextArea, "Unlock message", fmt.Sprintf("%s → door %d", areFilename, i))
			}

			if speakerRef := door.SpeakerNameRef; speakerRef != 0 && speakerRef != 0xFFFFFFFF {
				c.AddLabel(speakerRef, lb_area_door)
				c.AddLabel(speakerRef, lb_area_speaker)
				c.AddContext(speakerRef, ContextArea, "Door's speaker name", fmt.Sprintf("%s → door %d", areFilename, i))
			}
		}
	}

	if bgAutomapNotes, err := are.BgAutomapNotes(); err == nil {
		for i, note := range bgAutomapNotes {
			// Read only internal notes
			if textRef := note.NoteRef; note.NoteRefIsInternal.Value && textRef != 0 && textRef != 0xFFFFFFFF {
				c.AddLabel(textRef, lb_area_automap_note)
				c.AddContext(textRef, ContextArea, "Automap note", fmt.Sprintf("%s → automap note %d", areFilename, i))
			}
		}
	}

	if restEncounters, err := are.RestEncounters(); err == nil {
		for _, creatureTextRef := range restEncounters.CreatureTextRef {
			if creatureTextRef != 0 && creatureTextRef != 0xFFFFFFFF {
				c.AddLabel(creatureTextRef, lb_area_rest_encounter)
				c.AddContext(creatureTextRef, ContextArea, "Rest encounter message", areFilename)
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromItem(itmFilename string, itm *p.Itm) error {
	if unidentNameRef := itm.UnidentifiedNameRef; unidentNameRef != 0 && unidentNameRef != 0xFFFFFFFF {
		c.AddLabel(unidentNameRef, lb_item)
		c.AddContext(unidentNameRef, ContextItem, "General (unidentified) item name", itmFilename)
	}

	if identNameRef := itm.IdentifiedNameRef; identNameRef != 0 && identNameRef != 0xFFFFFFFF {
		c.AddLabel(identNameRef, lb_item)
		c.AddContext(identNameRef, ContextItem, "Identified item name", itmFilename)
	}

	if unidentDescRef := itm.UnidentifiedDescriptionRef; unidentDescRef != 0 && unidentDescRef != 0xFFFFFFFF {
		c.AddLabel(unidentDescRef, lb_item)
		c.AddContext(unidentDescRef, ContextItem, "General (unidentified) item description", itmFilename)
	}

	if identDescRef := itm.IdentifiedDescriptionRef; identDescRef != 0 && identDescRef != 0xFFFFFFFF {
		c.AddLabel(identDescRef, lb_item)
		c.AddContext(identDescRef, ContextItem, "Identified item description", itmFilename)
	}

	if globalEffects, err := itm.GlobalEffects(); err == nil {
		for i, effect := range globalEffects {
			if strref, context := getStrrefFromEffect(uint32(effect.Opcode), effect.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
				c.AddLabel(strref, lb_item)
				c.AddLabel(strref, lb_effect)
				c.AddContext(strref, ContextEffect, context, fmt.Sprintf("Item %s → global effect %d", itmFilename, i))
			}
		}
	}

	if abilities, err := itm.ExtendedHeaders(); err == nil {
		for i, ability := range abilities {
			if effects, err := ability.Effects(); err == nil {
				for j, effect := range effects {
					if strref, context := getStrrefFromEffect(uint32(effect.Opcode), effect.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
						c.AddLabel(strref, lb_item)
						c.AddLabel(strref, lb_ability)
						c.AddLabel(strref, lb_effect)
						c.AddContext(strref, ContextEffect, context, fmt.Sprintf("Item %s → ability %d → effect %d", itmFilename, i, j))
					}
				}
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromProjectile(proFilename string, pro *p.Pro) error {
	if messageRef := pro.MessageRef; messageRef != 0 && messageRef != 0xFFFFFFFF {
		c.AddLabel(messageRef, lb_projectile)
		c.AddContext(messageRef, ContextProjectile, "Projectile's message", proFilename)
	}

	return nil
}

func (c *TextCollection) LoadContextFromSpell(splFilename string, spl *p.Spl) error {
	if unidentNameRef := spl.UnidentifiedNameRef; unidentNameRef != 0 && unidentNameRef != 0xFFFFFFFF {
		c.AddLabel(unidentNameRef, lb_spell)
		c.AddContext(unidentNameRef, ContextSpell, "General (unidentified) spell name", splFilename)
	}

	if identNameRef := spl.IdentifiedNameRef; identNameRef != 0 && identNameRef != 9_999_999 && identNameRef != 0xFFFFFFFF {
		c.AddLabel(identNameRef, lb_spell)
		c.AddContext(identNameRef, ContextSpell, "Identified spell name", splFilename)
	}

	if unidentDescRef := spl.UnidentifiedDescriptionRef; unidentDescRef != 0 && unidentDescRef != 0xFFFFFFFF {
		c.AddLabel(unidentDescRef, lb_spell)
		c.AddContext(unidentDescRef, ContextSpell, "General (unidentified) spell description", splFilename)
	}

	if identDescRef := spl.IdentifiedDescriptionRef; identDescRef != 0 && identDescRef != 9_999_999 && identDescRef != 0xFFFFFFFF {
		c.AddLabel(identDescRef, lb_spell)
		c.AddContext(identDescRef, ContextSpell, "Identified spell description", splFilename)
	}

	if globalEffects, err := spl.Effects(); err == nil {
		for i, effect := range globalEffects {
			if strref, context := getStrrefFromEffect(uint32(effect.Opcode), effect.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
				c.AddLabel(strref, lb_spell)
				c.AddLabel(strref, lb_effect)
				c.AddContext(strref, ContextEffect, context, fmt.Sprintf("Spell %s → effect %d", splFilename, i))
			}
		}
	}

	if abilities, err := spl.ExtendedHeaders(); err == nil {
		for i, ability := range abilities {
			if effects, err := ability.Effects(); err == nil {
				for j, effect := range effects {
					if strref, context := getStrrefFromEffect(uint32(effect.Opcode), effect.Parameter1); strref != 0 && strref != 0xFFFFFFFF {
						c.AddLabel(strref, lb_spell)
						c.AddLabel(strref, lb_ability)
						c.AddLabel(strref, lb_effect)
						c.AddContext(strref, ContextEffect, context, fmt.Sprintf("Spell %s → ability %d → effect %d", splFilename, i, j))
					}
				}
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromStore(stoFilename string, sto *p.Sto) error {
	if nameRef := sto.NameRef; nameRef != 0 && nameRef != 0xFFFFFFFF {
		c.AddLabel(nameRef, lb_store)
		c.AddContext(nameRef, ContextStore, "Store name", stoFilename)
	}

	if drinks, err := sto.Drinks(); err == nil {
		for i, drink := range drinks {
			if nameRef := drink.DrinkNameRef; nameRef != 0 && nameRef != 0xFFFFFFFF {
				c.AddLabel(nameRef, lb_store)
				c.AddLabel(nameRef, lb_store_drink)
				c.AddContext(nameRef, ContextStore, "Drink name (at merchant)", fmt.Sprintf("%s → drink %d", stoFilename, i))
			}
		}
	}

	return nil
}

func (c *TextCollection) LoadContextFromEffect(effFilename string, eff *p.Eff) error {
	opcode := eff.Body.Opcode
	param1 := eff.Body.Parameter1

	// Special handling for opcode 0x14A (Show floating text)
	if opcode == 0x14A && param1 == 0 && eff.Body.Parameter2 == 1 {
		fromStrref := eff.Body.Parameter3
		count := eff.Body.Special

		for i := fromStrref; i < fromStrref+count; i++ {
			c.AddLabel(i, lb_effect)
			c.AddLabel(i, lb_cynicismQuote)
			c.AddContext(i, ContextEffect, "Show floating text", effFilename)
		}
	} else if strref, context := getStrrefFromEffect(opcode, param1); strref != 0 && strref != 0xFFFFFFFF {
		c.AddLabel(strref, lb_effect)
		c.AddContext(strref, ContextEffect, context, effFilename)
	}

	return nil
}

func getStrrefFromEffect(opcode, param1 uint32) (uint32, string) {
	opcodeToExplanationMap := map[uint32]string{
		0x67:  "Change name to specified",                                 // Change name
		0x8B:  "Display string (when spell is applied)",                   // Display string
		0xB4:  "“Can't use item” message",                                 // Can't use item
		0xCE:  "“Protection from spell” message (when spell is absorbed)", // Spell: Protection from spell
		0xFD:  "Add map marker",                                           // Spell effect: Add map marker
		0xFE:  "Remove map marker",                                        // Spell effect: Remove map marker
		0x10B: "Prevent string from being displayed",                      // Text: Protection from Display specific string
		0x122: "Change title to specified",                                // Text: Change title
		0x14A: "Show floating text",                                       // Text: Float text
		0x152: "“Disable rest” message",                                   // Text: Disable rest
	}

	if _, ok := opcodeToExplanationMap[opcode]; ok {
		return param1, opcodeToExplanationMap[opcode]
	} else {
		return 0xFFFFFFFF, ""
	}
}

func (c *TextCollection) FillKnownContext() {
	for i := uint32(24916); i <= 24940; i++ {
		c.AddLabel(i, lb_cynicismQuote)
		c.AddLabel(i, lb_effect)
		c.AddContext(i, ContextEffect, "Show floating text", "common cynicism quotes")
	}
}

// Effect state names
func (c *TextCollection) LoadContextFromEffText2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		effName, ok := twoda.Get(rowKey, "EFFECT_NAME")
		if !ok {
			continue
		}

		strrefStr, ok := twoda.Get(rowKey, "STRREF")
		if !ok {
			continue
		}

		strrefSigned, err := strconv.ParseInt(strrefStr, 10, 32)
		if err != nil {
			continue
		}

		if strref := uint32(strrefSigned); strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_effect)
			c.AddContext(strref, ContextEffect, "EFFTEXT.2DA (state name shown in the game message window)", effName)
		}
	}
	return nil
}

// Engine's standard text references
func (c *TextCollection) LoadContextFromEngineSt2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		strrefStr, ok := twoda.Get(rowKey, "StrRef")
		if !ok {
			continue
		}

		strrefSigned, err := strconv.ParseInt(strrefStr, 10, 32)
		if err != nil {
			continue
		}

		if strref := uint32(strrefSigned); strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_ui)
			c.AddContext(strref, ContextUI, "ENGINEST.2DA (used in UI)", rowKey)
		}
	}
	return nil
}

// Magic dispel primary type
func (c *TextCollection) LoadContextFromMSchool2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		strrefStr, ok := twoda.Get(rowKey, "RES_REF")
		if !ok {
			continue
		}

		strref, err := strconv.ParseUint(strrefStr, 10, 32)
		if err != nil {
			continue
		}

		if strref := uint32(strref); strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_spell)
			c.AddContext(strref, ContextSpell, "MSCHOOL.2DA (the text that appears when magic is dispelled based on its primary type)", rowKey)
		}
	}
	return nil
}

// Magic dispel secondary type
func (c *TextCollection) LoadContextFromMSecType2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		strrefStr, ok := twoda.Get(rowKey, "RES_REF")
		if !ok {
			continue
		}

		strref, err := strconv.ParseUint(strrefStr, 10, 32)
		if err != nil {
			continue
		}

		if strref := uint32(strref); strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_spell)
			c.AddContext(strref, ContextSpell, "MSECTYPE.2DA (the text that appears when magic is dispelled based on its secondary type)", rowKey)
		}
	}
	return nil
}

func parseStrrefWithPrefix(s string) (uint32, bool, error) {
	isStandalone := strings.HasPrefix(s, "O_")
	if isStandalone {
		s = strings.TrimPrefix(s, "O_")
	}
	strref, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false, err
	}
	return uint32(strref), isStandalone, nil
}

// Ranger's Tracking skill
func (c *TextCollection) LoadContextFromTracking2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		strrefStr, ok := twoda.Get(rowKey, "STRREF")
		if !ok {
			continue
		}

		strref, isStandalone, err := parseStrrefWithPrefix(strrefStr)
		if err != nil {
			continue
		}

		if strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_ranger_tracking)
			if isStandalone {
				c.AddContext(strref, ContextTracking2DA, "The text displayed by the rangers Tracking skill", fmt.Sprintf("Area %s", rowKey))
			} else {
				c.AddLabel(67807, lb_ranger_tracking)
				c.AddContext(67807, ContextTracking2DA, "The default text displayed by the rangers Tracking skill. (See the tag). Possible variables to embed", strrefStr)

				c.AddContext(strref, ContextTracking2DA, "Used to embed into #67807 when the area is", rowKey)
			}
		}
	}
	return nil
}

// Starting equipment
func (c *TextCollection) LoadContextFrom25StWeap2DA(filename string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		if namerefStr, ok := twoda.Get(rowKey, "NAME_REF"); ok {
			if nameref, err := strconv.ParseUint(namerefStr, 10, 32); err == nil {
				nameref := uint32(nameref)
				if nameref != 0 && nameref != 0xFFFFFFFF {
					c.AddLabel(nameref, lb_item)
					c.AddContext(nameref, ContextItem, fmt.Sprintf("%s (starting equipment name) for slots", filename), rowKey)
				}
			}
		}

		if descrefStr, ok := twoda.Get(rowKey, "DESC_REF"); ok {
			if descref, err := strconv.ParseUint(descrefStr, 10, 32); err == nil {
				descref := uint32(descref)
				if descref != 0 && descref != 0xFFFFFFFF {
					c.AddLabel(descref, lb_item)
					c.AddContext(descref, ContextItem, fmt.Sprintf("%s (starting equipment description) for slots", filename), rowKey)
				}
			}
		}
	}

	return nil
}

// Credits
func (c *TextCollection) LoadContextFrom25ECred2DA(_ string, twoda *p.TwoDA) error {
	if row, ok := twoda.Row("DEFAULT"); ok {
		for i, strrefStr := range row {
			strref, err := strconv.ParseUint(strrefStr, 10, 32)
			if err != nil {
				continue
			}

			if strref := uint32(strref); strref != 0 && strref != 0xFFFFFFFF {
				c.AddLabel(strref, lb_ui)
				c.AddContext(strref, ContextUI, "25ECRED.2DA (credits text)", fmt.Sprintf("%s.BMP", twoda.GetByIndexOrDefault("BMP", i)))
			}
		}
		return nil
	} else {
		return fmt.Errorf("DEFAULT row not found in 25ECRED.2DA")
	}
}

// 7 eyes
func (c *TextCollection) LoadContextFrom7Eyes2DA(_ string, twoda *p.TwoDA) error {
	for _, rowKey := range twoda.RowKeys {
		strrefStr, ok := twoda.Get(rowKey, "STRREF")
		if !ok {
			continue
		}

		strref, err := strconv.ParseUint(strrefStr, 10, 32)
		if err != nil {
			continue
		}

		if strref := uint32(strref); strref != 0 && strref != 0xFFFFFFFF {
			c.AddLabel(strref, lb_spell)
			c.AddContext(strref, ContextSpell, "7EYES.2DA (Spell Effect: Seven Eyes). The text is shown when an effect is blocked by an active spellstate", rowKey)
		}
	}
	return nil
}

// Subtitles
func (c *TextCollection) LoadContextFromCharSnd2DA(_ string, twoda *p.TwoDA) error {
	// for _, rowKey := range twoda.RowKeys {
	// 	row, ok := twoda.Row(rowKey)
	// 	if !ok {
	// 		continue
	// 	}
		
	// 	for _, strrefStr := range row {
	// 		strref, err := strconv.ParseUint(strrefStr, 10, 32)
	// 		if err != nil {
	// 			continue
	// 		}

	// 		if strref := uint32(strref); strref != 0 && strref != 0xFFFFFFFF {
	// 			c.AddLabel(strref, lb_ui)
	// 			c.AddContext(strref, ContextUI, "CHARSND.2DA (subtitles)", rowKey)
	// 		}
	// 	}
	// }
	
	// [TODO] @GooRoo: Parse `SNDSLOT.IDS` to enrich the context with sound slot names.
	return nil
}
