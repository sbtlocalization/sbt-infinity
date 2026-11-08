// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/text"
	"github.com/spf13/afero"
)

func TestSpeakerWithoutNameReference(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	d.RootState = &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, Dialog: d, State: &dialog.StateData{Text: "Hello"}}
	d.AllCreatures["TEST"] = []*dialog.Creature{{ResourceName: "AAA.CRE", LongName: "Speaker", LongNameId: 99}}
	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := canvas.Nodes[0].Character; got == nil || got.Name != "Speaker" || got.TextID != "" {
		t.Fatalf("character: %+v", got)
	}
	if canvas.Nodes[0].AlternativeCharacters != nil {
		t.Fatalf("unexpected alternatives: %+v", canvas.Nodes[0].AlternativeCharacters)
	}
	var out bytes.Buffer
	if err := dcanvas.Encode(canvas, &out); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), []byte("\"textId\"")) {
		t.Errorf("unexpected character textId: %s", out.Bytes())
	}
}

type reversedDirectoryFS struct{ afero.Fs }

type reversedDirectory struct{ afero.File }

func (f reversedDirectoryFS) Open(name string) (afero.File, error) {
	file, err := f.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	if name == "CRE" {
		return reversedDirectory{file}, nil
	}
	return file, nil
}

func (f reversedDirectory) Readdirnames(n int) ([]string, error) {
	names, err := f.File.Readdirnames(n)
	slices.Reverse(names)
	return names, err
}

func TestExportScannedCreatureSpeakers(t *testing.T) {
	for _, tc := range []struct {
		name       string
		names      []string
		resources  []string
		reverse    bool
		unsetName  bool
		unsetShort bool
	}{
		{"localized", []string{"Zulu", "Alpha", "Short", "Hello", "Middle"}, []string{"ZZZ", "MMM", "AAA"}, false, false, false},
		{"other language", []string{"Alice", "Zebra", "Short", "Hello", "Middle"}, []string{"AAA", "MMM", "ZZZ"}, true, false, false},
		{"unset name reference", []string{"Zulu", "Alpha", "Short", "Hello", "Middle"}, []string{"ZZZ", "MMM", "AAA"}, true, true, false},
		{"empty name", []string{"", "Alpha", "Short", "Hello", "Middle"}, []string{"ZZZ", "MMM", "AAA"}, true, false, false},
		{"no name variant", []string{"", "Alpha", "Short", "Hello", "Middle"}, []string{"ZZZ", "MMM", "AAA"}, true, false, true},
		{"single", []string{"Zulu", "Alpha", "Short", "Hello", "Middle"}, []string{"AAA"}, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if err := fs.Mkdir("CRE", 0755); err != nil {
				t.Fatal(err)
			}
			if err := afero.WriteFile(fs, "GENDER.IDS", []byte("IDS V1.0\n1 FEMALE\n"), 0644); err != nil {
				t.Fatal(err)
			}
			entries := make([]text.TlkWriteEntry, len(tc.names))
			for i, s := range tc.names {
				entries[i] = text.TlkWriteEntry{Text: s, HasText: true}
			}
			var tlk bytes.Buffer
			if err := text.WriteTlk(&tlk, entries, text.TlkWriteOptions{}); err != nil {
				t.Fatal(err)
			}
			if err := afero.WriteFile(fs, "dialog.tlk", tlk.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			for _, resource := range tc.resources {
				ref, portrait := uint32(0), "AAA_P"
				if resource == "ZZZ" {
					ref, portrait = 1, "ZZZ_P"
					if tc.unsetName {
						ref = ^uint32(0)
					}
				} else if resource == "MMM" {
					ref, portrait = 4, "MMM_P"
				}
				cre := make([]byte, 0x2d4)
				copy(cre, "CRE V1.0")
				binary.LittleEndian.PutUint32(cre[8:12], ref)
				shortRef := uint32(2)
				if tc.unsetShort && resource == "AAA" {
					shortRef = ^uint32(0)
				}
				binary.LittleEndian.PutUint32(cre[12:16], shortRef)
				copy(cre[0x34:0x3c], portrait)
				cre[0x237] = 1
				copy(cre[0x2cc:0x2d4], "TEST")
				if err := afero.WriteFile(fs, resource+".CRE", cre, 0644); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, "CRE/"+resource+".CRE", cre, 0644); err != nil {
					t.Fatal(err)
				}
			}
			dlg := make([]byte, 72+6)
			copy(dlg, "DLG V1.0")
			put := func(off int, n uint32) { binary.LittleEndian.PutUint32(dlg[off:off+4], n) }
			put(8, 1)
			put(12, 52)
			put(24, 68)
			put(28, 1)
			put(52, 3)
			put(64, 0)
			put(68, 76)
			put(72, 2)
			copy(dlg[76:], "OK")
			if err := afero.WriteFile(fs, "TEST.DLG", dlg, 0644); err != nil {
				t.Fatal(err)
			}
			var resourceFS afero.Fs = fs
			if tc.reverse {
				resourceFS = reversedDirectoryFS{fs}
			}
			builder := dialog.NewDialogBuilder(resourceFS, fs, true, false)
			defer builder.Close()
			collection, err := builder.LoadAllDialogs("dialog.tlk", "TEST.DLG")
			if err != nil {
				t.Fatal(err)
			}
			if len(collection.Dialogs) != 1 {
				t.Fatalf("dialogs: %d", len(collection.Dialogs))
			}
			canvas, err := collection.Dialogs[0].ToDCanvas(dialog.FormatOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := dcanvas.Encode(canvas, &out); err != nil {
				t.Fatal(err)
			}
			assertSchema(t, out.Bytes())
			canvas, err = dcanvas.Decode(&out)
			if err != nil {
				t.Fatal(err)
			}
			if err := dcanvas.Validate(canvas); err != nil {
				t.Fatal(err)
			}
			ch := canvas.Nodes[0].Character
			name, textID := tc.names[0], "#0"
			if name == "" {
				name, textID = tc.names[2], "#2"
				if tc.unsetShort {
					name, textID = "AAA", ""
				}
			}
			if ch == nil || ch.Name != name || ch.TextID != textID || ch.Portrait != "portraits/AAA_P.png" || ch.Gender != "FEMALE" {
				t.Fatalf("displayed character: %+v", ch)
			}
			alternatives := canvas.Nodes[0].AlternativeCharacters
			if len(alternatives) != len(tc.resources)-1 {
				t.Fatalf("alternatives: %+v", alternatives)
			}
			if len(alternatives) > 0 {
				if a := alternatives[0]; a.Name != tc.names[4] || a.TextID != "#4" || a.Portrait != "portraits/MMM_P.png" || a.Gender != "FEMALE" {
					t.Errorf("first alternative: %+v", a)
				}
				name, textID := tc.names[1], "#1"
				if tc.unsetName {
					name, textID = tc.names[2], "#2"
				}
				if a := alternatives[1]; a.Name != name || a.TextID != textID || a.Portrait != "portraits/ZZZ_P.png" || a.Gender != "FEMALE" {
					t.Errorf("second alternative: %+v", a)
				}
			}
		})
	}
}
