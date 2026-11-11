// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-License-Identifier: GPL-3.0-only

package dialog_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
	"github.com/sbtlocalization/sbt-infinity/dialog"
)

func TestExportReusesErrorNodeForRepeatedMissingState(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	root := &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, State: &dialog.StateData{Text: "Hello"}, Dialog: d}
	d.RootState = root
	for i := range uint32(2) {
		reply := &dialog.Node{Type: dialog.TransitionNodeType, Origin: dialog.NewNodeOrigin("TEST", i), Parent: root, Dialog: d, Transition: &dialog.TransitionData{HasText: true, Text: "Reply"}}
		missing := &dialog.Node{Type: dialog.ErrorNodeType, Origin: dialog.NewNodeOrigin("MISSING", 0), Parent: reply, Dialog: d}
		reply.Children = []*dialog.Node{missing}
		root.Children = append(root.Children, reply)
	}
	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := dcanvas.Encode(canvas, &out); err != nil {
		t.Fatal(err)
	}
	decoded, err := dcanvas.Decode(&out)
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(decoded); err != nil {
		t.Fatal(err)
	}
	count, incoming := 0, 0
	for _, n := range decoded.Nodes {
		if n.ID == "error-MISSING[0]" {
			count++
			if n.Role != "error" || n.DomainID != "MISSING[0]" {
				t.Errorf("error node: %+v", n)
			}
		}
	}
	for _, e := range decoded.Edges {
		if e.ToNode == "error-MISSING[0]" {
			incoming++
		}
	}
	if count != 1 || incoming != 2 {
		t.Errorf("error nodes = %d, incoming edges = %d", count, incoming)
	}
	if got := d.NodeCount(); got != len(decoded.Nodes) {
		t.Errorf("node count = %d, want %d", got, len(decoded.Nodes))
	}
}

func TestExportParallelSkippedTransitionsKeepDistinctEdges(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	root := &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, State: &dialog.StateData{Text: "Hello"}, Dialog: d}
	d.RootState = root
	for i, condition := range []string{"First()", "Second()", "Third()"} {
		tr := &dialog.Node{Type: dialog.TransitionNodeType, Origin: dialog.NewNodeOrigin("TEST", uint32(i)), Parent: root, Dialog: d, Transition: &dialog.TransitionData{HasTrigger: true, Trigger: condition}}
		kind := dialog.LoopNodeType
		if i == 0 {
			kind = dialog.StateNodeType
		}
		target := &dialog.Node{Type: kind, Origin: dialog.NewNodeOrigin("TEST", 1), Parent: tr, Dialog: d, State: &dialog.StateData{Text: "Next"}}
		tr.Children = []*dialog.Node{target}
		root.Children = append(root.Children, tr)
	}
	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := dcanvas.Encode(canvas, &out); err != nil {
		t.Fatal(err)
	}
	decoded, err := dcanvas.Decode(&out)
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(decoded); err != nil {
		t.Fatal(err)
	}
	ids, conditions := map[string]bool{}, map[string]bool{}
	for _, e := range decoded.Edges {
		if e.FromNode != "state-TEST[0]" || e.ToNode != "state-TEST[1]" {
			t.Fatalf("edge endpoints: %+v", e)
		}
		if ids[e.ID] {
			t.Errorf("duplicate edge ID %q", e.ID)
		}
		ids[e.ID], conditions[e.Condition] = true, true
	}
	if len(ids) != 3 || !ids["state-TEST[0]-state-TEST[1]"] || len(conditions) != 3 || !conditions["First()"] || !conditions["Second()"] || !conditions["Third()"] {
		t.Errorf("edges = %+v", decoded.Edges)
	}
}

func TestExportExcludesConvergenceFromLayoutButKeepsEdges(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	root := &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, State: &dialog.StateData{Text: "Root"}, Dialog: d}
	d.RootState = root
	attach := func(parent *dialog.Node, n *dialog.Node) *dialog.Node {
		n.Parent, n.Dialog = parent, d
		parent.Children = append(parent.Children, n)
		return n
	}
	transition := func(index uint32, data dialog.TransitionData) *dialog.Node {
		return &dialog.Node{Type: dialog.TransitionNodeType, Origin: dialog.NewNodeOrigin("TEST", index), Transition: &data}
	}
	state := func(index uint32, kind dialog.NodeType) *dialog.Node {
		return &dialog.Node{Type: kind, Origin: dialog.NewNodeOrigin("TEST", index), State: &dialog.StateData{Text: "State"}}
	}
	first := attach(root, transition(0, dialog.TransitionData{}))
	target := attach(first, state(1, dialog.StateNodeType))
	loop := attach(target, transition(1, dialog.TransitionData{HasTrigger: true, Trigger: "Loop()"}))
	attach(loop, state(0, dialog.LoopNodeType))
	reply := attach(root, transition(2, dialog.TransitionData{HasText: true, Text: "Reply"}))
	attach(reply, state(1, dialog.ConvergenceNodeType))
	parallel := attach(root, transition(3, dialog.TransitionData{HasTrigger: true, Trigger: "Parallel()"}))
	attach(parallel, state(1, dialog.ConvergenceNodeType))

	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := dcanvas.Encode(canvas, &encoded); err != nil {
		t.Fatal(err)
	}
	decoded, err := dcanvas.Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(decoded); err != nil {
		t.Fatal(err)
	}
	layoutGraph := &dcanvas.Canvas{}
	for _, n := range decoded.Nodes {
		copy := *n
		copy.X, copy.Y = 0, 0
		layoutGraph.Nodes = append(layoutGraph.Nodes, &copy)
	}
	var convergences int
	for _, e := range decoded.Edges {
		convergence := e.FromNode == "transition-TEST[2]" && e.ToNode == "state-TEST[1]" || e.Condition == "Parallel()"
		if convergence {
			convergences++
			if e.Kind != dcanvas.KindNormal {
				t.Errorf("convergence edge has kind %q", e.Kind)
			}
			continue
		}
		if e.Kind != dcanvas.KindLoop {
			layoutGraph.Edges = append(layoutGraph.Edges, e)
		}
	}
	if convergences != 2 || len(decoded.Edges) != 5 {
		t.Fatalf("convergences = %d, edges = %d", convergences, len(decoded.Edges))
	}
	if err := layout.Layout(layoutGraph); err != nil {
		t.Fatal(err)
	}
	for i, n := range decoded.Nodes {
		if n.X != layoutGraph.Nodes[i].X || n.Y != layoutGraph.Nodes[i].Y {
			t.Errorf("loop or convergence influenced layout of %s", n.ID)
		}
	}
}

func TestExportReplacesInvalidBytesInJournalText(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	d.RootState = &dialog.Node{Type: dialog.TransitionNodeType, Origin: d.Id, Transition: &dialog.TransitionData{HasJournalText: true, JournalText: string([]byte{0xff, 0xfe})}}
	canvas, err := d.ToDCanvas(dialog.FormatOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := dcanvas.Encode(canvas, &out); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.Nodes[0]["x-journalText"]; got != "\ufffd\ufffd" {
		t.Fatalf("journal text = %q", got)
	}
}

func TestExportGraphAndJournalRoundTrip(t *testing.T) {
	d := dialog.NewDialog(dialog.NewNodeOrigin("TEST", 0))
	root := &dialog.Node{Type: dialog.StateNodeType, Origin: d.Id, Dialog: d, State: &dialog.StateData{TextRef: 4, Text: "Line"}}
	d.RootState = root
	attach := func(parent *dialog.Node, n *dialog.Node) *dialog.Node {
		n.Parent = parent
		n.Dialog = d
		parent.Children = append(parent.Children, n)
		return n
	}
	transition := func(index uint32, data dialog.TransitionData) *dialog.Node {
		return &dialog.Node{Type: dialog.TransitionNodeType, Origin: dialog.NewNodeOrigin("TEST", index), Transition: &data}
	}
	skipped := attach(root, transition(1, dialog.TransitionData{HasTrigger: true, Trigger: "  Check()  "}))
	next := attach(skipped, &dialog.Node{Type: dialog.StateNodeType, Origin: dialog.NewNodeOrigin("TEST", 1), State: &dialog.StateData{TextRef: 5, Text: "Next"}})
	loopReply := attach(next, transition(2, dialog.TransitionData{HasText: true, TextRef: 6, Text: "Again", Sound: "REPLY", HasTrigger: true, Trigger: "ReplyCheck()"}))
	attach(loopReply, &dialog.Node{Type: dialog.LoopNodeType, Origin: d.Id})
	skippedLoop := attach(next, transition(6, dialog.TransitionData{HasTrigger: true, Trigger: "  LoopCheck()  "}))
	attach(skippedLoop, &dialog.Node{Type: dialog.LoopNodeType, Origin: d.Id})
	action := attach(root, transition(3, dialog.TransitionData{HasAction: true, Action: " Do() "}))
	attach(action, &dialog.Node{Type: dialog.ErrorNodeType, Origin: dialog.NewNodeOrigin("MISSING", 7)})
	journal := attach(root, transition(4, dialog.TransitionData{HasJournalText: true, JournalTextRef: 10, JournalText: "Journal", JournalSound: "JOURNAL"}))
	attach(journal, &dialog.Node{Type: dialog.ErrorNodeType, Origin: dialog.NewNodeOrigin("MISSING", 8)})
	attach(root, transition(5, dialog.TransitionData{IsDialogEnd: true}))

	canvas, err := d.ToDCanvas(dialog.FormatOptions{SoundPrefix: "snd/", SoundSuffix: ".ogg"})
	if err != nil {
		t.Fatal(err)
	}
	withoutLoops := &dcanvas.Canvas{}
	for _, n := range canvas.Nodes {
		copy := *n
		copy.X, copy.Y = 0, 0
		withoutLoops.Nodes = append(withoutLoops.Nodes, &copy)
	}
	for _, e := range canvas.Edges {
		if e.Kind != dcanvas.KindLoop {
			withoutLoops.Edges = append(withoutLoops.Edges, e)
		}
	}
	if err := layout.Layout(withoutLoops); err != nil {
		t.Fatal(err)
	}
	for i, n := range canvas.Nodes {
		if n.X != withoutLoops.Nodes[i].X || n.Y != withoutLoops.Nodes[i].Y {
			t.Errorf("loop changed position of %s", n.ID)
		}
	}
	if got := root.ToUrl("https://example.test/"); got != "https://example.test/dialog/TEST-0#state-0" {
		t.Errorf("URL: %s", got)
	}
	var buf bytes.Buffer
	if err := dcanvas.Encode(canvas, &buf); err != nil {
		t.Fatal(err)
	}
	first := append([]byte(nil), buf.Bytes()...)
	assertSchema(t, first)
	decoded, err := dcanvas.Decode(bytes.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	if err := dcanvas.Validate(decoded); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err := dcanvas.Encode(decoded, &buf); err != nil {
		t.Fatal(err)
	}
	assertSchema(t, buf.Bytes())
	var before, after map[string]any
	if err := json.Unmarshal(first, &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(buf.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("round trip changed document: %s\n%s", first, buf.Bytes())
	}

	nodes := map[string]map[string]any{}
	for _, n := range after["nodes"].([]any) {
		v := n.(map[string]any)
		nodes[v["id"].(string)] = v
	}
	if _, ok := nodes["transition-TEST[1]"]; ok {
		t.Error("empty transition retained")
	}
	if got := nodes["transition-TEST[2]"]; got["d-kind"] != "reply" || got["text"] != "Again" || got["d-textId"] != "#6" || got["d-sound"] != "snd/REPLY.ogg" {
		t.Errorf("reply: %v", got)
	}
	if got := nodes["transition-TEST[3]"]; got["d-kind"] != "reply" || got["text"] != "" || got["d-action"] != "Do()" {
		t.Errorf("action-only reply: %v", got)
	}
	if got := nodes["transition-TEST[4]"]; got["d-kind"] != "reply" || got["d-role"] != nil || got["x-journalText"] != "Journal" || got["x-journalTextId"] != "#10" || got["x-journalSound"] != "snd/JOURNAL.ogg" {
		t.Errorf("journal: %v", got)
	}
	if got := nodes["transition-TEST[5]"]; got["d-kind"] != "reply" || got["d-role"] != "dialog-end" || got["text"] != "" {
		t.Errorf("terminal: %v", got)
	}
	if got := nodes["error-MISSING[7]"]; got["d-role"] != "error" || got["d-kind"] != nil || got["text"] != "**Error loading state**" || got["color"] != "1" {
		t.Errorf("error: %v", got)
	}
	for _, n := range nodes {
		for k := range n {
			if strings.HasPrefix(k, "x-") && k != "x-journalText" && k != "x-journalTextId" && k != "x-journalSound" {
				t.Errorf("obsolete field %s", k)
			}
		}
		if n["width"] != float64(400) || n["height"] != float64(300) {
			t.Errorf("node size: %v", n)
		}
		if n["d-role"] == "state" || n["d-role"] == "transition" || n["d-character"] != nil {
			t.Errorf("unexpected role or character: %v", n)
		}
	}
	var direct, loop, skippedLoopEdge, replyCondition bool
	for _, v := range after["edges"].([]any) {
		e := v.(map[string]any)
		for k := range e {
			if strings.HasPrefix(k, "x-") {
				t.Errorf("obsolete edge field %s", k)
			}
		}
		if e["label"] != nil || e["d-textId"] != nil {
			t.Errorf("edge contains text: %v", e)
		}
		if e["fromNode"] == "state-TEST[0]" && e["toNode"] == "state-TEST[1]" {
			direct = e["d-kind"] == "normal" && e["d-condition"] == "Check()"
		}
		if e["fromNode"] == "transition-TEST[2]" && e["toNode"] == "state-TEST[0]" {
			loop = e["d-kind"] == "loop"
		}
		if e["fromNode"] == "state-TEST[1]" && e["toNode"] == "state-TEST[0]" {
			skippedLoopEdge = e["d-kind"] == "loop" && e["d-condition"] == "LoopCheck()"
		}
		if e["toNode"] == "transition-TEST[2]" {
			replyCondition = e["d-condition"] == "ReplyCheck()" && e["d-kind"] == "normal"
		}
	}
	if !direct || !loop || !skippedLoopEdge || !replyCondition {
		t.Errorf("direct=%v loop=%v skippedLoop=%v replyCondition=%v", direct, loop, skippedLoopEdge, replyCondition)
	}
	if nodes["state-TEST[0]"]["y"].(float64) >= nodes["state-TEST[1]"]["y"].(float64) {
		t.Error("loop affected top-down layout")
	}
}

func assertSchema(t *testing.T, data []byte) {
	t.Helper()
	file, err := os.Open("testdata/dCanvas-1.1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("dCanvas-1.1.schema.json", file); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("dCanvas-1.1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(document); err != nil {
		t.Fatalf("schema: %v\n%s", err, data)
	}
}
