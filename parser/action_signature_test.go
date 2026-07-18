// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import "testing"

func TestParseActionSignature_Basic(t *testing.T) {
	sig, err := ParseActionSignature("DisplayString(O:Object*,I:StrRef*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}

	if sig.Name != "DisplayString" {
		t.Errorf("Name = %q, want %q", sig.Name, "DisplayString")
	}

	want := []ActionParameter{
		{Type: "O", Name: "Object"},
		{Type: "I", Name: "StrRef"},
	}
	if len(sig.Parameters) != len(want) {
		t.Fatalf("len(Parameters) = %d, want %d", len(sig.Parameters), len(want))
	}
	for i, p := range want {
		if sig.Parameters[i] != p {
			t.Errorf("Parameters[%d] = %+v, want %+v", i, sig.Parameters[i], p)
		}
	}

	if !sig.HasStrRef() {
		t.Error("HasStrRef() = false, want true")
	}
	if sig.StrRefPositional != 1 {
		t.Errorf("StrRefPositional = %d, want 1", sig.StrRefPositional)
	}
	if sig.StrRefIntSlot != 0 {
		t.Errorf("StrRefIntSlot = %d, want 0", sig.StrRefIntSlot)
	}
}

func TestParseActionSignature_NoParameters(t *testing.T) {
	sig, err := ParseActionSignature("Wait()")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}

	if sig.Name != "Wait" {
		t.Errorf("Name = %q, want %q", sig.Name, "Wait")
	}
	if len(sig.Parameters) != 0 {
		t.Errorf("len(Parameters) = %d, want 0", len(sig.Parameters))
	}
	if sig.HasStrRef() {
		t.Error("HasStrRef() = true, want false")
	}
}

func TestParseActionSignature_NoStrRef(t *testing.T) {
	// Trigger-style signature: parsed for completeness, never carries a StrRef.
	sig, err := ParseActionSignature("Global(S:Name*,S:Area*,I:Value*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}

	if sig.HasStrRef() {
		t.Error("HasStrRef() = true, want false")
	}
	if sig.StrRefPositional != -1 {
		t.Errorf("StrRefPositional = %d, want -1", sig.StrRefPositional)
	}
	if sig.StrRefIntSlot != -1 {
		t.Errorf("StrRefIntSlot = %d, want -1", sig.StrRefIntSlot)
	}
}

func TestParseActionSignature_StrRefNameCaseInsensitive(t *testing.T) {
	for _, name := range []string{"StrRef", "STRREF", "Strref", "strref"} {
		sig, err := ParseActionSignature("DisplayString(O:Object*,I:" + name + "*)")
		if err != nil {
			t.Fatalf("ParseActionSignature(%q) failed: %v", name, err)
		}
		if !sig.HasStrRef() {
			t.Errorf("HasStrRef() = false for parameter name %q, want true", name)
		}
	}
}

func TestParseActionSignature_IntSlotSkipsNonIntegerParameters(t *testing.T) {
	// Intervening object/point parameters must not advance the integer-slot count.
	sig, err := ParseActionSignature("DisplayStringPoint(O:Object*,P:Point*,I:StrRef*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}

	if sig.StrRefPositional != 2 {
		t.Errorf("StrRefPositional = %d, want 2", sig.StrRefPositional)
	}
	if sig.StrRefIntSlot != 0 {
		t.Errorf("StrRefIntSlot = %d, want 0", sig.StrRefIntSlot)
	}
}

func TestParseActionSignature_TormentReordering(t *testing.T) {
	// Planescape: Torment reorders parameters relative to BG2/IWD. Placing
	// StrRef before vs. after an intervening P:Location* parameter, with
	// another integer parameter on the far side, must yield different
	// integer-slot indices — proving the slot is counted by type, not by
	// raw position.
	before, err := ParseActionSignature("DisplayStringPointLog(I:StrRef*,P:Location*,I:Other*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}
	after, err := ParseActionSignature("DisplayStringPointLog(I:Other*,P:Location*,I:StrRef*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}

	if before.StrRefIntSlot != 0 {
		t.Errorf("before.StrRefIntSlot = %d, want 0", before.StrRefIntSlot)
	}
	if after.StrRefIntSlot != 1 {
		t.Errorf("after.StrRefIntSlot = %d, want 1", after.StrRefIntSlot)
	}
	if before.StrRefIntSlot == after.StrRefIntSlot {
		t.Error("expected different integer-slot indices for the two orderings")
	}
}

func TestParseActionSignature_HasObject(t *testing.T) {
	cases := []struct {
		identifier string
		want       bool
	}{
		{"DisplayString(O:Object*,I:StrRef*)", true},
		{"EraseJournalEntry(I:STRREF*)", false},
		{"Wait()", false},
		{"DisplayStringPointLog(I:StrRef*,P:Location*,I:Other*)", false},
	}
	for _, c := range cases {
		sig, err := ParseActionSignature(c.identifier)
		if err != nil {
			t.Fatalf("ParseActionSignature(%q) failed: %v", c.identifier, err)
		}
		if sig.HasObject() != c.want {
			t.Errorf("HasObject() = %v for %q, want %v", sig.HasObject(), c.identifier, c.want)
		}
	}
}

func TestParseActionSignature_InvalidIdentifier(t *testing.T) {
	if _, err := ParseActionSignature("DisplayString"); err == nil {
		t.Error("expected error for missing parameter list, got nil")
	}
	if _, err := ParseActionSignature("DisplayString(Object)"); err == nil {
		t.Error("expected error for parameter missing a type, got nil")
	}
}

func TestBuildDisplayFamilyActions(t *testing.T) {
	ids := &Ids{
		Entries: map[int32]string{
			0: "Wait(I:Time*)",
			1: "DisplayString(O:Object*,I:StrRef*)",
			2: "DisplayStringHead(O:Object*,I:StrRef*)",
			3: "ActionOverride(O:Object*,A:Action*)",
		},
	}

	family, err := BuildDisplayFamilyActions(ids)
	if err != nil {
		t.Fatalf("BuildDisplayFamilyActions failed: %v", err)
	}

	if len(family) != 2 {
		t.Fatalf("len(family) = %d, want 2", len(family))
	}

	for _, name := range []string{"DisplayString", "DisplayStringHead"} {
		if _, ok := family[name]; !ok {
			t.Errorf("family missing %q", name)
		}
	}
	for _, name := range []string{"Wait", "ActionOverride"} {
		if _, ok := family[name]; ok {
			t.Errorf("family unexpectedly contains %q (no StrRef parameter)", name)
		}
	}
}
