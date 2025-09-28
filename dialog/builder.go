// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"log"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	p "github.com/sbtlocalization/infinity-tools/parser"
	"github.com/spf13/afero"
)

type DialogBuilder struct {
	dlgFsys        afero.Fs
	tlkFsys        afero.Fs
	loadedDlgFiles map[string]*p.DlgFile
	tlkFile        *p.TlkFile
}

func NewDialogBuilder(dlgFsys afero.Fs, tlkFsys afero.Fs) *DialogBuilder {
	return &DialogBuilder{
		dlgFsys:        dlgFsys,
		tlkFsys:        tlkFsys,
		loadedDlgFiles: make(map[string]*p.DlgFile),
		tlkFile:        nil,
	}
}

func (b *DialogBuilder) LoadAllRootStates(dlgNames ...string) (*DialogCollection, error) {
	collection := NewDialogCollection()

	for _, dlgName := range dlgNames {
		dlg, err := b.readDlgFile(dlgName)
		if err != nil {
			log.Printf("error loading DLG file %s: %v", dlgName, err)
			continue
		}

		states, err := dlg.States()
		if err != nil {
			log.Printf("error getting states from DLG file %s: %v", dlgName, err)
			continue
		}

		for stateIndex, stateEntry := range states {
			if _, isRoot := stateEntry.GetTriggerText(); isRoot {
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

	if b.tlkFile == nil || b.tlkFile.FileName() != tlkName {
		tlkFile, err := b.readTlkFile(tlkName)
		if err != nil {
			return nil, fmt.Errorf("error loading TLK file %s: %v", tlkName, err)
		}
		b.tlkFile = tlkFile
	}

	for _, dlgName := range dlgNames {
		dlg, err := b.readDlgFile(dlgName)
		if err != nil {
			log.Printf("Error loading DLG file %s: %v", dlgName, err)
			continue
		}

		b.loadDialogs(dlgName, dlg, collection)

		for _, loadedDlg := range b.loadedDlgFiles {
			loadedDlg.Close()
		}
		clear(b.loadedDlgFiles)
	}

	return collection, nil
}

func (b *DialogBuilder) readTlkFile(tlkFileName string) (*p.TlkFile, error) {
	file, err := b.tlkFsys.Open(tlkFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open TLK file %s: %w", tlkFileName, err)
	}

	tlk := p.NewTlk()
	stream := kaitai.NewStream(file)
	err = tlk.Read(stream, nil, tlk)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLK file %s: %w", tlkFileName, err)
	}

	tlkFile := &p.TlkFile{
		Tlk:  tlk,
		File: file,
	}
	return tlkFile, nil
}

func (b *DialogBuilder) readDlgFile(dlgFileName string) (*p.DlgFile, error) {
	if dlgFile, exists := b.loadedDlgFiles[dlgFileName]; exists {
		return dlgFile, nil
	}

	fullName := dlgFileName
	if !strings.HasSuffix(strings.ToLower(fullName), ".dlg") {
		fullName = fullName + ".DLG"
	}

	file, err := b.dlgFsys.Open(fullName)
	if err != nil {
		return nil, fmt.Errorf("failed to open DLG file %s: %w", fullName, err)
	}

	dlg := p.NewDlg()
	stream := kaitai.NewStream(file)
	err = dlg.Read(stream, nil, dlg)
	if err != nil {
		return nil, fmt.Errorf("failed to read DLG file %s: %w", dlgFileName, err)
	}

	dlgFile := &p.DlgFile{
		Dlg:  dlg,
		File: file,
	}
	b.loadedDlgFiles[dlgFileName] = dlgFile
	return dlgFile, nil
}

func (b *DialogBuilder) loadDialogs(dlgName string, dlgFile *p.DlgFile, collection *DialogCollection) {
	states, err := dlgFile.States()
	if err != nil {
		log.Printf("Error getting states from DLG file %s: %v", dlgName, err)
		return
	}

	for stateIndex, stateEntry := range states {
		if _, isRoot := stateEntry.GetTriggerText(); isRoot {
			d, err := b.loadDialog(NewNodeOrigin(dlgName, uint32(stateIndex)))
			if err != nil {
				log.Printf("Error loading dialog for state %d: %v", stateIndex, err)
				continue
			}
			collection.Dialogs = append(collection.Dialogs, d)
		}
	}
}

func (b *DialogBuilder) loadDialog(rootStateOrigin NodeOrigin) (*Dialog, error) {
	dialog := NewDialog(rootStateOrigin)
	rootState, err := b.loadTree(dialog, make([]NodeOrigin, 0), rootStateOrigin)
	if err != nil {
		log.Fatalln("Can't load root state:", err)
		return nil, err
	}

	dialog.RootState = rootState

	return dialog, nil
}

func (b *DialogBuilder) loadTree(dialog *Dialog, previousStates []NodeOrigin, stateOrigin NodeOrigin) (*Node, error) {
	state, err := b.loadStateWithTransitions(stateOrigin)
	if err != nil {
		return &Node{
			Type:     ErrorNodeType,
			Origin:   stateOrigin,
			Parent:   nil,
			Children: make([]*Node, 0),
		}, nil
	}
	dialog.AllStates[stateOrigin] = struct{}{}

	for _, child := range state.Children {
		if child.Type == TransitionNodeType && !child.Transition.IsDialogEnd {
			nextStateOrigin := child.Transition.NextStateOrigin
			if _, ok := dialog.AllStates[nextStateOrigin]; ok {
				child.Children[0] = &Node{
					Type:     LoopNodeType,
					Origin:   nextStateOrigin,
					Parent:   child,
					Children: make([]*Node, 0),
				}
			} else {
				nextState, err := b.loadTree(dialog, append(previousStates, stateOrigin), nextStateOrigin)
				if err != nil {
					log.Printf("Error loading state %s: %v", nextStateOrigin, err)
					continue
				}
				nextState.Parent = child
				if len(child.Children) != 1 {
					log.Fatalln("Transition node should have exactly one child")
				}
				child.Children[0] = nextState
			}
		}
	}

	return state, nil
}

func (b *DialogBuilder) loadStateWithTransitions(origin NodeOrigin) (*Node, error) {
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
	trigger, triggerExists := state.GetTriggerText()
	stateNode := &Node{
		Type:     StateNodeType,
		Origin:   origin,
		Parent:   nil,
		Children: make([]*Node, state.NumTransitions),
		State: &StateData{
			TextRef:    state.TextRef,
			Text:       b.tlkFile.GetText(state.TextRef),
			HasTrigger: triggerExists,
			Trigger:    trigger,
		},
	}

	transitions, err := state.Transitions()
	if err != nil {
		log.Printf("Error getting transitions from state %s: %v", origin, err)
		return nil, err
	}

	for index, transition := range transitions {
		transitionOrigin := NewNodeOrigin(origin.DlgName, state.FirstTransitionIndex+uint32(index))
		trigger, withTrigger := transition.GetTriggerText()
		textRef, text, withText := transition.GetText(b.tlkFile)
		journalTextRef, journalText, withJournalText := transition.GetJournalText(b.tlkFile)
		action, withAction := transition.GetActionText()

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
				HasJournalText:  withJournalText,
				JournalTextRef:  journalTextRef,
				JournalText:     journalText,
				HasTrigger:      withTrigger,
				Trigger:         trigger,
				HasAction:       withAction,
				Action:          action,
				IsDialogEnd:     transition.Flags.DialogEnd,
				NextStateOrigin: nextStateOrigin,
			},
			Parent:   stateNode,
			Children: make([]*Node, numChildren),
		}
		stateNode.Children[index] = transitionNode
	}

	return stateNode, nil
}
