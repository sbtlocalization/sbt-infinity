// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"testing"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/sbt-infinity/dialog"
)

func TestExportLineAndTerminalReply(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	root := &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, Dialog: d, State: &dialog.StateData{TextRef: 42, Text: "Hello", Trigger: "  True()  ", Sound: "HELLO"}}
	reply := &dialog.Node{Type: dialog.TransitionNodeType, Origin: dialog.NewNodeOrigin("TEST", 1), Dialog: d, Parent: root, Transition: &dialog.TransitionData{IsDialogEnd: true}}
	root.Children = []*dialog.Node{reply}
	d.RootState = root

	canvas, err := d.ToDCanvas(dialog.FormatOptions{SoundPrefix: "sounds/", SoundSuffix: ".wav"})
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := dcanvas.Encode(canvas, &encoded); err != nil {
		t.Fatal(err)
	}
	canvas, err = dcanvas.Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	if canvas.Version != "1.1" {
		t.Fatalf("version = %q", canvas.Version)
	}
	if len(canvas.Nodes) != 2 {
		t.Fatalf("nodes = %d", len(canvas.Nodes))
	}
	line, end := canvas.Nodes[0], canvas.Nodes[1]
	if line.Kind != dcanvas.KindLine || line.DomainID != "TEST[0]" || line.TextID != "#42" || line.Condition != "True()" || line.Sound != "sounds/HELLO.wav" {
		t.Errorf("line = %+v", line)
	}
	if end.Kind != dcanvas.KindReply || end.Role != "dialog-end" || end.Character != nil || end.Text != "" {
		t.Errorf("end = %+v", end)
	}
	if len(canvas.Edges) != 1 || canvas.Edges[0].Kind != dcanvas.KindNormal {
		t.Errorf("edges = %+v", canvas.Edges)
	}
}
