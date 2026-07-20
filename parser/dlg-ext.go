// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"fmt"

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

func (s *Dlg_StateEntry) GetTriggerText() (string, bool, error) {
	trigger, err := s.Trigger()
	if err != nil {
		return "", false, fmt.Errorf("failed to read state trigger: %w", err)
	}
	if trigger == nil {
		return "", false, nil
	}

	text, err := trigger.Text()
	if err != nil {
		return "", false, fmt.Errorf("failed to read state trigger text: %w", err)
	}

	return text, true, nil
}

func (t *Dlg_TransitionEntry) GetTriggerText() (string, bool, error) {
	if !t.Flags.WithTrigger {
		return "", false, nil
	}

	trigger, err := t.Trigger()
	if err != nil {
		return "", false, fmt.Errorf("failed to read transition trigger: %w", err)
	}
	if trigger == nil {
		return "", false, fmt.Errorf("transition promises a trigger but none was read")
	}

	text, err := trigger.Text()
	if err != nil {
		return "", false, fmt.Errorf("failed to read transition trigger text: %w", err)
	}

	return text, true, nil
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

func (t *Dlg_TransitionEntry) GetActionText() (string, bool, error) {
	if !t.Flags.WithAction {
		return "", false, nil
	}

	action, err := t.Action()
	if err != nil {
		return "", false, fmt.Errorf("failed to read transition action: %w", err)
	}
	if action == nil {
		return "", false, fmt.Errorf("transition promises an action but none was read")
	}

	text, err := action.Text()
	if err != nil {
		return "", false, fmt.Errorf("failed to read transition action text: %w", err)
	}

	return text, true, nil
}
