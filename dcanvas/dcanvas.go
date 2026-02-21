// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

// Package dcanvas provides types for the dCanvas 2.0 format, a strict superset
// of JSON Canvas 1.0 with x- extension fields for game dialog metadata.
package dcanvas

// Canvas is the top-level dCanvas 2.0 document.
type Canvas struct {
	Version string  `json:"x-dCanvasVersion"`
	Nodes   []*Node `json:"nodes"`
	Edges   []*Edge `json:"edges"`
}

// Node represents a canvas node with dCanvas 2.0 extension fields.
type Node struct {
	// Standard JSON Canvas 1.0 fields
	ID     string `json:"id"`
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Color  string `json:"color,omitempty"`
	Text   string `json:"text"`

	// dCanvas 2.0 extension fields
	NodeRole      string     `json:"x-nodeRole"`
	NodeId        string     `json:"x-nodeId,omitempty"`
	Character     *Character `json:"x-character,omitempty"`
	TextId        string     `json:"x-textId,omitempty"`
	JournalTextId string     `json:"x-journalTextId,omitempty"`
	JournalText   string     `json:"x-journalText,omitempty"`
	Trigger       string     `json:"x-trigger,omitempty"`
	Action        string     `json:"x-action,omitempty"`
}

// Edge represents a connection between two nodes with an optional condition.
type Edge struct {
	// Standard JSON Canvas 1.0 fields
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	FromSide string `json:"fromSide,omitempty"`
	FromEnd  string `json:"fromEnd,omitempty"`
	ToNode   string `json:"toNode"`
	ToSide   string `json:"toSide,omitempty"`
	ToEnd    string `json:"toEnd,omitempty"`
	Color    string `json:"color,omitempty"`
	Label    string `json:"label,omitempty"`

	// dCanvas 2.0 extension field
	Condition string `json:"x-condition,omitempty"`
}

// Character holds speaker information for a dialog node.
type Character struct {
	Name     string `json:"name"`
	Portrait string `json:"portrait,omitempty"`
}
