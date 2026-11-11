// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
	cmdDialog "github.com/sbtlocalization/sbt-infinity/cmd/dialog"
	cmdText "github.com/sbtlocalization/sbt-infinity/cmd/text"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/text"
)

type fixtureTransition struct {
	flags, text, journal, target uint32
	resource, trigger, action    string
}

type fixtureState struct {
	text        uint32
	root        bool
	transitions []fixtureTransition
}

func fixtureDialog(states []fixtureState) []byte {
	count := 0
	for _, s := range states {
		count += len(s.transitions)
	}
	stateOffset := 52
	transitionOffset := stateOffset + 16*len(states)
	stateTriggerOffset := transitionOffset + 32*count
	triggerOffset := stateTriggerOffset + 8
	actionOffset := triggerOffset + 8*count
	data := make([]byte, actionOffset+8*count)
	copy(data, "DLG V1.0")
	put := func(off int, n uint32) { binary.LittleEndian.PutUint32(data[off:], n) }
	put(8, uint32(len(states)))
	put(12, uint32(stateOffset))
	put(16, uint32(count))
	put(20, uint32(transitionOffset))
	put(24, uint32(stateTriggerOffset))
	put(28, 1)
	put(32, uint32(triggerOffset))
	put(36, uint32(count))
	put(40, uint32(actionOffset))
	put(44, uint32(count))
	addText := func(offset int, value string) {
		put(offset, uint32(len(data)))
		put(offset+4, uint32(len(value)))
		data = append(data, value...)
	}
	addText(stateTriggerOffset, "True()")
	index := 0
	for i, s := range states {
		o := stateOffset + i*16
		put(o, s.text)
		put(o+4, uint32(index))
		put(o+8, uint32(len(s.transitions)))
		if !s.root {
			put(o+12, ^uint32(0))
		}
		for _, tr := range s.transitions {
			o := transitionOffset + index*32
			put(o, tr.flags)
			put(o+4, tr.text)
			put(o+8, tr.journal)
			put(o+12, uint32(index))
			put(o+16, uint32(index))
			copy(data[o+20:o+28], tr.resource)
			put(o+28, tr.target)
			addText(triggerOffset+index*8, tr.trigger)
			addText(actionOffset+index*8, tr.action)
			index++
		}
	}
	return data
}

func graphResources(reverse bool) []resource {
	branch1 := fixtureTransition{flags: 0x17, text: 5, journal: 8, resource: "fOo%", target: 1, trigger: "Never()", action: "AddJournalEntry(9,1)"}
	branch2 := fixtureTransition{flags: 2, resource: "FOO%", target: 2, trigger: "Second()"}
	branches := []fixtureTransition{branch1, branch2}
	if reverse {
		branches[0], branches[1] = branches[1], branches[0]
	}
	foo := fixtureDialog([]fixtureState{
		{text: 1, root: true, transitions: branches},
		{text: 2, transitions: []fixtureTransition{
			{flags: 2, resource: "bAr", target: 0, trigger: "First()"},
			{flags: 2, resource: "FOO%", target: 0, trigger: "Ancestor()"},
		}},
		{text: 3, transitions: []fixtureTransition{
			{flags: 2, resource: "BAR", target: 0, trigger: "Third()"},
			{flags: 2, resource: "bar", target: 0, trigger: "Parallel()"},
		}},
	})
	bar := fixtureDialog([]fixtureState{{text: 4, root: true, transitions: []fixtureTransition{
		{flags: 1, text: 6, resource: "foo%", target: 1},
		{flags: 2, resource: "bAr", target: 0, trigger: "Self()"},
		{flags: 9, text: 7},
	}}})
	return []resource{{"FoO%", 1011, foo}, {"BaR", 1011, bar}, {"ACTION", 1008, []byte("IDS V1.0\n173 AddJournalEntry(I:Entry*,I:Type*JourType)\n")}}
}

func graphFixture(t *testing.T, reverse bool) (string, string) {
	t.Helper()
	dir := t.TempDir()
	key, bif := makeArchive(graphResources(reverse))
	entries := make([]text.TlkWriteEntry, 12)
	for i := range entries {
		entries[i] = text.TlkWriteEntry{Text: fmt.Sprintf("Text %d", i), HasText: true}
	}
	var tlk bytes.Buffer
	if err := text.WriteTlk(&tlk, entries, text.TlkWriteOptions{}); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"chitin.key": key, "test.bif": bif, "dialog.tlk": tlk.Bytes()} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "chitin.key"), filepath.Join(dir, "dialog.tlk")
}

func exportGraph(t *testing.T, key, tlk, resource, out, stem string) *dcanvas.Canvas {
	t.Helper()
	cmd := cmdDialog.NewCommand()
	cmd.PersistentFlags().String("key", key, "")
	cmd.SetArgs([]string{"export", "--speakers=false", "--tlk", tlk, "--output", out, resource})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, stem+".d.canvas"))
	if err != nil {
		t.Fatal(err)
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
	canvas, err := dcanvas.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(canvas); err != nil {
		t.Fatal(err)
	}
	return canvas
}

func assertAcyclic(t *testing.T, c *dcanvas.Canvas) {
	t.Helper()
	adj := map[string][]string{}
	for _, e := range c.Edges {
		if e.Kind != dcanvas.KindLoop {
			adj[e.FromNode] = append(adj[e.FromNode], e.ToNode)
		}
	}
	color := map[string]int{}
	var visit func(string)
	visit = func(id string) {
		if color[id] == 1 {
			t.Fatalf("non-loop edges contain a cycle at %s", id)
		}
		if color[id] == 2 {
			return
		}
		color[id] = 1
		for _, next := range adj[id] {
			visit(next)
		}
		color[id] = 2
	}
	for _, n := range c.Nodes {
		visit(n.ID)
	}
}

func TestGraphClassificationAndAddressExports(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%v", reverse), func(t *testing.T) {
			key, tlk := graphFixture(t, reverse)
			c := exportGraph(t, key, tlk, "foo%.dlg", filepath.Join(t.TempDir(), "mixed"), "FoO%-0")
			canonical := exportGraph(t, key, tlk, "FoO%.DLG", filepath.Join(t.TempDir(), "canonical"), "FoO%-0")
			if !reflect.DeepEqual(c, canonical) {
				t.Fatal("explicit casing changed the exported graph")
			}
			assertAcyclic(t, c)
			nodes := map[string]*dcanvas.Node{}
			for _, n := range c.Nodes {
				if nodes[n.ID] != nil {
					t.Fatalf("duplicate canvas ID %q", n.ID)
				}
				nodes[n.ID] = n
			}
			for id, domain := range map[string]string{
				"state-FoO%[0]": "state-0", "state-FoO%[1]": "state-1", "state-FoO%[2]": "state-2",
				"state-BaR[0]": "state-BaR-0", "transition-BaR[0]": "transition-BaR-0", "transition-BaR[2]": "transition-BaR-2",
			} {
				if n := nodes[id]; n == nil || n.DomainID != domain {
					t.Errorf("node %s = %+v, want d-id %q", id, n, domain)
				}
			}
			branchID := "transition-FoO%[0]"
			if reverse {
				branchID = "transition-FoO%[1]"
			}
			if n := nodes[branchID]; n == nil || n.DomainID != fmt.Sprintf("transition-%d", map[bool]int{false: 0, true: 1}[reverse]) || n.Action != "AddJournalEntry(9,1)" {
				t.Errorf("root reply = %+v", n)
			}
			if len(nodes) != 7 {
				t.Errorf("got %d retained nodes, want 7", len(nodes))
			}
			expected := map[string]dcanvas.EdgeKind{
				"state-FoO%[2]>state-BaR[0]:Third()":     dcanvas.KindNormal,
				"state-FoO%[2]>state-BaR[0]:Parallel()":  dcanvas.KindNormal,
				"state-FoO%[1]>state-FoO%[0]:Ancestor()": dcanvas.KindLoop,
				"state-BaR[0]>state-BaR[0]:Self()":       dcanvas.KindLoop,
			}
			expected["state-FoO%[0]>"+branchID+":Never()"] = dcanvas.KindNormal
			if reverse {
				expected["state-FoO%[1]>state-BaR[0]:First()"] = dcanvas.KindLoop
				expected["transition-BaR[0]>state-FoO%[1]:"] = dcanvas.KindNormal
			} else {
				expected["state-FoO%[1]>state-BaR[0]:First()"] = dcanvas.KindNormal
				expected["transition-BaR[0]>state-FoO%[1]:"] = dcanvas.KindLoop
			}
			for _, e := range c.Edges {
				key := fmt.Sprintf("%s>%s:%s", e.FromNode, e.ToNode, e.Condition)
				if kind, ok := expected[key]; ok {
					if e.Kind != kind {
						t.Errorf("edge %s kind = %s, want %s", key, e.Kind, kind)
					}
					delete(expected, key)
				}
			}
			if len(expected) != 0 {
				t.Errorf("missing edges: %v", expected)
			}
			layoutGraph := &dcanvas.Canvas{}
			withConvergence := &dcanvas.Canvas{}
			for _, n := range c.Nodes {
				copy := *n
				copy.X, copy.Y = 0, 0
				layoutGraph.Nodes = append(layoutGraph.Nodes, &copy)
				other := copy
				withConvergence.Nodes = append(withConvergence.Nodes, &other)
			}
			for _, e := range c.Edges {
				if e.Kind == dcanvas.KindLoop {
					continue
				}
				withConvergence.Edges = append(withConvergence.Edges, e)
				convergence := e.FromNode == "state-FoO%[2]" && e.ToNode == "state-BaR[0]" &&
					(e.Condition == "Parallel()" || !reverse && e.Condition == "Third()") ||
					reverse && e.FromNode == branchID && e.ToNode == "state-FoO%[1]"
				if !convergence {
					layoutGraph.Edges = append(layoutGraph.Edges, e)
				}
			}
			if err := layout.Layout(layoutGraph); err != nil {
				t.Fatal(err)
			}
			for i, n := range c.Nodes {
				if n.X != layoutGraph.Nodes[i].X || n.Y != layoutGraph.Nodes[i].Y {
					t.Errorf("loop or convergence influenced layout of %s", n.ID)
				}
			}
			if !reverse {
				if err := layout.Layout(withConvergence); err != nil {
					t.Fatal(err)
				}
				changed := false
				for i, n := range layoutGraph.Nodes {
					changed = changed || n.X != withConvergence.Nodes[i].X || n.Y != withConvergence.Nodes[i].Y
				}
				if !changed {
					t.Error("fixture does not demonstrate convergence affecting layout")
				}
			}
			builder := dialog.NewDialogBuilder(fs.NewInfinityFs(key), nil, false, false)
			defer builder.Close()
			collection, err := builder.LoadAllDialogs("", "foo%")
			if err != nil || len(collection.Dialogs) != 1 {
				t.Fatalf("load dialogs: %+v, %v", collection, err)
			}
			if got := collection.Dialogs[0].NodeCount(); got != len(c.Nodes) {
				t.Errorf("node count = %d, canvas nodes = %d", got, len(c.Nodes))
			}
			checkListLength(t, key, tlk, len(c.Nodes))
			if !reverse {
				bar := exportGraph(t, key, tlk, "bAr", filepath.Join(t.TempDir(), "bar"), "BaR-0")
				checkTextExport(t, key, tlk, map[string]*dcanvas.Canvas{"FoO%-0": c, "BaR-0": bar})
			}
		})
	}
}

func checkListLength(t *testing.T, key, tlk string, want int) {
	t.Helper()
	cmd := cmdDialog.NewCommand()
	cmd.PersistentFlags().String("key", key, "")
	cmd.PersistentFlags().String("tlk", tlk, "")
	cmd.SetArgs([]string{"list", "--with-length", "--json", "foo%"})
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = writer
	err = cmd.Execute()
	os.Stdout = stdout
	writer.Close()
	output, readErr := io.ReadAll(reader)
	reader.Close()
	if err != nil || readErr != nil {
		t.Fatalf("list: %v, read: %v", err, readErr)
	}
	var result struct {
		Nodes int `json:"nodes"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("list output %q: %v", output, err)
	}
	if result.Nodes != want {
		t.Errorf("list nodes = %d, want %d", result.Nodes, want)
	}
}

func checkTextExport(t *testing.T, key, tlk string, canvases map[string]*dcanvas.Canvas) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "text.xlsx")
	cmd := cmdText.NewCommand()
	cmd.PersistentFlags().String("key", key, "")
	cmd.SetArgs([]string{"export", "--tlk", tlk, "--context-from", "DLG", "--dlg-base-url", "https://viewer.example", "--output", out})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	wb, err := xlsx.OpenFile(out)
	if err != nil {
		t.Fatal(err)
	}
	contexts := map[int]string{}
	labels := map[int]string{}
	for i := 1; i < wb.Sheets[0].MaxRow; i++ {
		row, err := wb.Sheets[0].Row(i)
		if err != nil {
			t.Fatal(err)
		}
		id, err := row.GetCell(0).Int()
		if err != nil {
			t.Fatal(err)
		}
		labels[id], contexts[id] = row.GetCell(2).Value, row.GetCell(3).Value
	}
	for id, fragment := range map[int]string{1: "state-0", 2: "state-1", 3: "state-2", 4: "state-BaR-0", 5: "transition-0", 6: "transition-BaR-0", 7: "transition-BaR-2", 8: "transition-0", 9: "transition-0"} {
		want := "https://viewer.example/dialog/FoO%25-0#" + fragment
		if !strings.Contains(contexts[id], want) {
			t.Errorf("text %d context = %q, want %q", id, contexts[id], want)
		}
		parsed, err := url.Parse(want)
		if err != nil || parsed.Path != "/dialog/FoO%-0" {
			t.Fatalf("URL %q: %+v, %v", want, parsed, err)
		}
		matches := 0
		for _, n := range canvases["FoO%-0"].Nodes {
			if n.DomainID == parsed.Fragment {
				matches++
			}
		}
		if matches != 1 {
			t.Errorf("fragment %q matches %d nodes", parsed.Fragment, matches)
		}
	}
	for id, label := range map[int]string{1: "question", 5: "answer", 8: "journal", 9: "AddJournalEntry"} {
		if !strings.Contains(labels[id], label) {
			t.Errorf("text %d labels = %q, missing %q", id, labels[id], label)
		}
	}
	if !strings.Contains(contexts[4], "https://viewer.example/dialog/BaR-0#state-0") || !strings.Contains(contexts[6], "https://viewer.example/dialog/BaR-0#transition-0") {
		t.Errorf("cross-resource IDs not scoped to the root canvas: state=%q reply=%q", contexts[4], contexts[6])
	}
	origins := map[int]string{
		1: "state-FoO%[0]", 2: "state-FoO%[1]", 3: "state-FoO%[2]", 4: "state-BaR[0]",
		5: "transition-FoO%[0]", 6: "transition-BaR[0]", 7: "transition-BaR[2]",
		8: "transition-FoO%[0]", 9: "transition-FoO%[0]",
	}
	for ref, origin := range origins {
		seen := map[string]int{}
		for _, line := range strings.Split(contexts[ref], "\n") {
			if !strings.HasPrefix(line, "- https://viewer.example/dialog/") {
				continue
			}
			link, err := url.Parse(strings.TrimPrefix(line, "- "))
			if err != nil {
				t.Fatal(err)
			}
			stem := strings.TrimPrefix(link.Path, "/dialog/")
			canvas := canvases[stem]
			if canvas == nil {
				t.Errorf("text %d URL targets unknown canvas %q", ref, stem)
				continue
			}
			seen[stem]++
			rawFragment := strings.SplitN(line, "#", 2)[1]
			matches := 0
			for _, n := range canvas.Nodes {
				if n.DomainID == rawFragment && n.ID != origin {
					t.Errorf("text %d raw fragment %s selects %s before decoding", ref, rawFragment, n.ID)
				}
				if n.DomainID == link.Fragment {
					matches++
					if n.ID != origin {
						t.Errorf("text %d URL %s points to %s, want %s", ref, link, n.ID, origin)
					}
				}
			}
			if matches != 1 {
				t.Errorf("text %d URL %s matches %d nodes", ref, link, matches)
			}
		}
		for stem := range canvases {
			if seen[stem] != 1 {
				t.Errorf("text %d has %d links for %s", ref, seen[stem], stem)
			}
		}
	}
}
