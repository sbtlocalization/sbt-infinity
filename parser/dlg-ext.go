// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"log"

	"github.com/spf13/afero"
)

type DlgFile struct {
	*Dlg
	File afero.File
}

func (d *DlgFile) FileName() string {
	return d.File.Name()
}

func (d *DlgFile) Close() error {
	return d.File.Close()
}

func (s *Dlg_StateEntry) GetTriggerText() (string, bool) {
	trigger, err := s.Trigger()
	if trigger == nil || err != nil {
		return "", false
	}

	text, err := trigger.Text()
	if err != nil {
		log.Println("Error getting trigger text for state entry:", err)
		return "", false
	}

	return text, true
}

func (t *Dlg_TransitionEntry) GetTriggerText() (string, bool) {
	if !t.Flags.WithTrigger {
		return "", false
	}

	trigger, err := t.Trigger()
	if trigger == nil || err != nil {
		log.Fatalln("There is no trigger although there should be one:", err)
		return "", false
	}

	text, err := trigger.Text()
	if err != nil {
		log.Println("Error getting trigger text for transition entry:", err)
		return "", false
	}

	return text, true
}

func (t *Dlg_TransitionEntry) GetText(tlkFile *TlkFile) (uint32, string, bool) {
	if !t.Flags.WithText {
		return 0xFFFFFFFF, "", false
	} else {
		text := ""
		if tlkFile != nil {
			text = tlkFile.GetText(t.TextRef)
		}
		return t.TextRef, text, true
	}
}

func (t *Dlg_TransitionEntry) GetJournalText(tlkFile *TlkFile) (uint32, string, bool) {
	if !t.Flags.WithJournalEntry {
		return 0xFFFFFFFF, "", false
	} else {
		text := ""
		if tlkFile != nil {
			text = tlkFile.GetText(t.JournalTextRef)
		}
		return t.JournalTextRef, text, true
	}
}

func (t *Dlg_TransitionEntry) GetActionText() (string, bool) {
	if !t.Flags.WithAction {
		return "", false
	}

	action, err := t.Action()
	if err != nil {
		log.Fatalln("There is no action although there should be one:", err)
		return "", false
	}

	text, err := action.Text()
	if err != nil {
		log.Println("Error getting action text for transition entry:", err)
		return "", false
	}

	return text, true
}
