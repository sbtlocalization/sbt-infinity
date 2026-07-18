// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"strconv"
	"strings"
	"testing"
)

// Action ids used across the BCS tests. Mirror the kind of numeric ids a real
// ACTION.IDS assigns; the exact values are arbitrary as long as they match the
// glued markers in the script source.
const (
	actDisplayString     = 100 // DisplayString(O:Object*,I:StrRef*)
	actDisplayStringHead = 138 // DisplayStringHead(O:Object*,I:StrRef*)
	actAttack            = 3   // Attack(O:Target*) — no StrRef
	actSetGlobal         = 30  // SetGlobal(S:Name*,S:Area*,I:Value*) — no StrRef
	// A signature whose StrRef lands in a non-zero integer slot (int1):
	// the first integer parameter is not the StrRef, so StrRefIntSlot == 1
	// maps to physical field 3.
	actDisplayStringWait = 200 // DisplayStringWait(O:Object*,I:Time*,I:StrRef*)
	// An objectless display-family signature (IWD shape): the compiled block
	// still carries three all-zero object slots, but the rendered call must
	// not show them.
	actEraseJournalEntry = 253 // EraseJournalEntry(I:STRREF*)
)

// OBJECT.IDS values used in the tests.
const (
	objLastTrigger = 14
	objMyself      = 1
)

func testActionIds() *Ids {
	return &Ids{
		Entries: map[int32]string{
			actDisplayString:     "DisplayString(O:Object*,I:StrRef*)",
			actDisplayStringHead: "DisplayStringHead(O:Object*,I:StrRef*)",
			actAttack:            "Attack(O:Target*)",
			actSetGlobal:         "SetGlobal(S:Name*,S:Area*,I:Value*)",
			actDisplayStringWait: "DisplayStringWait(O:Object*,I:Time*,I:StrRef*)",
			actEraseJournalEntry: "EraseJournalEntry(I:STRREF*)",
		},
	}
}

func testObjectIds() *Ids {
	return &Ids{
		Entries: map[int32]string{
			objLastTrigger: "LastTrigger",
			objMyself:      "Myself",
		},
	}
}

func testDisplayFamilyActions(t *testing.T) DisplayFamilyActions {
	t.Helper()
	family, err := BuildDisplayFamilyActions(testActionIds())
	if err != nil {
		t.Fatalf("BuildDisplayFamilyActions failed: %v", err)
	}
	return family
}

// obBlock renders an OB object block. ident is an OBJECT.IDS value; name is the
// trailing object name string (empty for none). All enumerated EA/RACE/… fields
// are left at zero, i.e. a plain identifier reference.
func obBlock(ident int, name string) string {
	return strings.Join([]string{
		"OB",
		"0 0 0 0 0 0 0", // ea general race class specific gender alignment
		intStr(ident),   // identifiers (OBJECT.IDS)
		"0 0 0 0",       // four trailing unknown integers
		"[0.0]",         // point
		quote(name),     // name string
		"OB",
	}, "\n")
}

// Real scripts keep the action id as its own field — an integer glued to
// the AC marker (e.g. "100AC") is the enclosing response's weight, not the
// action id.
func acBlock(id int, o1, o2, o3 string, int0, x, y, int1, int2 int) string {
	return strings.Join([]string{
		"AC",
		intStr(id),
		o1,
		o2,
		o3,
		// int0 x y int1 int2, then two empty strings
		strings.Join([]string{intStr(int0), intStr(x), intStr(y), intStr(int1), intStr(int2)}, " ") + ` "" ""`,
		"AC",
	}, "\n")
}

func emptyObject() string { return obBlock(0, "") }

func intStr(v int) string { return strconv.Itoa(v) }

func quote(s string) string { return `"` + s + `"` }

// wrapScript wraps action blocks in the surrounding SC/CR/RS/RE structure of a
// compiled script so the parser sees a realistic block layout.
func wrapScript(actions ...string) string {
	var sb strings.Builder
	sb.WriteString("SC\nCR\nCO\nCO\nRS\n")
	for _, a := range actions {
		sb.WriteString("RE\n0\n")
		sb.WriteString(a)
		sb.WriteString("\nRE\n")
	}
	sb.WriteString("RS\nCR\nSC\n")
	return sb.String()
}

func decode(t *testing.T, script string) []ScriptCall {
	t.Helper()
	calls, err := DecodeCompiledScript(
		strings.NewReader(script),
		testActionIds(),
		testDisplayFamilyActions(t),
		testObjectIds(),
	)
	if err != nil {
		t.Fatalf("DecodeCompiledScript failed: %v", err)
	}
	return calls
}

// TestDecodeCompiledScript_CTFB is the end-to-end anchor: a CTFB-shaped script
// decodes to {14381, "DisplayString(LastTrigger,14381)"}, mirroring the real
// CTFB.BCS (BG1, "Trap Sprung"). The StrRef sits in int0 and the object lives
// in the o2 slot.
func TestDecodeCompiledScript_CTFB(t *testing.T) {
	action := acBlock(
		actDisplayString,
		emptyObject(),               // o1: ActionOverride target (empty)
		obBlock(objLastTrigger, ""), // o2: the normal object → LastTrigger
		emptyObject(),               // o3
		14381, 0, 0, 0, 0,           // int0=StrRef, point, int1, int2
	)
	script := wrapScript(action)

	calls := decode(t, script)
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].StrRef != 14381 {
		t.Errorf("StrRef = %d, want 14381", calls[0].StrRef)
	}
	if calls[0].Call != "DisplayString(LastTrigger,14381)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "DisplayString(LastTrigger,14381)")
	}
}

// TestDecodeCompiledScript_NonDisplayFamily verifies actions without a StrRef
// parameter contribute no pairs.
func TestDecodeCompiledScript_NonDisplayFamily(t *testing.T) {
	script := wrapScript(
		acBlock(actAttack, emptyObject(), obBlock(objMyself, ""), emptyObject(), 0, 0, 0, 0, 0),
		acBlock(actSetGlobal, emptyObject(), emptyObject(), emptyObject(), 5, 0, 0, 0, 0),
	)

	calls := decode(t, script)
	if len(calls) != 0 {
		t.Fatalf("len(calls) = %d, want 0 (no display-family actions)", len(calls))
	}
}

// TestDecodeCompiledScript_IntSlotNonZero exercises the integer-slot mapping by
// placing the StrRef in int1 (logical slot 1 → physical field 3), proving the
// decoder reads the correct field rather than always int0.
func TestDecodeCompiledScript_IntSlotNonZero(t *testing.T) {
	sig, err := ParseActionSignature("DisplayStringWait(O:Object*,I:Time*,I:StrRef*)")
	if err != nil {
		t.Fatalf("ParseActionSignature failed: %v", err)
	}
	if sig.StrRefIntSlot != 1 {
		t.Fatalf("StrRefIntSlot = %d, want 1 (test premise)", sig.StrRefIntSlot)
	}

	// int0=Time=7 (a decoy in physical field 0), int1=StrRef=14381 (physical
	// field 3). A decoder that read field 0 would wrongly return 7.
	action := acBlock(
		actDisplayStringWait,
		emptyObject(),
		obBlock(objMyself, ""),
		emptyObject(),
		7, 0, 0, 14381, 0,
	)

	calls := decode(t, wrapScript(action))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].StrRef != 14381 {
		t.Errorf("StrRef = %d, want 14381 (read from int1/physical field 3)", calls[0].StrRef)
	}
	if calls[0].Call != "DisplayStringWait(Myself,14381)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "DisplayStringWait(Myself,14381)")
	}
}

// TestDecodeCompiledScript_ObjectConvention verifies the o1=ActionOverride /
// o2=normal convention: even with a distinct object in o1, the rendered call
// uses o2, and both LastTrigger and Myself resolve from OBJECT.IDS by value.
func TestDecodeCompiledScript_ObjectConvention(t *testing.T) {
	action := acBlock(
		actDisplayString,
		obBlock(objLastTrigger, ""), // o1: ActionOverride target — must be ignored
		obBlock(objMyself, ""),      // o2: the rendered object
		emptyObject(),
		77, 0, 0, 0, 0,
	)

	calls := decode(t, wrapScript(action))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].Call != "DisplayString(Myself,77)" {
		t.Errorf("Call = %q, want %q (o2 object, not o1)", calls[0].Call, "DisplayString(Myself,77)")
	}
}

// TestDecodeCompiledScript_ObjectName verifies an object name string is shown
// in preference to an OBJECT.IDS identifier.
func TestDecodeCompiledScript_ObjectName(t *testing.T) {
	action := acBlock(
		actDisplayString,
		emptyObject(),
		obBlock(0, "Gorion"), // o2: named object
		emptyObject(),
		42, 0, 0, 0, 0,
	)

	calls := decode(t, wrapScript(action))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].Call != "DisplayString(Gorion,42)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "DisplayString(Gorion,42)")
	}
}

// TestDecodeCompiledScript_ObjectlessSignature is the parser-seam test for the
// phantom-object bug: an all-zero o2 slot resolves to "Nothing" via OBJECT.IDS,
// so a signature without an object parameter must suppress the slot entirely —
// EraseJournalEntry(26331), not EraseJournalEntry(Nothing,26331).
func TestDecodeCompiledScript_ObjectlessSignature(t *testing.T) {
	action := acBlock(
		actEraseJournalEntry,
		emptyObject(),
		emptyObject(),
		emptyObject(),
		26331, 0, 0, 0, 0,
	)

	calls := decode(t, wrapScript(action))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].StrRef != 26331 {
		t.Errorf("StrRef = %d, want 26331", calls[0].StrRef)
	}
	if calls[0].Call != "EraseJournalEntry(26331)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "EraseJournalEntry(26331)")
	}
}

// Mirrors real compiled scripts: a response weight glued to the opening AC
// marker ("100AC"), the action id glued to the first object's OB marker
// ("151OB", as in BG1's CTFB.BCS), and an object identifier glued to a
// trailing OB marker.
func TestDecodeCompiledScript_GluedMarkers(t *testing.T) {
	// o2 with the identifier glued to its closing OB marker.
	o2 := strings.Join([]string{
		"OB",
		"0 0 0 0 0 0 0",
		"14",      // identifiers = LastTrigger
		"0 0 0 0", // unknowns
		"[0.0]",
		`""`,
		"OB",
	}, "\n")
	// Force a glued "151OB" sequence in the o3 slot to exercise the splitter.
	o3 := "OB\n0 0 0 0 0 0 0\n0\n0 0 0 151OB"

	action := strings.Join([]string{
		"100AC",                         // glued response weight + opening AC marker
		intStr(actDisplayString) + "OB", // glued action id + first object's opening OB marker
		"0 0 0 0 0 0 0 0 0 0 0 0",
		`""`,
		"OB",
		o2,
		o3,
		`14381 0 0 0 0 "" ""`,
		"AC",
	}, "\n")

	calls := decode(t, wrapScript(action))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1", len(calls))
	}
	if calls[0].StrRef != 14381 {
		t.Errorf("StrRef = %d, want 14381", calls[0].StrRef)
	}
	if calls[0].Call != "DisplayString(LastTrigger,14381)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "DisplayString(LastTrigger,14381)")
	}
}

// TestDecodeCompiledScript_MalformedBlockSkipped verifies a malformed action
// block is skipped without aborting the rest of the script: a broken block
// followed by a good one still yields the good one's pair.
func TestDecodeCompiledScript_MalformedBlockSkipped(t *testing.T) {
	// A malformed action: action id then garbage where the first OB should be.
	bad := "100AC\nnot-a-number\nrubbish\nAC"
	good := acBlock(
		actDisplayString,
		emptyObject(),
		obBlock(objLastTrigger, ""),
		emptyObject(),
		14381, 0, 0, 0, 0,
	)

	calls := decode(t, wrapScript(bad, good))
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d, want 1 (good block survives)", len(calls))
	}
	if calls[0].Call != "DisplayString(LastTrigger,14381)" {
		t.Errorf("Call = %q, want %q", calls[0].Call, "DisplayString(LastTrigger,14381)")
	}
}
