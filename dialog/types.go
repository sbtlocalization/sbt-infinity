// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"strings"
)

type NodeType int

const (
	StateNodeType NodeType = iota
	TransitionNodeType
	LoopNodeType
	ErrorNodeType
)

type NodeOrigin struct {
	DlgName string // DLG filename without extension
	Index   uint32 // Index within that file
}

func NewNodeOrigin(dlgName string, index uint32) NodeOrigin {
	if strings.HasSuffix(strings.ToLower(dlgName), ".dlg") {
		dlgName = dlgName[:len(dlgName)-4]
	}
	return NodeOrigin{DlgName: dlgName, Index: index}
}

func (t NodeType) String() string {
	switch t {
	case StateNodeType:
		return "state"
	case TransitionNodeType:
		return "transition"
	case LoopNodeType:
		return "state"
	case ErrorNodeType:
		return "error"
	default:
		return "unknown"
	}
}

func (o NodeOrigin) String() string {
	return fmt.Sprintf("%s[%d]", o.DlgName, o.Index)
}

type Node struct {
	Type     NodeType
	Origin   NodeOrigin
	Parent   *Node
	Children []*Node
	Dialog   *Dialog

	State      *StateData
	Transition *TransitionData
}

type StateData struct {
	TextRef    uint32
	Text       string
	HasTrigger bool
	Trigger    string
}

type TransitionData struct {
	HasText         bool
	TextRef         uint32
	Text            string
	HasJournalText  bool
	JournalTextRef  uint32
	JournalText     string
	HasTrigger      bool
	Trigger         string
	HasAction       bool
	Action          string
	IsDialogEnd     bool
	NextStateOrigin NodeOrigin
}

func (n *Node) String() string {
	return fmt.Sprintf("%s-%s", n.Type, n.Origin)
}

type DialogID = NodeOrigin

type Dialog struct {
	Id           DialogID
	RootState    *Node
	AllStates    map[NodeOrigin]struct{}
	AllCreatures map[string]*Creature
}

func NewDialog(id DialogID) *Dialog {
	return &Dialog{
		Id:           id,
		RootState:    nil,
		AllStates:    make(map[NodeOrigin]struct{}),
		AllCreatures: make(map[string]*Creature),
	}
}

type DialogCollection struct {
	Dialogs []*Dialog
}

func NewDialogCollection() *DialogCollection {
	return &DialogCollection{
		Dialogs: make([]*Dialog, 0),
	}
}

type Creature struct {
	ShortNameId uint32
	ShortName   string
	LongNameId  uint32
	LongName    string
	Portrait    string
	Dialog      string
}
