// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/afero"
)

type DialogBuilder struct {
	infFsys        afero.Fs
	tlkFsys        afero.Fs
	loadedDlgFiles map[string]*p.DlgFile
	dlgNames       map[string]string
	tlkFile        *p.TlkFile
	creatures      map[string][]*Creature
	verbose        bool
}

func NewDialogBuilder(infFsys afero.Fs, tlkFsys afero.Fs, withCreatures, verbose bool) *DialogBuilder {
	var creatures map[string][]*Creature = nil
	if tlkFsys == nil {
		withCreatures = false
	}
	if withCreatures {
		creatures = make(map[string][]*Creature)
	}
	return &DialogBuilder{
		infFsys:        infFsys,
		tlkFsys:        tlkFsys,
		loadedDlgFiles: make(map[string]*p.DlgFile),
		dlgNames:       make(map[string]string),
		tlkFile:        nil,
		creatures:      creatures,
		verbose:        verbose,
	}
}

func (b *DialogBuilder) LoadAllRootStates(dlgNames ...string) (*DialogCollection, error) {
	collection := NewDialogCollection()

	for _, dlgName := range dlgNames {
		dlg, err := b.readDlgFile(dlgName)
		if err != nil {
			b.Printf("error loading DLG file %s: %v", dlgName, err)
			continue
		}

		dlgName = b.canonicalDlgName(dlgName)
		states, err := dlg.States()
		if err != nil {
			b.Printf("error getting states from DLG file %s: %v", dlgName, err)
			continue
		}

		for stateIndex, stateEntry := range states {
			_, isRoot, err := stateEntry.GetTriggerText()
			if err != nil {
				b.Printf("warning: state %d in DLG file %s: %v\n", stateIndex, dlgName, err)
			}
			if isRoot {
				d := NewDialog(NewNodeOrigin(dlgName, uint32(stateIndex)))
				collection.Dialogs = append(collection.Dialogs, d)
			}
		}
		for _, loadedDlg := range b.loadedDlgFiles {
			loadedDlg.Close()
		}
		clear(b.loadedDlgFiles)
	}

	return collection, nil
}

func (b *DialogBuilder) LoadAllDialogs(tlkName string, dlgNames ...string) (*DialogCollection, error) {
	collection := NewDialogCollection()

	if b.tlkFsys != nil && (b.tlkFile == nil || b.tlkFile.FileName() != tlkName) {
		tlkFile, err := b.readTlkFile(tlkName)
		if err != nil {
			return nil, fmt.Errorf("error loading TLK file %s: %v", tlkName, err)
		}
		b.tlkFile = tlkFile
	}

	if b.creatures != nil {
		_ = b.loadCreatures()
	}

	for _, dlgName := range dlgNames {
		dlg, err := b.readDlgFile(dlgName)
		if err != nil {
			b.Printf("error loading DLG file %s: %v", dlgName, err)
			continue
		}

		b.loadDialogs(b.canonicalDlgName(dlgName), dlg, collection)

		for _, loadedDlg := range b.loadedDlgFiles {
			loadedDlg.Close()
		}
		clear(b.loadedDlgFiles)
	}

	return collection, nil
}

func (b *DialogBuilder) loadIds(name string) *p.Ids {
	file, err := b.infFsys.Open(name)
	if err != nil {
		b.Printf("warning: unable to open %s: %v\n", name, err)
		return nil
	}
	defer file.Close()

	ids, err := p.ParseIds(file)
	if err != nil {
		b.Printf("warning: unable to parse %s: %v\n", name, err)
		return nil
	}
	return ids
}

func (b *DialogBuilder) loadCreatures() error {
	if len(b.creatures) > 0 {
		// already loaded
		return nil
	}

	if b.tlkFile == nil {
		return fmt.Errorf("TLK file must be loaded before loading creatures")
	}

	genderIds := b.loadIds("GENDER.IDS")

	dir, err := b.infFsys.Open("CRE")
	if err != nil {
		return fmt.Errorf("unable to read CRE files: %v", err)
	}
	defer dir.Close()

	creFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read CRE file names: %v", err)
	}

	for _, creFileName := range creFiles {
		creFile, err := b.infFsys.Open(creFileName)
		if err != nil {
			b.Printf("unable to open CRE file %s: %v", creFileName, err)
			continue
		}
		cre := p.NewCre()
		stream := kaitai.NewStream(creFile)
		err = cre.Read(stream, nil, cre)
		creFile.Close()
		if err != nil {
			b.Printf("unable to read CRE file %s: %v", creFileName, err)
			continue
		}
		if cre.Body == nil || cre.Body.Header == nil {
			b.Printf("unsupported CRE version in %s\n", creFileName)
			continue
		}

		shortName, longName := "", ""
		hasShortName := cre.ShortNameRef < b.tlkFile.NumEntries
		hasLongName := cre.LongNameRef < b.tlkFile.NumEntries
		if hasShortName {
			shortName = b.tlkFile.GetText(cre.ShortNameRef)
		}
		if hasLongName {
			longName = b.tlkFile.GetText(cre.LongNameRef)
		}

		creature := &Creature{
			ResourceName:   strings.ToUpper(creFileName),
			HasLongNameId:  hasLongName,
			HasShortNameId: hasShortName,
			ShortNameId:    cre.ShortNameRef,
			ShortName:      shortName,
			LongNameId:     cre.LongNameRef,
			LongName:       longName,
		}

		if cre.Version == "V1.0" || cre.Version == "V1.1" {
			creature.Portrait = cre.Body.Header.SmallPortrait + ".BMP"
			creature.Dialog = cre.Body.Header.Dialog
			if genderIds != nil {
				creature.Sex = genderIds.Entries[int32(cre.Body.Header.Sex)]
			}
		}

		b.Printf("Loading character: %s (dialog: %s)\n", longName, cre.Body.Header.Dialog)

		key := strings.ToUpper(creature.Dialog)
		b.creatures[key] = append(b.creatures[key], creature)
	}
	for _, creatures := range b.creatures {
		slices.SortFunc(creatures, func(a, b *Creature) int { return strings.Compare(a.ResourceName, b.ResourceName) })
	}

	return nil
}

func (b *DialogBuilder) Close() error {
	if b.tlkFile == nil {
		return nil
	}
	err := b.tlkFile.Close()
	b.tlkFile = nil
	return err
}

func (b *DialogBuilder) readTlkFile(tlkFileName string) (*p.TlkFile, error) {
	return p.ReadTlkFile(b.tlkFsys, tlkFileName)
}

func (b *DialogBuilder) canonicalDlgName(name string) string {
	origin := NewNodeOrigin(name, 0)
	if canonical, ok := b.dlgNames[strings.ToLower(origin.DlgName)]; ok {
		return canonical
	}
	return origin.DlgName
}

func (b *DialogBuilder) readDlgFile(dlgFileName string) (*p.DlgFile, error) {
	key := strings.ToLower(NewNodeOrigin(dlgFileName, 0).DlgName)
	if dlgFile, exists := b.loadedDlgFiles[key]; exists {
		return dlgFile, nil
	}

	fullName := dlgFileName
	if !strings.HasSuffix(strings.ToLower(fullName), ".dlg") {
		fullName = fullName + ".DLG"
	}

	file, err := b.infFsys.Open(fullName)
	if err != nil {
		return nil, fmt.Errorf("failed to open DLG file %s: %w", fullName, err)
	}

	b.dlgNames[key] = NewNodeOrigin(filepath.Base(file.Name()), 0).DlgName
	dlg := p.NewDlg()
	stream := kaitai.NewStream(file)
	err = dlg.Read(stream, nil, dlg)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read DLG file %s: %w", dlgFileName, err)
	}

	dlgFile := &p.DlgFile{
		Dlg:  dlg,
		File: file,
	}
	b.loadedDlgFiles[key] = dlgFile
	return dlgFile, nil
}

func (b *DialogBuilder) loadDialogs(dlgName string, dlgFile *p.DlgFile, collection *DialogCollection) {
	states, err := dlgFile.States()
	if err != nil {
		b.Printf("error getting states from DLG file %s: %v", dlgName, err)
		return
	}

	for stateIndex, stateEntry := range states {
		_, isRoot, err := stateEntry.GetTriggerText()
		if err != nil {
			b.Printf("warning: state %d in DLG file %s: %v\n", stateIndex, dlgName, err)
		}
		if isRoot {
			d, err := b.loadDialog(NewNodeOrigin(dlgName, uint32(stateIndex)))
			if err != nil {
				b.Printf("error loading dialog for state %d: %v", stateIndex, err)
				continue
			}
			collection.Dialogs = append(collection.Dialogs, d)
		}
	}
}

func (b *DialogBuilder) loadDialog(rootStateOrigin NodeOrigin) (*Dialog, error) {
	dialog := NewDialog(rootStateOrigin)
	rootState, err := b.loadTree(dialog, make(map[NodeOrigin]bool), make(map[NodeOrigin]bool), rootStateOrigin)
	if err != nil {
		b.Printf("can't load root state: %v", err)
		return nil, err
	}

	dialog.RootState = rootState

	return dialog, nil
}

func stateKey(origin NodeOrigin) NodeOrigin {
	origin.DlgName = strings.ToLower(origin.DlgName)
	return origin
}

func (b *DialogBuilder) loadTree(dialog *Dialog, active, expanded map[NodeOrigin]bool, stateOrigin NodeOrigin) (*Node, error) {
	state, err := b.loadStateWithTransitions(dialog, stateOrigin)
	if err != nil {
		return &Node{
			Type:     ErrorNodeType,
			Origin:   stateOrigin,
			Parent:   nil,
			Children: make([]*Node, 0),
			Dialog:   dialog,
		}, nil
	}
	dialog.AllStates[stateOrigin] = struct{}{}
	key := stateKey(stateOrigin)
	active[key] = true
	expanded[key] = true
	defer delete(active, key)

	if creatures, ok := b.creatures[strings.ToUpper(stateOrigin.DlgName)]; ok {
		dialog.AllCreatures[stateOrigin.DlgName] = creatures
	}

	for _, child := range state.Children {
		if child.Type == TransitionNodeType && !child.Transition.IsDialogEnd {
			nextStateOrigin := child.Transition.NextStateOrigin
			_, _ = b.readDlgFile(nextStateOrigin.DlgName)
			nextStateOrigin.DlgName = b.canonicalDlgName(nextStateOrigin.DlgName)
			child.Transition.NextStateOrigin = nextStateOrigin
			key := stateKey(nextStateOrigin)
			if expanded[key] {
				kind := ConvergenceNodeType
				if active[key] {
					kind = LoopNodeType
				}
				child.Children[0] = &Node{
					Type:     kind,
					Origin:   nextStateOrigin,
					Parent:   child,
					Children: make([]*Node, 0),
					Dialog:   dialog,
				}
			} else {
				nextState, err := b.loadTree(dialog, active, expanded, nextStateOrigin)
				if err != nil {
					b.Printf("error loading state %s: %v", nextStateOrigin, err)
					continue
				}
				nextState.Parent = child
				if len(child.Children) != 1 {
					return nil, fmt.Errorf("transition node should have exactly one child")
				}
				child.Children[0] = nextState
			}
		}
	}

	return state, nil
}

func (b *DialogBuilder) loadStateWithTransitions(dialog *Dialog, origin NodeOrigin) (*Node, error) {
	dlg, err := b.readDlgFile(origin.DlgName)
	if err != nil {
		return nil, fmt.Errorf("error loading DLG file %s: %v", origin.DlgName, err)
	}

	states, err := dlg.States()
	if err != nil {
		return nil, fmt.Errorf("error getting states from DLG file %s: %v", origin.DlgName, err)
	}
	if int(origin.Index) >= len(states) {
		return nil, fmt.Errorf("state index %d out of range in DLG file %s", origin.Index, origin.DlgName)
	}

	state := states[origin.Index]
	trigger, triggerExists, err := state.GetTriggerText()
	if err != nil {
		b.Printf("warning: state %s: %v\n", origin, err)
	}
	text := ""
	sound := ""
	if b.tlkFile != nil {
		text = b.tlkFile.GetText(state.TextRef)
		sound = b.tlkFile.GetSound(state.TextRef)
	}
	stateNode := &Node{
		Type:     StateNodeType,
		Origin:   origin,
		Parent:   nil,
		Children: make([]*Node, state.NumTransitions),
		Dialog:   dialog,
		State: &StateData{
			TextRef:    state.TextRef,
			Text:       text,
			Sound:      sound,
			HasTrigger: triggerExists,
			Trigger:    trigger,
		},
	}

	transitions, err := state.Transitions()
	if err != nil {
		b.Printf("error getting transitions from state %s: %v", origin, err)
		return nil, err
	}

	for index, transition := range transitions {
		transitionOrigin := NewNodeOrigin(origin.DlgName, state.FirstTransitionIndex+uint32(index))
		trigger, withTrigger, err := transition.GetTriggerText()
		if err != nil {
			b.Printf("warning: transition %s: %v\n", transitionOrigin, err)
		}
		textRef, text, withText := transition.GetText(b.tlkFile)
		journalTextRef, journalText, withJournalText := transition.GetJournalText(b.tlkFile)
		action, withAction, err := transition.GetActionText()
		if err != nil {
			b.Printf("warning: transition %s: %v\n", transitionOrigin, err)
		}

		var transitionSound, journalSound string
		if b.tlkFile != nil {
			if withText {
				transitionSound = b.tlkFile.GetSound(textRef)
			}
			if withJournalText {
				journalSound = b.tlkFile.GetSound(journalTextRef)
			}
		}

		numChildren := 1
		if transition.Flags.DialogEnd {
			numChildren = 0
		}

		nextStateOrigin := NewNodeOrigin(transition.NextStateResource, transition.NextStateIndex)

		transitionNode := &Node{
			Type:   TransitionNodeType,
			Origin: transitionOrigin,
			Transition: &TransitionData{
				HasText:         withText,
				TextRef:         textRef,
				Text:            text,
				Sound:           transitionSound,
				HasJournalText:  withJournalText,
				JournalTextRef:  journalTextRef,
				JournalText:     journalText,
				JournalSound:    journalSound,
				HasTrigger:      withTrigger,
				Trigger:         trigger,
				HasAction:       withAction,
				Action:          action,
				IsDialogEnd:     transition.Flags.DialogEnd,
				NextStateOrigin: nextStateOrigin,
			},
			Parent:   stateNode,
			Children: make([]*Node, numChildren),
			Dialog:   dialog,
		}
		stateNode.Children[index] = transitionNode
	}

	return stateNode, nil
}

func (b *DialogBuilder) Printf(format string, args ...any) {
	if b.verbose {
		fmt.Printf(format, args...)
	}
}
