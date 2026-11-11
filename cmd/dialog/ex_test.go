// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sbtlocalization/dcanvas"
	cmdDialog "github.com/sbtlocalization/sbt-infinity/cmd/dialog"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/text"
	"golang.org/x/image/bmp"
)

type resource struct {
	name string
	kind uint16
	data []byte
}

func TestExportCommandWithCreaturePortraits(t *testing.T) {
	dir := t.TempDir()
	var tlk bytes.Buffer
	entries := []text.TlkWriteEntry{
		{Text: "Zulu", HasText: true}, {Text: "Alpha", HasText: true},
		{Text: "Hello", HasText: true, HasSound: true, AudioName: "VOICE"}, {Text: "Bye", HasText: true},
	}
	if err := text.WriteTlk(&tlk, entries, text.TlkWriteOptions{}); err != nil {
		t.Fatal(err)
	}
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("dialog.tlk", tlk.Bytes())
	dlg := make([]byte, 52+16+32+8+6)
	copy(dlg, "DLG V1.0")
	put32 := func(buf []byte, off int, v uint32) { binary.LittleEndian.PutUint32(buf[off:off+4], v) }
	put32(dlg, 8, 1)
	put32(dlg, 12, 52)
	put32(dlg, 16, 1)
	put32(dlg, 20, 68)
	put32(dlg, 24, 100)
	put32(dlg, 28, 1)
	put32(dlg, 52, 2)
	put32(dlg, 60, 1)
	put32(dlg, 64, 0)
	put32(dlg, 68, 9)
	put32(dlg, 72, 3)
	put32(dlg, 100, 108)
	put32(dlg, 104, 6)
	copy(dlg[108:], "True()")
	cre := func(ref uint32, portrait string) []byte {
		b := make([]byte, 0x2d4)
		copy(b, "CRE V1.0")
		put32(b, 8, ref)
		put32(b, 12, ref)
		copy(b[0x34:], portrait)
		copy(b[0x2cc:], "TEST")
		b[0x237] = 1
		return b
	}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var pic bytes.Buffer
	if err := bmp.Encode(&pic, img); err != nil {
		t.Fatal(err)
	}
	resources := []resource{
		{"TEST", 1011, dlg}, {"ZZZ", 1009, cre(1, "ZZZ_P")}, {"AAA", 1009, cre(0, "AAA_P")},
		{"AAA_P", 1, pic.Bytes()}, {"ZZZ_P", 1, pic.Bytes()}, {"GENDER", 1008, []byte("IDS V1.0\n1 FEMALE\n")},
	}
	key, bif := makeArchive(resources)
	write("chitin.key", key)
	write("test.bif", bif)
	out := filepath.Join(dir, "out")
	cmd := cmdDialog.NewCommand()
	cmd.PersistentFlags().String("key", filepath.Join(dir, "chitin.key"), "")
	cmd.SetArgs([]string{"export", "--tlk", filepath.Join(dir, "dialog.tlk"), "--output", out, "--report", filepath.Join(dir, "report.json"), "--scan-overlaps", "TEST.DLG"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "TEST-0.d.canvas"))
	if err != nil {
		t.Fatal(err)
	}
	canvas, err := dcanvas.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	if canvas.Version != "1.1" || len(canvas.Nodes) != 2 || len(canvas.Edges) != 1 {
		t.Fatalf("canvas: %+v", canvas)
	}
	line, end := canvas.Nodes[0], canvas.Nodes[1]
	if line.ID != "state-TEST[0]" || line.Text != "Hello" || line.Sound != "sounds/VOICE.wav" || line.Character == nil || line.Character.Name != "Zulu" || line.Character.TextID != "#0" || line.Character.Gender != "FEMALE" {
		t.Errorf("line: %+v", line)
	}
	if len(line.AlternativeCharacters) != 1 || line.AlternativeCharacters[0].Name != "Alpha" || line.AlternativeCharacters[0].TextID != "#1" {
		t.Errorf("alternatives: %+v", line.AlternativeCharacters)
	}
	if end.Kind != dcanvas.KindReply || end.Role != "dialog-end" || end.Text != "Bye" || end.Character == nil || end.Character.Name != "End dialog" {
		t.Errorf("end: %+v", end)
	}
	for _, portrait := range []string{"AAA_P.png", "ZZZ_P.png"} {
		if _, err := os.Stat(filepath.Join(out, "portraits", portrait)); err != nil {
			t.Error(err)
		}
	}
	var report struct {
		Sounds  []string `json:"sounds"`
		Strrefs []uint32 `json:"strrefs"`
	}
	reportBytes, err := os.ReadFile(filepath.Join(dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Sounds) != 1 || report.Sounds[0] != "VOICE" || len(report.Strrefs) != 2 || report.Strrefs[0] != 2 || report.Strrefs[1] != 3 {
		t.Errorf("report: %+v", report)
	}
	schema, err := jsonschema.Compile("../../dialog/testdata/dCanvas-1.1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(document); err != nil {
		t.Fatal(err)
	}
}

func TestExportCommandEmptyTextAndMissingText(t *testing.T) {
	dir := t.TempDir()
	var tlk bytes.Buffer
	if err := text.WriteTlk(&tlk, []text.TlkWriteEntry{{Text: "", HasText: true}}, text.TlkWriteOptions{}); err != nil {
		t.Fatal(err)
	}
	dlg := make([]byte, 146)
	copy(dlg, "DLG V1.0")
	put := func(off int, n uint32) { binary.LittleEndian.PutUint32(dlg[off:off+4], n) }
	put(8, 1)
	put(12, 52)
	put(16, 2)
	put(20, 68)
	put(24, 132)
	put(28, 1)
	put(52, 0)
	put(60, 2)
	put(64, 0)
	put(68, 8)
	put(100, 9)
	put(104, 0)
	put(132, 140)
	put(136, 6)
	copy(dlg[140:], "True()")
	key, bif := makeArchive([]resource{{"TEST", 1011, dlg}})
	for name, data := range map[string][]byte{"dialog.tlk": tlk.Bytes(), "chitin.key": key, "test.bif": bif} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(dir, "out")
	cmd := cmdDialog.NewCommand()
	cmd.PersistentFlags().String("key", filepath.Join(dir, "chitin.key"), "")
	cmd.SetArgs([]string{"export", "--speakers=false", "--tlk", filepath.Join(dir, "dialog.tlk"), "--output", out, "TEST.DLG"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "TEST-0.d.canvas"))
	if err != nil {
		t.Fatal(err)
	}
	canvas, err := dcanvas.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	if len(canvas.Nodes) != 3 {
		t.Fatalf("nodes: %+v", canvas.Nodes)
	}
	line, missing, empty := canvas.Nodes[0], canvas.Nodes[1], canvas.Nodes[2]
	if line.Text != "" || line.TextID != "#0" || line.Character == nil || line.Character.Name != "TEST" {
		t.Errorf("line: %+v", line)
	}
	if missing.Text != "(no text)" || missing.TextID != "" || missing.Character == nil || missing.Character.Name != "End dialog" {
		t.Errorf("missing: %+v", missing)
	}
	if empty.Text != "" || empty.TextID != "#0" || empty.Character == nil || empty.Character.Name != "End dialog" {
		t.Errorf("empty: %+v", empty)
	}
}

func TestExportLegacyBytesInTextAndJournal(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	var tlk bytes.Buffer
	entries := []text.TlkWriteEntry{
		{Text: "caf\xe9", HasText: true},
		{Text: "r\xe9ponse", HasText: true},
		{Text: "journ\xe9e", HasText: true},
	}
	if err := text.WriteTlk(&tlk, entries, text.TlkWriteOptions{}); err != nil {
		t.Fatal(err)
	}
	write("dialog.tlk", tlk.Bytes())
	dlg := make([]byte, 52+16+32+8+6)
	copy(dlg, "DLG V1.0")
	put := func(off int, n uint32) { binary.LittleEndian.PutUint32(dlg[off:off+4], n) }
	put(8, 1)
	put(12, 52)
	put(16, 1)
	put(20, 68)
	put(24, 100)
	put(28, 1)
	put(52, 0)
	put(60, 1)
	put(64, 0)
	put(68, 0x19)
	put(72, 1)
	put(76, 2)
	put(100, 108)
	put(104, 6)
	copy(dlg[108:], "True()")
	key, bif := makeArchive([]resource{{"TEST", 1011, dlg}})
	write("chitin.key", key)
	write("test.bif", bif)
	out := filepath.Join(dir, "out")
	if err := os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "TEST-0.d.canvas"), []byte("previous export"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := cmdDialog.NewCommand()
	cmd.PersistentFlags().String("key", filepath.Join(dir, "chitin.key"), "")
	cmd.SetArgs([]string{"export", "--speakers=false", "--tlk", filepath.Join(dir, "dialog.tlk"), "--output", out, "--report", filepath.Join(dir, "report.json"), "TEST.DLG"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "TEST-0.d.canvas"))
	if err != nil {
		t.Fatal(err)
	}
	canvas, err := dcanvas.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	if canvas.Nodes[0].Text != "caf\ufffd" || canvas.Nodes[1].Text != "r\ufffdponse" {
		t.Errorf("text: %+v %+v", canvas.Nodes[0], canvas.Nodes[1])
	}
	var document struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Nodes[1]["x-journalText"] != "journ\ufffde" {
		t.Errorf("journal: %v", document.Nodes[1])
	}
	if _, err := os.Stat(filepath.Join(dir, "report.json")); err != nil {
		t.Fatal(err)
	}

	t.Run("encoding failure preserves existing file", func(t *testing.T) {
		fileName := filepath.Join(out, "TEST-0.d.canvas")
		original := []byte("previous export")
		if err := os.WriteFile(fileName, original, 0644); err != nil {
			t.Fatal(err)
		}
		previousColor := dialog.FinalTransitionColor
		dialog.FinalTransitionColor = string([]byte{0xff})
		t.Cleanup(func() { dialog.FinalTransitionColor = previousColor })
		cmd := cmdDialog.NewCommand()
		cmd.PersistentFlags().String("key", filepath.Join(dir, "chitin.key"), "")
		cmd.SetArgs([]string{"export", "--speakers=false", "--tlk", filepath.Join(dir, "dialog.tlk"), "--output", out, "TEST.DLG"})
		if err := cmd.Execute(); err == nil {
			t.Fatal("expected encoding failure")
		}
		got, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, original) {
			t.Fatalf("existing file was modified: %q", got)
		}
	})
}

func makeArchive(resources []resource) ([]byte, []byte) {
	bif := make([]byte, 20+16*len(resources))
	copy(bif, "BIFFV1  ")
	binary.LittleEndian.PutUint32(bif[8:], uint32(len(resources)))
	binary.LittleEndian.PutUint32(bif[16:], 20)
	key := make([]byte, 36+14*len(resources)+9)
	copy(key, "KEY V1  ")
	binary.LittleEndian.PutUint32(key[8:], 1)
	binary.LittleEndian.PutUint32(key[12:], uint32(len(resources)))
	binary.LittleEndian.PutUint32(key[16:], 24)
	binary.LittleEndian.PutUint32(key[20:], 36)
	binary.LittleEndian.PutUint32(key[28:], uint32(36+14*len(resources)))
	binary.LittleEndian.PutUint16(key[32:], 9)
	copy(key[36+14*len(resources):], "test.bif\x00")
	for i, r := range resources {
		o := 20 + i*16
		binary.LittleEndian.PutUint32(bif[o:], uint32(i))
		binary.LittleEndian.PutUint32(bif[o+4:], uint32(len(bif)))
		binary.LittleEndian.PutUint32(bif[o+8:], uint32(len(r.data)))
		binary.LittleEndian.PutUint16(bif[o+12:], r.kind)
		bif = append(bif, r.data...)
		k := 36 + i*14
		copy(key[k:k+8], r.name)
		binary.LittleEndian.PutUint16(key[k+8:], r.kind)
		binary.LittleEndian.PutUint32(key[k+10:], uint32(i))
	}
	binary.LittleEndian.PutUint32(key[24:], uint32(len(bif)))
	return key, bif
}
