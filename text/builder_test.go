// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

func newTestCollection(ids ...uint32) *TextCollection {
	c := &TextCollection{Entries: make(map[uint32]*TextEntry)}
	for _, id := range ids {
		c.Entries[id] = &TextEntry{
			Id:      id,
			Labels:  make(map[string]LabelFamily),
			Context: make(map[ContextType]map[string][]string),
		}
	}
	return c
}

func newActionTransitionDialog(action string) *dialog.DialogCollection {
	return newActionTransitionDialogAt("TEST.DLG", 0, action)
}

func newActionTransitionDialogAt(dlgName string, index uint32, action string) *dialog.DialogCollection {
	id := dialog.NewNodeOrigin(dlgName, index)
	d := dialog.NewDialog(id)

	root := &dialog.Node{
		Type:   dialog.StateNodeType,
		Origin: dialog.NewNodeOrigin(dlgName, index),
		Dialog: d,
		State:  &dialog.StateData{},
	}

	transition := &dialog.Node{
		Type:   dialog.TransitionNodeType,
		Origin: dialog.NewNodeOrigin(dlgName, 0),
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
			3: "SetQuestDone(I:StrRef*)",
			4: "EraseJournalEntry(I:StrRef*)",
			5: "AddJournalEntry(I:Entry*,I:Type*JourType)",
			6: "FloatMessage(O:Object*,I:StrRef*)",
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

	values, ok := scripts["DisplayStringHead(Myself,14381)"]
	if !ok {
		t.Fatalf("expected context keyed by decoded call, got keys %v", keysOf(scripts))
	}
	if len(values) != 1 || values[0] != "TEST" {
		t.Fatalf("expected extension-less dialog name as context value, got %v", values)
	}

	for _, label := range []string{lb_dialog_action, "DisplayStringHead", "TEST"} {
		if _, ok := entry.Labels[label]; !ok {
			t.Errorf("expected %q label on 14381, got %v", label, keysOf(entry.Labels))
		}
	}

	// Untouched ids must not gain any dialog-script context.
	if other := c.Entries[99999]; len(other.Context) != 0 || len(other.Labels) != 0 {
		t.Errorf("untouched id 99999 was modified: context=%v labels=%v", other.Context, other.Labels)
	}
}

func TestLoadContextFromDialogs_DialogScriptOnlyMatchingCall(t *testing.T) {
	c := newTestCollection(14381)
	dlg := newActionTransitionDialog("SetGlobal(\"v\",\"GLOBAL\",1)\nFloatMessage(Myself,14381)\nFloatMessage(Player1,14381)")
	family := buildTestDisplayFamilyActions(t)

	c.LoadContextFromDialogs("https://example.com", dlg, family)

	scripts := c.Entries[14381].Context[ContextDialogScript]
	for _, call := range []string{
		"FloatMessage(Myself,14381)",
		"FloatMessage(Player1,14381)",
	} {
		values := scripts[call]
		if len(values) != 1 || values[0] != "TEST" {
			t.Errorf("context values for %q = %v, want [TEST]", call, values)
		}
	}
	if len(scripts) != 2 {
		t.Errorf("context calls = %v, want only the two display-family calls", keysOf(scripts))
	}
}

func TestLoadContextFromDialogs_JournalActionsKeepNodeURLs(t *testing.T) {
	c := newTestCollection(14381)
	family := buildTestDisplayFamilyActions(t)
	action := "SetQuestDone(14381)\nEraseJournalEntry(14381)\nAddJournalEntry(14381,1)"

	for _, root := range []struct {
		name  string
		index uint32
	}{
		{"TEST.DLG", 0},
		{"TEST.DLG", 1},
		{"OTHER.DLG", 0},
	} {
		c.LoadContextFromDialogs("https://example.com", newActionTransitionDialogAt(root.name, root.index, action), family)
	}

	entry := c.Entries[14381]
	scripts := entry.Context[ContextDialogScript]
	for _, call := range []string{"SetQuestDone(14381)", "EraseJournalEntry(14381)"} {
		values := scripts[call]
		if len(values) != 3 || slices.ContainsFunc(values, func(value string) bool { return strings.Contains(value, "://") }) {
			t.Errorf("context values for %q = %v, want dialog names without node URLs", call, values)
		}
	}

	journalURLs := scripts["AddJournalEntry(14381,1)"]
	if len(journalURLs) != 3 {
		t.Fatalf("journal URLs = %v, want one for every dialog root", journalURLs)
	}
	for _, url := range journalURLs {
		if !strings.HasPrefix(url, "https://example.com/dialog/") {
			t.Errorf("journal context value = %q, want node URL", url)
		}
	}

	for _, label := range []string{
		lb_dialog_action,
		lb_dialog_journal,
		"SetQuestDone",
		"EraseJournalEntry",
		"AddJournalEntry",
		"TEST",
		"OTHER",
	} {
		if _, ok := entry.Labels[label]; !ok {
			t.Errorf("expected %q label, got %v", label, keysOf(entry.Labels))
		}
	}
}

func TestLoadContextFromDialogs_NonIntegerStrRefSkipped(t *testing.T) {
	c := newTestCollection(14381)
	dlg := newActionTransitionDialog("SetGlobal(\"v\",\"GLOBAL\",1)\nFloatMessage(Myself,GLOBAL_STRREF)")
	family := buildTestDisplayFamilyActions(t)

	c.LoadContextFromDialogs("https://example.com", dlg, family)

	entry := c.Entries[14381]
	if len(entry.Context[ContextDialogScript]) != 0 || len(entry.Labels) != 0 {
		t.Errorf("ignored calls modified the entry: context=%v labels=%v", entry.Context[ContextDialogScript], entry.Labels)
	}
}

func TestLoadContextFromCompiledScript(t *testing.T) {
	c := newTestCollection(14381, 99999)
	lastTriggerCall := "DisplayString(LastTrigger,14381)"
	myselfCall := "DisplayString(Myself,14381)"

	c.LoadContextFromCompiledScript("SECOND.BCS", []p.ScriptCall{{StrRef: 14381, Call: lastTriggerCall}}, nil)
	c.LoadContextFromCompiledScript("FIRST.BCS", []p.ScriptCall{
		{StrRef: 14381, Call: lastTriggerCall},
		{StrRef: 14381, Call: myselfCall},
	}, nil)

	entry := c.Entries[14381]
	scripts, ok := entry.Context[ContextScript]
	if !ok {
		t.Fatalf("expected SCRIPTS context for 14381, got none")
	}
	if got := scripts[lastTriggerCall]; !slices.Equal(got, []string{"SECOND.BCS", "FIRST.BCS"}) {
		t.Errorf("scripts for shared call = %v, want both source files", got)
	}
	if got := scripts[myselfCall]; !slices.Equal(got, []string{"FIRST.BCS"}) {
		t.Errorf("scripts for distinct object call = %v, want FIRST.BCS", got)
	}

	for _, label := range []string{lb_script_action, "DisplayString"} {
		if _, ok := entry.Labels[label]; !ok {
			t.Errorf("expected %q label on 14381, got %v", label, keysOf(entry.Labels))
		}
	}
	for _, label := range []string{"FIRST.BCS", "SECOND.BCS"} {
		if _, ok := entry.Labels[label]; ok {
			t.Errorf("script file %q must not be a label, got %v", label, keysOf(entry.Labels))
		}
	}

	if other := c.Entries[99999]; len(other.Context) != 0 || len(other.Labels) != 0 {
		t.Errorf("untouched id 99999 was modified: context=%v labels=%v", other.Context, other.Labels)
	}
}

func TestLoadContextFromCompiledScript_AddsUsageLabels(t *testing.T) {
	c := newTestCollection(14381)
	usage := make(ScriptUsageIndex)
	usage.Add("CTFB", "trap", "AR1234.ARE")
	usage.Add("CTFB", "trap", "AR1234.ARE")
	usage.Add("CTFB", "creature", "CHICKEN.CRE")

	c.LoadContextFromCompiledScript("CTFB.BCS", []p.ScriptCall{
		{StrRef: 14381, Call: "DisplayString(LastTrigger,14381)"},
	}, usage)

	entry := c.Entries[14381]
	for _, label := range []string{"AR1234.ARE", "CHICKEN.CRE", "trap", "creature"} {
		if _, ok := entry.Labels[label]; !ok {
			t.Errorf("expected usage label %q, got %v", label, keysOf(entry.Labels))
		}
	}
	context := exportContextCell(t, c)
	if strings.Contains(context, "trap") || strings.Contains(context, "AR1234") {
		t.Errorf("usage annotation leaked into context body: %q", context)
	}
}

func TestLoadContextFromCompiledScript_UsageLabelsUseFamilyCap(t *testing.T) {
	c := newTestCollection(14381)
	usage := make(ScriptUsageIndex)
	for i := range 11 {
		usage.Add("SHOUT", "creature", fmt.Sprintf("CRE%02d.CRE", i))
	}

	c.LoadContextFromCompiledScript("SHOUT.BCS", []p.ScriptCall{
		{StrRef: 14381, Call: "DisplayStringHead(Myself,14381)"},
	}, usage)

	got := exportCell(t, c, 2)
	want := "*.CRE,DisplayStringHead,creature,script action"
	if got != want {
		t.Errorf("exported labels = %q, want %q", got, want)
	}
}

// A valid 156-byte STO header whose drinks section claims one entry at an
// offset sitting on EOF, so Sto.Drinks() fails while the header (and thus the
// store name) stays readable — the seam a swallowed lazy-getter error hides.
func brokenDrinksStore(t *testing.T, nameRef uint32) *p.Sto {
	t.Helper()

	const headerSize = 156
	buf := make([]byte, headerSize)
	copy(buf[0:4], "STOR")
	copy(buf[4:8], "V1.0")
	binary.LittleEndian.PutUint32(buf[12:16], nameRef)    // NameRef
	binary.LittleEndian.PutUint32(buf[76:80], headerSize) // OfsDrinks → EOF
	binary.LittleEndian.PutUint32(buf[80:84], 1)          // NumDrinks

	sto := p.NewSto()
	if err := sto.Read(kaitai.NewStream(bytes.NewReader(buf)), nil, sto); err != nil {
		t.Fatalf("crafting broken STO: %v", err)
	}
	// Do not probe Drinks() here: the lazy getter memoizes on first call
	// (including a failing one), so a probe would return a cached nil error and
	// mask the failure the loader must surface.
	return sto
}

func TestLoadContextFromStore_LazyGetterFailureSurfacedAndPartial(t *testing.T) {
	const nameRef = 14381
	c := newTestCollection(nameRef)
	sto := brokenDrinksStore(t, nameRef)

	err := c.LoadContextFromStore("SHOP.STO", sto)
	if err == nil {
		t.Fatal("LoadContextFromStore() err = nil, want the drinks-read failure surfaced")
	}

	entry := c.Entries[nameRef]
	values := entry.Context[ContextStore]["Store name"]
	if !slices.Contains(values, "SHOP.STO") {
		t.Errorf("readable store name context was dropped: %v", entry.Context[ContextStore])
	}
	if _, ok := entry.Labels[lb_store]; !ok {
		t.Errorf("expected %q label on %d, got %v", lb_store, nameRef, keysOf(entry.Labels))
	}
}

func makeTwoDA(t *testing.T, body string) *p.TwoDA {
	t.Helper()
	twoda, err := p.ParseTwoDA(strings.NewReader(body))
	if err != nil {
		t.Fatalf("ParseTwoDA: %v", err)
	}
	return twoda
}

func minimalCreature(t *testing.T, soundRef uint32, dialog string) *p.Cre {
	t.Helper()

	buf := make([]byte, 0x2d4)
	copy(buf[0:4], "CRE ")
	copy(buf[4:8], "V1.0")
	binary.LittleEndian.PutUint32(buf[0xa4:0xa8], soundRef)
	copy(buf[0x2cc:0x2d4], dialog)

	cre := p.NewCre()
	if err := cre.Read(kaitai.NewStream(bytes.NewReader(buf)), nil, cre); err != nil {
		t.Fatalf("crafting minimal CRE: %v", err)
	}
	return cre
}

func minimalArea(t *testing.T, script string) *p.Are {
	t.Helper()

	buf := make([]byte, 0x11c)
	copy(buf[0:4], "AREA")
	copy(buf[4:8], "V1.0")
	are := p.NewAre()
	if err := are.Read(kaitai.NewStream(bytes.NewReader(buf)), nil, are); err != nil {
		t.Fatalf("crafting minimal ARE: %v", err)
	}
	are.AreaScriptBcs = script
	return are
}

func TestLoadContextFromArea_ScriptUsageKeepsAreaExtension(t *testing.T) {
	usage := make(ScriptUsageIndex)
	if err := newTestCollection().LoadContextFromArea("AR9711.ARE", minimalArea(t, "AREA9711"), usage); err != nil {
		t.Fatalf("LoadContextFromArea: %v", err)
	}

	want := scriptUsage{slot: "area", file: "AR9711.ARE"}
	if got := usage["AREA9711"]; len(got) != 1 || got[0] != want {
		t.Errorf("area script usage = %v, want %v", got, want)
	}
}

func TestLoadContextFromCreature_ScriptSlotsCollapseToCreatureUsage(t *testing.T) {
	cre := minimalCreature(t, 0, "")
	cre.Body.Header.OverrideScript = "OVERRIDE"
	cre.Body.Header.ClassScript = "CLASS"
	cre.Body.Header.RaceScript = "RACE"
	cre.Body.Header.GeneralScript = "GENERAL"
	cre.Body.Header.DefaultScript = "DEFAULT"
	usage := make(ScriptUsageIndex)

	if err := newTestCollection().LoadContextFromCreature("CHICKEN.CRE", cre, &p.Ids{}, usage); err != nil {
		t.Fatalf("LoadContextFromCreature: %v", err)
	}

	for _, script := range []string{"OVERRIDE", "CLASS", "RACE", "GENERAL", "DEFAULT"} {
		want := scriptUsage{slot: "creature", file: "CHICKEN.CRE"}
		if got := usage[script]; len(got) != 1 || got[0] != want {
			t.Errorf("usage for %s = %v, want %v", script, got, want)
		}
	}
}

func TestLoadContextFromCreature_VerbalConstantsAreSpeakerLabels(t *testing.T) {
	const ref = 14381
	c := newTestCollection(ref)
	ids := &p.Ids{Entries: map[int32]string{0: "SELECT_COMMON"}}

	for i := range 11 {
		filename := fmt.Sprintf("CRE%02d.CRE", i)
		dialog := fmt.Sprintf("DLG%05d", i)
		if err := c.LoadContextFromCreature(filename, minimalCreature(t, ref, dialog), ids, nil); err != nil {
			t.Fatalf("LoadContextFromCreature(%q): %v", filename, err)
		}
	}

	got := exportCell(t, c, 2)
	want := "*.CRE,many speakers"
	if got != want {
		t.Errorf("exported verbal-constant labels = %q, want %q", got, want)
	}
}

func TestLoadContextFromCreature_SourceFileLabelRawCase(t *testing.T) {
	const ref = 14381
	c := newTestCollection(ref)

	cre := p.NewCre()
	cre.LongNameRef = ref
	cre.ShortNameRef = ref
	// Body is nil, so the loader records the name labels and then bails out
	// with an "unsupported version" error before reaching sounds/effects.
	_ = c.LoadContextFromCreature("Fumble.CRE", cre, nil, nil)

	entry := c.Entries[ref]
	if _, ok := entry.Labels["Fumble.CRE"]; !ok {
		t.Errorf("expected raw-case source-file label %q, got %v", "Fumble.CRE", keysOf(entry.Labels))
	}
	// The context body lower-cases the name; the label must keep raw case.
	if _, ok := entry.Labels["fumble.cre"]; ok {
		t.Errorf("label must keep raw filesystem case, found lower-cased %q", "fumble.cre")
	}
	if got := entry.Context[ContextCreature]["Long name"]; !slices.Contains(got, "fumble.cre") {
		t.Errorf("context body should still lower-case the name, got %v", got)
	}
}

func TestLoadContextFromTracking2DA_SourceFileLabelIsArea(t *testing.T) {
	const standalone = 14381
	const embedded = 24000
	c := newTestCollection(standalone, embedded, 67807)

	twoda := makeTwoDA(t, "2DA V1.0\n-1\nSTRREF\nAR1000 O_14381\nAR2000 24000\n")
	if err := c.LoadContextFromTracking2DA("TRACKING.2DA", twoda); err != nil {
		t.Fatalf("LoadContextFromTracking2DA: %v", err)
	}

	if _, ok := c.Entries[standalone].Labels["AR1000.ARE"]; !ok {
		t.Errorf("expected label AR1000.ARE, got %v", keysOf(c.Entries[standalone].Labels))
	}
	if _, ok := c.Entries[embedded].Labels["AR2000.ARE"]; !ok {
		t.Errorf("expected label AR2000.ARE, got %v", keysOf(c.Entries[embedded].Labels))
	}
	// The 2DA's own filename must not stand in for the area name here.
	if _, ok := c.Entries[standalone].Labels["TRACKING.2DA"]; ok {
		t.Errorf("tracking label must be the area name, not the 2DA file name")
	}
}

func TestLoadContextFromEffText2DA_SourceFileLabel(t *testing.T) {
	const ref = 35501
	c := newTestCollection(ref)

	twoda := makeTwoDA(t, "2DA V1.0\n-1\nEFFECT_NAME STRREF\n25 POISON 35501\n")
	if err := c.LoadContextFromEffText2DA("EFFTEXT.2DA", twoda); err != nil {
		t.Fatalf("LoadContextFromEffText2DA: %v", err)
	}

	if _, ok := c.Entries[ref].Labels["EFFTEXT.2DA"]; !ok {
		t.Errorf("expected source-file label %q, got %v", "EFFTEXT.2DA", keysOf(c.Entries[ref].Labels))
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
