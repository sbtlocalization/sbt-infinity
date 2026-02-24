// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dcanvas

import "testing"

func node(x, y, w, h int) *Node {
	return &Node{X: x, Y: y, Width: w, Height: h}
}

func TestHasOverlappingNodes_Empty(t *testing.T) {
	c := &Canvas{}
	if c.HasOverlappingNodes() {
		t.Error("expected false for empty canvas")
	}
}

func TestHasOverlappingNodes_Single(t *testing.T) {
	c := &Canvas{Nodes: []*Node{node(0, 0, 400, 300)}}
	if c.HasOverlappingNodes() {
		t.Error("expected false for single node")
	}
}

func TestHasOverlappingNodes_NoOverlap(t *testing.T) {
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(600, 0, 400, 300),
	}}
	if c.HasOverlappingNodes() {
		t.Error("expected false for non-overlapping nodes")
	}
}

func TestHasOverlappingNodes_Touching(t *testing.T) {
	// Touching at edge — not overlapping
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(400, 0, 400, 300),
	}}
	if c.HasOverlappingNodes() {
		t.Error("expected false for edge-touching nodes")
	}
}

func TestHasOverlappingNodes_Overlap(t *testing.T) {
	// Overlaps the first node by 1 pixel horizontally
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(399, 0, 400, 300),
	}}
	if !c.HasOverlappingNodes() {
		t.Error("expected true for overlapping nodes")
	}
}

func TestHasOverlappingNodes_OverlapDiagonal(t *testing.T) {
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(200, 150, 400, 300),
	}}
	if !c.HasOverlappingNodes() {
		t.Error("expected true for diagonally overlapping nodes")
	}
}

func TestHasOverlappingNodes_NoOverlapVertical(t *testing.T) {
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(0, 500, 400, 300),
	}}
	if c.HasOverlappingNodes() {
		t.Error("expected false for vertically separated nodes")
	}
}

func TestHasOverlappingNodes_TouchingVertical(t *testing.T) {
	// Touching vertically — not overlapping
	c := &Canvas{Nodes: []*Node{
		node(0, 0, 400, 300),
		node(0, 300, 400, 300),
	}}
	if c.HasOverlappingNodes() {
		t.Error("expected false for vertically touching nodes")
	}
}
