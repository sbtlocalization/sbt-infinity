package parser

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
)

type DlgFile struct {
	*Dlg
	File afero.File
}

type TlkFile struct {
	*Tlk
	File afero.File
}

func (d *DlgFile) FileName() string {
	return d.File.Name()
}

func (d *DlgFile) Close() error {
	return d.File.Close()
}

func (t *TlkFile) GetText(strref uint32) string {
	invalid_result := fmt.Sprintf("<invalid text reference #%d>", strref)

	if t == nil || t.Tlk == nil {
		log.Fatal("TLK file is not loaded")
		return invalid_result
	}

	if strref == 0xFFFFFFFF || strref > t.NumEntries {
		log.Printf("TLK entry #%d does not exist", strref)
		return invalid_result
	}

	text, err := t.Entries[strref].Text()
	if err != nil {
		log.Printf("Error retrieving TLK text for entry #%d: %v", strref, err)
		return invalid_result
	}

	return text
}

func (t *TlkFile) FileName() string {
	return t.File.Name()
}

func (t *TlkFile) Close() error {
	return t.File.Close()
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
		return t.TextRef, tlkFile.GetText(t.TextRef), true
	}
}

func (t *Dlg_TransitionEntry) GetJournalText(tlkFile *TlkFile) (uint32, string, bool) {
	if !t.Flags.WithJournalEntry {
		return 0xFFFFFFFF, "", false
	} else {
		return t.JournalTextRef, tlkFile.GetText(t.JournalTextRef), true
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
