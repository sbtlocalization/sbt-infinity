// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"slices"
	"strings"
	"testing"

	"github.com/sbtlocalization/sbt-infinity/dialog"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

func newTestCollection(ids ...uint32) *TextCollection {
	c := &TextCollection{Entries: make(map[uint32]*TextEntry)}
	for _, id := range ids {
		c.Entries[id] = &TextEntry{
			Id:      id,
			Labels:  make(map[string]struct{}),
			Context: make(map[ContextType]map[string][]string),
		}
	}
	return c
}

func newActionTransitionDialog(action string) *dialog.DialogCollection {
	id := dialog.NewNodeOrigin("TEST.DLG", 0)
	d := dialog.NewDialog(id)

	root := &dialog.Node{
		Type:   dialog.StateNodeType,
		Origin: dialog.NewNodeOrigin("TEST.DLG", 0),
		Dialog: d,
		State:  &dialog.StateData{},
	}

	transition := &dialog.Node{
		Type:   dialog.TransitionNodeType,
		Origin: dialog.NewNodeOrigin("TEST.DLG", 0),
		Dialog: d,
		Parent: root,
		Transition: &dialog.TransitionData{
			HasAction: true,
			Action:    action,
		},
	}
	root.Children = []*dialog.Node{transition}
	d.RootState = root

	return &dialog.DialogCollection{Dialogs: []*dialog.Dialog{d}}
}

func buildTestDisplayFamilyActions(t *testing.T) p.DisplayFamilyActions {
	t.Helper()
	ids := &p.Ids{
		Entries: map[int32]string{
			1: "DisplayStringHead(O:Object*,I:StrRef*)",
			2: "SetGlobal(S:Name*,S:Area*,I:Value*)",
		},
	}
	family, err := p.BuildDisplayFamilyActions(ids)
	if err != nil {
		t.Fatalf("BuildDisplayFamilyActions: %v", err)
	}
	return family
}

func TestLoadContextFromDialogs_DialogScript(t *testing.T) {
	c := newTestCollection(14381, 99999)
	dlg := newActionTransitionDialog("DisplayStringHead(Myself,14381)")
	family := buildTestDisplayFamilyActions(t)

	c.LoadContextFromDialogs("https://example.com", dlg, family)

	entry := c.Entries[14381]
	scripts, ok := entry.Context[ContextDialogScript]
	if !ok {
		t.Fatalf("expected DIALOG SCRIPTS context for 14381, got none")
	}

	values, ok := scripts["TEST[0]"]
	if !ok {
		t.Fatalf("expected context keyed by dialog name TEST[0], got keys %v", keysOf(scripts))
	}
	if len(values) != 1 {
		t.Fatalf("expected one context value, got %v", values)
	}
	if !strings.HasPrefix(values[0], "Action: DisplayStringHead(Myself,14381) @ ") {
		t.Errorf("unexpected context value: %q", values[0])
	}

	if _, ok := entry.Labels[lb_dialog_action]; !ok {
		t.Errorf("expected %q label on 14381, got %v", lb_dialog_action, keysOf(entry.Labels))
	}

	// Untouched ids must not gain any dialog-script context.
	if other := c.Entries[99999]; len(other.Context) != 0 || len(other.Labels) != 0 {
		t.Errorf("untouched id 99999 was modified: context=%v labels=%v", other.Context, other.Labels)
	}
}

func TestLoadContextFromDialogs_DialogScriptOnlyMatchingCall(t *testing.T) {
	c := newTestCollection(14381, 14382)
	dlg := newActionTransitionDialog("SetGlobal(\"v\",\"GLOBAL\",1)\nDisplayStringHead(Myself,14381)\nDisplayStringHead(Player1,14382)")
	family := buildTestDisplayFamilyActions(t)

	c.LoadContextFromDialogs("https://example.com", dlg, family)

	for strref, call := range map[uint32]string{
		14381: "DisplayStringHead(Myself,14381)",
		14382: "DisplayStringHead(Player1,14382)",
	} {
		values := c.Entries[strref].Context[ContextDialogScript]["TEST[0]"]
		if len(values) != 1 {
			t.Fatalf("expected one context value for %d, got %v", strref, values)
		}
		if !strings.HasPrefix(values[0], "Action: "+call+" @ ") {
			t.Errorf("unexpected context value for %d: %q", strref, values[0])
		}
		if strings.Contains(values[0], "SetGlobal") {
			t.Errorf("context value for %d leaks sibling call: %q", strref, values[0])
		}
	}
}

func TestLoadContextFromDialogs_NonIntegerStrRefSkipped(t *testing.T) {
	c := newTestCollection(14381)
	// SetGlobal is not display-family; nothing should be linked.
	dlg := newActionTransitionDialog(`SetGlobal("v","GLOBAL",1)`)
	family := buildTestDisplayFamilyActions(t)

	c.LoadContextFromDialogs("https://example.com", dlg, family)

	if entry := c.Entries[14381]; len(entry.Context[ContextDialogScript]) != 0 {
		t.Errorf("expected no dialog-script context, got %v", entry.Context[ContextDialogScript])
	}
}

func TestLoadContextFromCompiledScript(t *testing.T) {
	c := newTestCollection(14381, 99999)
	calls := []p.ScriptCall{
		{StrRef: 14381, Call: "DisplayString(LastTrigger,14381)"},
	}

	// A nil usage index (e.g. --context-from bcs alone) keeps the bare key.
	c.LoadContextFromCompiledScript("CTFB.BCS", calls, nil)

	entry := c.Entries[14381]
	scripts, ok := entry.Context[ContextScript]
	if !ok {
		t.Fatalf("expected SCRIPTS context for 14381, got none")
	}

	values, ok := scripts["CTFB.BCS"]
	if !ok {
		t.Fatalf("expected context keyed by CTFB.BCS, got keys %v", keysOf(scripts))
	}
	if len(values) != 1 || values[0] != "DisplayString(LastTrigger,14381)" {
		t.Errorf("unexpected context values: %v", values)
	}

	if _, ok := entry.Labels[lb_script_action]; !ok {
		t.Errorf("expected %q label on 14381, got %v", lb_script_action, keysOf(entry.Labels))
	}

	// Untouched ids must not gain any script context.
	if other := c.Entries[99999]; len(other.Context) != 0 || len(other.Labels) != 0 {
		t.Errorf("untouched id 99999 was modified: context=%v labels=%v", other.Context, other.Labels)
	}
}

func TestLoadContextFromCompiledScript_WithUsage(t *testing.T) {
	c := newTestCollection(14381)
	calls := []p.ScriptCall{
		{StrRef: 14381, Call: "DisplayString(LastTrigger,14381)"},
	}

	usage := make(ScriptUsageIndex)
	usage.Add("CTFB", "trap", "AR1234")

	c.LoadContextFromCompiledScript("CTFB.BCS", calls, usage)

	scripts := c.Entries[14381].Context[ContextScript]
	if _, ok := scripts["CTFB.BCS [trap @ AR1234]"]; !ok {
		t.Fatalf("expected context keyed by annotated script name, got keys %v", keysOf(scripts))
	}
}

func keysOf[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
