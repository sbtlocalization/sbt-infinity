// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/tealeg/xlsx/v4"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

func TestToAutoListShort(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		values []string
		want   string
	}{
		{"empty", "k", nil, ""},
		{"single value", "k", []string{"A.SPL"}, "k: A.SPL"},
		{"duplicates collapse to one, inline", "k", []string{"A.SPL", "A.SPL", "A.SPL"}, "k: A.SPL"},
		{"unique values sorted inline", "k", []string{"C.SPL", "A.SPL", "B.SPL"}, "k: A.SPL, B.SPL, C.SPL"},
		{"many duplicates, few unique, still inline", "k",
			[]string{"B.SPL", "A.SPL", "B.SPL", "A.SPL", "B.SPL", "A.SPL", "B.SPL"},
			"k: A.SPL, B.SPL"},
		{"exactly three unique stays inline", "k",
			[]string{"C.SPL", "B.SPL", "A.SPL"},
			"k: A.SPL, B.SPL, C.SPL"},
		{"one over the limit caps with (+N more)", "k",
			[]string{"D.SPL", "C.SPL", "B.SPL", "A.SPL"},
			"k: A.SPL, B.SPL, C.SPL (+1 more)"},
		{"many unique caps with (+N more)", "k",
			[]string{"H.SPL", "G.SPL", "F.SPL", "E.SPL", "D.SPL", "C.SPL", "B.SPL", "A.SPL"},
			"k: A.SPL, B.SPL, C.SPL (+5 more)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toAutoListShort(tt.key, tt.values); got != tt.want {
				t.Errorf("toAutoListShort(%q, %v) = %q, want %q", tt.key, tt.values, got, tt.want)
			}
		})
	}
}

func TestToAutoListLong(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		values []string
		want   string
	}{
		{"empty", "k", nil, ""},
		{"at or below the limit stays inline", "k",
			[]string{"C.SPL", "A.SPL", "B.SPL"},
			"k: A.SPL, B.SPL, C.SPL"},
		{"one over the limit becomes bullet list, sorted", "k",
			[]string{"D.SPL", "C.SPL", "B.SPL", "A.SPL"},
			"k: \n- A.SPL\n- B.SPL\n- C.SPL\n- D.SPL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toAutoListLong(tt.key, tt.values); got != tt.want {
				t.Errorf("toAutoListLong(%q, %v) = %q, want %q", tt.key, tt.values, got, tt.want)
			}
		})
	}
}

func TestToAutoListIsDeterministic(t *testing.T) {
	first := toAutoListShort("k", []string{"C.SPL", "A.SPL", "B.SPL"})
	second := toAutoListShort("k", []string{"B.SPL", "C.SPL", "A.SPL"})
	if first != second {
		t.Errorf("toAutoListShort not order-independent: %q != %q", first, second)
	}
}

func exportContextCell(t *testing.T, collection *TextCollection) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "context.xlsx")
	if err := collection.ExportToXlsx(path, nil); err != nil {
		t.Fatalf("ExportToXlsx: %v", err)
	}

	workbook, err := xlsx.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	row, err := workbook.Sheets[0].Row(1)
	if err != nil {
		t.Fatalf("reading exported row: %v", err)
	}
	return row.GetCell(3).Value
}

func TestExportToXlsx_DialogScriptContext(t *testing.T) {
	c := newTestCollection(14381)
	c.Entries[14381].Context[ContextDialogScript] = map[string][]string{
		"AardvarkMessage(Myself,14381)": {"OTHER"},
		"AddJournalEntry(14381,1)": {
			"https://example.com/dialog/test-1#transition-test-2-",
			"https://example.com/dialog/test-0#transition-test-2-",
		},
		"EraseJournalEntry(14381)": {"TEST", "OTHER", "TEST"},
		"SetQuestDone(14381)":      {"ZED", "ALPHA", "MIDDLE", "BETA", "ALPHA"},
	}

	got := exportContextCell(t, c)
	want := "DIALOG SCRIPTS: ----------\n" +
		"AddJournalEntry(14381,1):\n" +
		"- https://example.com/dialog/test-1#transition-test-2-\n" +
		"- https://example.com/dialog/test-0#transition-test-2-\n" +
		"AardvarkMessage(Myself,14381): OTHER\n" +
		"EraseJournalEntry(14381): OTHER, TEST\n" +
		"SetQuestDone(14381): ALPHA, BETA, MIDDLE (+1 more)"
	if got != want {
		t.Errorf("exported dialog script context =\n%q\nwant\n%q", got, want)
	}
}

func TestExportToXlsx_LongDialogScriptContextShowsEveryDialog(t *testing.T) {
	c := newTestCollection(14381)
	c.Long = true
	c.Entries[14381].Context[ContextDialogScript] = map[string][]string{
		"SetQuestDone(14381)": {"ZED", "ALPHA", "MIDDLE", "BETA", "ALPHA"},
	}

	got := exportContextCell(t, c)
	want := "DIALOG SCRIPTS: ----------\n" +
		"SetQuestDone(14381): \n" +
		"- ALPHA\n" +
		"- BETA\n" +
		"- MIDDLE\n" +
		"- ZED"
	if got != want {
		t.Errorf("exported long dialog script context =\n%q\nwant\n%q", got, want)
	}
}

func TestLocation(t *testing.T) {
	collapsed := &TextCollection{}
	if got := collapsed.location("Spell X.SPL → ability 0 → effect 1", "X.SPL"); got != "X.SPL" {
		t.Errorf("default location = %q, want bare file name", got)
	}

	long := &TextCollection{Long: true}
	verbose := "Spell X.SPL → ability 0 → effect 1"
	if got := long.location(verbose, "X.SPL"); got != verbose {
		t.Errorf("--long location = %q, want verbose string", got)
	}
}

func trackingTwoDA(t *testing.T, strrefValue string) *p.TwoDA {
	t.Helper()
	src := "2DA V1.0\n0\nSTRREF\nAR1000 " + strrefValue + "\n"
	twoda, err := p.ParseTwoDA(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseTwoDA: %v", err)
	}
	return twoda
}

func TestLoadContextFromTracking2DA_AreaSuffix(t *testing.T) {
	for _, long := range []bool{false, true} {
		c := newTestCollection(12345)
		c.Long = long

		if err := c.LoadContextFromTracking2DA("TRACKING.2DA", trackingTwoDA(t, "O_12345")); err != nil {
			t.Fatalf("LoadContextFromTracking2DA: %v", err)
		}

		values := c.Entries[12345].Context[ContextTracking2DA]["The text displayed by the rangers Tracking skill"]
		if len(values) != 1 || values[0] != "AR1000.ARE" {
			t.Errorf("long=%v: tracking area value = %v, want [\"AR1000.ARE\"]", long, values)
		}
	}
}
