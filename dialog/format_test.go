// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/sbt-infinity/dialog"
)

func TestEscapedFragmentCannotMatchAnotherLiteralDomainID(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("ROOT", 0))
	n := &dialog.Node{Type: dialog.StateNodeType, Origin: dialog.NewNodeOrigin("BaR%", 0), Dialog: d}
	link := n.ToUrl("https://example.test")
	fragment := strings.SplitN(link, "#", 2)[1]
	if fragment == "state-BaR%25-0" {
		t.Fatalf("raw fragment collides with another resource's d-id: %s", fragment)
	}
	decoded, err := url.PathUnescape(fragment)
	if err != nil || decoded != "state-BaR%-0" {
		t.Fatalf("fragment %q decoded to %q: %v", fragment, decoded, err)
	}
}

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
	if line.Kind != dcanvas.KindLine || line.DomainID != "state-0" || line.TextID != "#42" || line.Condition != "True()" || line.Sound != "sounds/HELLO.wav" {
		t.Errorf("line = %+v", line)
	}
	if end.Kind != dcanvas.KindReply || end.Role != "dialog-end" || end.Character == nil || end.Character.Name != "End dialog" || end.Text != "(no text)" || end.TextID != "" {
		t.Errorf("end = %+v", end)
	}
	if len(canvas.Edges) != 1 || canvas.Edges[0].Kind != dcanvas.KindNormal {
		t.Errorf("edges = %+v", canvas.Edges)
	}
}

func TestExportTextPlaceholder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		node       *dialog.Node
		wantText   string
		wantTextID string
		wantName   string
	}{
		{"empty reply without reference", &dialog.Node{Type: dialog.TransitionNodeType, Transition: &dialog.TransitionData{HasAction: true, Action: "Do()"}}, "(no text)", "", "Answer"},
		{"empty terminal without reference", &dialog.Node{Type: dialog.TransitionNodeType, Transition: &dialog.TransitionData{IsDialogEnd: true}}, "(no text)", "", "End dialog"},
		{"empty reply with reference", &dialog.Node{Type: dialog.TransitionNodeType, Transition: &dialog.TransitionData{HasText: true, TextRef: 42}}, "", "#42", "Answer"},
		{"empty state with reference", &dialog.Node{Type: dialog.StateNodeType, State: &dialog.StateData{TextRef: 42}}, "", "#42", "FOREIGN"},
		{"whitespace reply", &dialog.Node{Type: dialog.TransitionNodeType, Transition: &dialog.TransitionData{HasText: true, TextRef: 42, Text: " \n"}}, " \n", "#42", "Answer"},
		{"invalid reference diagnostic", &dialog.Node{Type: dialog.StateNodeType, State: &dialog.StateData{TextRef: ^uint32(0), Text: "<invalid text reference #4294967295>"}}, "<invalid text reference #4294967295>", "#4294967295", "FOREIGN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := dialog.NewDialog(dialog.NewNodeOrigin("ROOT", 0))
			tc.node.Origin = dialog.NewNodeOrigin("FOREIGN", 18)
			tc.node.Dialog = d
			d.RootState = tc.node
			canvas, err := d.ToDCanvas(dialog.FormatOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var encoded bytes.Buffer
			if err := dcanvas.Encode(canvas, &encoded); err != nil {
				t.Fatal(err)
			}
			assertSchema(t, encoded.Bytes())
			var document struct {
				Nodes []map[string]any `json:"nodes"`
			}
			if err := json.Unmarshal(encoded.Bytes(), &document); err != nil {
				t.Fatal(err)
			}
			_, hasTextID := document.Nodes[0]["d-textId"]
			if hasTextID != (tc.wantTextID != "") {
				t.Errorf("unexpected d-textId presence: %v", document.Nodes[0])
			}
			canvas, err = dcanvas.Decode(&encoded)
			if err != nil {
				t.Fatal(err)
			}
			if err := dcanvas.Validate(canvas); err != nil {
				t.Fatal(err)
			}
			got := canvas.Nodes[0]
			if got.Text != tc.wantText || got.TextID != tc.wantTextID || got.Character == nil || got.Character.Name != tc.wantName {
				t.Errorf("node = %+v, character = %+v", got, got.Character)
			}
		})
	}
}

func TestExportDeduplicatesIdenticalCharacters(t *testing.T) {
	base := dialog.Creature{LongName: "Gorris", HasLongNameId: true, LongNameId: 40000, Portrait: "GORRIS.BMP", Sex: "MALE"}
	creature := func(resource string, change func(*dialog.Creature)) *dialog.Creature {
		c := base
		c.ResourceName = resource
		if change != nil {
			change(&c)
		}
		return &c
	}
	for _, tc := range []struct {
		name      string
		creatures []*dialog.Creature
		want      [][4]string
	}{
		{
			"primary and alternatives",
			[]*dialog.Creature{
				creature("H.CRE", func(c *dialog.Creature) { c.LongName = "Other" }),
				creature("G.CRE", func(c *dialog.Creature) { c.LongNameId++ }),
				creature("F.CRE", func(c *dialog.Creature) { c.Sex = "FEMALE" }),
				creature("E.CRE", func(c *dialog.Creature) { c.Portrait = "OTHER.BMP" }),
				creature("D.CRE", func(c *dialog.Creature) { c.LongName = "Baern"; c.LongNameId++ }),
				creature("C.CRE", nil),
				creature("B.CRE", func(c *dialog.Creature) { c.LongName = "Baern"; c.LongNameId++ }),
				creature("A.CRE", nil),
			},
			[][4]string{
				{"Gorris", "#40000", "portraits/GORRIS.png", "MALE"},
				{"Baern", "#40001", "portraits/GORRIS.png", "MALE"},
				{"Gorris", "#40000", "portraits/OTHER.png", "MALE"},
				{"Gorris", "#40000", "portraits/GORRIS.png", "FEMALE"},
				{"Gorris", "#40001", "portraits/GORRIS.png", "MALE"},
				{"Other", "#40000", "portraits/GORRIS.png", "MALE"},
			},
		},
		{
			"duplicates without name references",
			[]*dialog.Creature{
				creature("A.CRE", func(c *dialog.Creature) { c.HasLongNameId = false }),
				creature("B.CRE", func(c *dialog.Creature) { c.HasLongNameId = false }),
			},
			[][4]string{{"Gorris", "", "portraits/GORRIS.png", "MALE"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
			d.RootState = &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, Dialog: d, State: &dialog.StateData{Text: "Hello"}}
			d.AllCreatures["TEST"] = tc.creatures
			canvas, err := d.ToDCanvas(dialog.FormatOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var encoded bytes.Buffer
			if err := dcanvas.Encode(canvas, &encoded); err != nil {
				t.Fatal(err)
			}
			assertSchema(t, encoded.Bytes())
			if len(tc.want) == 1 && bytes.Contains(encoded.Bytes(), []byte("d-alternativeCharacters")) {
				t.Errorf("unexpected alternatives: %s", encoded.Bytes())
			}
			canvas, err = dcanvas.Decode(&encoded)
			if err != nil {
				t.Fatal(err)
			}
			if err := dcanvas.Validate(canvas); err != nil {
				t.Fatal(err)
			}
			line := canvas.Nodes[0]
			characters := append([]*dcanvas.Character{line.Character}, line.AlternativeCharacters...)
			var got [][4]string
			for _, c := range characters {
				got = append(got, [4]string{c.Name, c.TextID, c.Portrait, c.Gender})
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("characters = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExportMissingCreatureNamesUseNodeDialog(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("ROOT", 0))
	d.RootState = &dialog.Node{Type: dialog.StateNodeType, Origin: dialog.NewNodeOrigin("DTEALNIS", 18), Dialog: d, State: &dialog.StateData{Text: "Hello"}}
	d.AllCreatures["DTEALNIS"] = []*dialog.Creature{
		{ResourceName: "ZZZ.CRE", Portrait: "OTHER.BMP"},
		{ResourceName: "MMM.CRE", LongName: "Other", HasLongNameId: true, LongNameId: 7},
		{ResourceName: "AAA.CRE", Portrait: "PORTRAIT.BMP", Sex: "FEMALE"},
	}
	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := dcanvas.Encode(canvas, &encoded); err != nil {
		t.Fatal(err)
	}
	assertSchema(t, encoded.Bytes())
	canvas, err = dcanvas.Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	line := canvas.Nodes[0]
	if ch := line.Character; ch == nil || ch.Name != "DTEALNIS" || ch.TextID != "" || ch.Portrait != "portraits/PORTRAIT.png" || ch.Gender != "FEMALE" {
		t.Errorf("character = %+v", ch)
	}
	if len(line.AlternativeCharacters) != 2 {
		t.Fatalf("alternatives = %+v", line.AlternativeCharacters)
	}
	if ch := line.AlternativeCharacters[0]; ch.Name != "Other" || ch.TextID != "#7" {
		t.Errorf("first alternative = %+v", ch)
	}
	if ch := line.AlternativeCharacters[1]; ch.Name != "DTEALNIS" || ch.TextID != "" || ch.Portrait != "portraits/OTHER.png" {
		t.Errorf("second alternative = %+v", ch)
	}
}
