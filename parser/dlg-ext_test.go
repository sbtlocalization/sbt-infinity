// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"bytes"
	"testing"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

const noTrigger = uint32(0xFFFFFFFF)

func newDlgStream(t *testing.T, buf []byte) (*Dlg, *kaitai.Stream) {
	t.Helper()
	stream := kaitai.NewStream(bytes.NewReader(buf))
	dlg := &Dlg{_io: stream}
	dlg._root = dlg
	return dlg, stream
}

// A text entry followed inline by its bytes: OfsText points just past the two
// u4 fields, so a single-entry buffer is self-contained.
func textEntryBuffer(text string) []byte {
	var buf bytes.Buffer
	writeU4le(&buf, 8) // OfsText
	writeU4le(&buf, uint32(len(text)))
	buf.WriteString(text)
	return buf.Bytes()
}

func writeU4le(buf *bytes.Buffer, v uint32) {
	buf.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func TestTransitionGetTriggerText_ReadFailureReturnsError(t *testing.T) {
	dlg, stream := newDlgStream(t, []byte{0, 0, 0, 0}) // too small for a text entry
	tr := &Dlg_TransitionEntry{
		_io:                    stream,
		_root:                  dlg,
		Flags:                  &Dlg_TransitionEntry_Flags{WithTrigger: true},
		TransitionTriggerIndex: 1000, // seeks past end → read fails
	}

	text, ok, err := tr.GetTriggerText()
	if err == nil {
		t.Fatal("GetTriggerText() err = nil, want a read failure")
	}
	if ok || text != "" {
		t.Errorf("GetTriggerText() = (%q, %v), want (\"\", false)", text, ok)
	}
}

func TestTransitionGetActionText_ReadFailureReturnsError(t *testing.T) {
	dlg, stream := newDlgStream(t, []byte{0, 0, 0, 0})
	tr := &Dlg_TransitionEntry{
		_io:                   stream,
		_root:                 dlg,
		Flags:                 &Dlg_TransitionEntry_Flags{WithAction: true},
		TransitionActionIndex: 1000,
	}

	text, ok, err := tr.GetActionText()
	if err == nil {
		t.Fatal("GetActionText() err = nil, want a read failure")
	}
	if ok || text != "" {
		t.Errorf("GetActionText() = (%q, %v), want (\"\", false)", text, ok)
	}
}

func TestTransitionGetTriggerText_NoFlagIsSilent(t *testing.T) {
	dlg, stream := newDlgStream(t, []byte{0, 0, 0, 0})
	tr := &Dlg_TransitionEntry{
		_io:   stream,
		_root: dlg,
		Flags: &Dlg_TransitionEntry_Flags{WithTrigger: false},
	}

	text, ok, err := tr.GetTriggerText()
	if err != nil {
		t.Fatalf("GetTriggerText() err = %v, want nil for an absent trigger", err)
	}
	if ok || text != "" {
		t.Errorf("GetTriggerText() = (%q, %v), want (\"\", false)", text, ok)
	}
}

func TestTransitionGetTriggerText_WellFormed(t *testing.T) {
	dlg, stream := newDlgStream(t, textEntryBuffer("hello"))
	tr := &Dlg_TransitionEntry{
		_io:                    stream,
		_root:                  dlg,
		Flags:                  &Dlg_TransitionEntry_Flags{WithTrigger: true},
		TransitionTriggerIndex: 0,
	}

	text, ok, err := tr.GetTriggerText()
	if err != nil {
		t.Fatalf("GetTriggerText() err = %v, want nil", err)
	}
	if !ok || text != "hello" {
		t.Errorf("GetTriggerText() = (%q, %v), want (\"hello\", true)", text, ok)
	}
}

func TestStateGetTriggerText_ReadFailureReturnsError(t *testing.T) {
	dlg, stream := newDlgStream(t, []byte{0, 0, 0, 0})
	st := &Dlg_StateEntry{
		_io:               stream,
		_root:             dlg,
		StateTriggerIndex: 1000, // present but unreadable
	}

	text, ok, err := st.GetTriggerText()
	if err == nil {
		t.Fatal("GetTriggerText() err = nil, want a read failure")
	}
	if ok || text != "" {
		t.Errorf("GetTriggerText() = (%q, %v), want (\"\", false)", text, ok)
	}
}

// A state that genuinely carries no trigger (index 0xFFFFFFFF) is a root state:
// text-absent with no error, so callers can classify it without a warning.
func TestStateGetTriggerText_NoTriggerIsSilent(t *testing.T) {
	dlg, stream := newDlgStream(t, []byte{0, 0, 0, 0})
	st := &Dlg_StateEntry{
		_io:               stream,
		_root:             dlg,
		StateTriggerIndex: noTrigger,
	}

	text, ok, err := st.GetTriggerText()
	if err != nil {
		t.Fatalf("GetTriggerText() err = %v, want nil for a root state", err)
	}
	if ok || text != "" {
		t.Errorf("GetTriggerText() = (%q, %v), want (\"\", false)", text, ok)
	}
}
