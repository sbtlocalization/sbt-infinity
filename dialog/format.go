// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"path"
	"strings"

	"github.com/nulab/autog"
	"github.com/nulab/autog/graph"
	"github.com/sbtlocalization/sbt-infinity/dcanvas"
)

const (
	Width  = 400
	Height = 300
)

var (
	StateColor           = "3"
	TransitionColor      = "4"
	FinalTransitionColor = "1"
)

type FormatOptions struct {
	SoundPrefix string
	SoundSuffix string
}

func formatSound(name, prefix, suffix string) string {
	if name == "" {
		return ""
	}
	return prefix + name + suffix
}

func (d *Dialog) ToDCanvas(opts FormatOptions) *dcanvas.Canvas {
	c := &dcanvas.Canvas{
		Version: "2.0",
	}

	// Create color mapping for unique DlgName values (built on-demand)
	stateColors := []string{"3", "6", "2", "5"}
	dlgNameToColor := make(map[string]string)
	colorIndex := 0

	edges := make(map[string]*dcanvas.Edge)
	nodes := make(map[string]*dcanvas.Node)
	layoutEdges := make([][]string, 0)
	for _, dNode := range d.All() {
		if dNode.IsEmptyTransition() {
			continue
		} else {
			// Build color mapping on-demand for StateNodeType
			if dNode.Type == StateNodeType {
				if _, exists := dlgNameToColor[dNode.Origin.DlgName]; !exists {
					dlgNameToColor[dNode.Origin.DlgName] = stateColors[colorIndex%len(stateColors)]
					colorIndex++
				}
			}

			cNode := newNode(d, dNode, dlgNameToColor, opts)
			if cNode != nil {
				c.Nodes = append(c.Nodes, cNode)
				nodes[cNode.ID] = cNode
			}

			if dNode.Parent != nil {
				cEdge, loop := newEdge(dNode)
				if cEdge == nil {
					continue
				}

				edges[cEdge.ID] = cEdge
				if !loop {
					layoutEdges = append(layoutEdges, []string{cEdge.FromNode, cEdge.ToNode})
				}
				c.Edges = append(c.Edges, cEdge)
			}
		}
	}

	// Validate graph structure before layout
	if len(nodes) == 0 {
		fmt.Printf("warning(%s): no nodes to layout, skipping autolayout", d.Id)
		return c
	}

	// Add panic recovery for the autog.Layout call
	var layout graph.Layout
	layoutSuccess := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("warning(%s): layout algorithm failed: %v\n", d.Id, r)
				fmt.Println("Nodes will be placed at default positions")
				layoutSuccess = false
			}
		}()

		src := graph.EdgeSlice(layoutEdges)
		layout = autog.Layout(
			src,
			autog.WithNodeFixedSize(Width, Height),
			autog.WithLayerSpacing(200),
		)
		layoutSuccess = true
	}()

	// Apply layout if successful
	if layoutSuccess {
		for _, n := range layout.Nodes {
			if cNode, ok := nodes[n.ID]; ok {
				cNode.X = int(n.X)
				cNode.Y = int(n.Y)
			} else {
				fmt.Printf("warning(%s): node %s not found in canvas nodes\n", d.Id, n.ID)
			}
		}
	}

	return c
}

func newNode(d *Dialog, node *Node, dlgNameToColor map[string]string, opts FormatOptions) *dcanvas.Node {
	cNode := &dcanvas.Node{
		ID:     node.String(),
		Type:   "text",
		X:      0,
		Y:      0,
		Width:  Width,
		Height: Height,
		NodeId: node.Origin.String(),
	}

	switch node.Type {
	case StateNodeType:
		cNode.NodeRole = "state"
		cNode.TextId = fmt.Sprintf("#%d", node.State.TextRef)
		cNode.Sound = formatSound(node.State.Sound, opts.SoundPrefix, opts.SoundSuffix)
		cNode.Trigger = strings.TrimSpace(node.State.Trigger)
		cNode.Text = node.State.Text

		if cre, ok := d.AllCreatures[node.Origin.DlgName]; ok {
			cNode.Character = newCharacter(cre.LongName, cre.Portrait)
		}

		if color, exists := dlgNameToColor[node.Origin.DlgName]; exists {
			cNode.Color = color
		} else {
			cNode.Color = StateColor
		}

	case TransitionNodeType:
		cNode.NodeRole = "transition"
		if node.Transition.HasText {
			cNode.TextId = fmt.Sprintf("#%d", node.Transition.TextRef)
			cNode.Text = node.Transition.Text
			cNode.Sound = formatSound(node.Transition.Sound, opts.SoundPrefix, opts.SoundSuffix)
		}
		if node.Transition.HasJournalText {
			cNode.JournalTextId = fmt.Sprintf("#%d", node.Transition.JournalTextRef)
			cNode.JournalText = node.Transition.JournalText
			cNode.JournalSound = formatSound(node.Transition.JournalSound, opts.SoundPrefix, opts.SoundSuffix)
		}
		cNode.Action = strings.TrimSpace(node.Transition.Action)

		if node.Transition.IsDialogEnd {
			cNode.Character = newCharacter("End dialog", "")
			cNode.Color = FinalTransitionColor
		} else {
			cNode.Character = newCharacter("Answer", "")
			cNode.Color = TransitionColor
		}

	case ErrorNodeType:
		cNode.NodeRole = "state"
		cNode.Color = FinalTransitionColor
		cNode.Text = "**Error loading state**"

	case LoopNodeType:
		return nil
	}

	return cNode
}

func newEdge(node *Node) (*dcanvas.Edge, bool) {
	if node.Parent == nil {
		return nil, false
	}

	loop := node.Type == LoopNodeType

	fromNode, toNode := node.Parent.String(), node.String()

	var condition string
	if node.Type == TransitionNodeType && node.Transition.HasTrigger {
		condition = strings.TrimSpace(node.Transition.Trigger)
	}

	if node.Parent.IsEmptyTransition() && node.Parent.Parent != nil {
		// skip empty transition nodes, connect parent state to next state directly
		fromNode = node.Parent.Parent.String()

		if node.Parent.Type == TransitionNodeType && node.Parent.Transition.HasTrigger {
			condition = strings.TrimSpace(node.Parent.Transition.Trigger)
		}
	}

	cEdge := &dcanvas.Edge{
		ID:        fmt.Sprintf("%s-%s", fromNode, toNode),
		FromNode:  fromNode,
		FromSide:  "bottom",
		ToNode:    toNode,
		ToSide:    "top",
		ToEnd:     "arrow",
		Condition: condition,
	}

	return cEdge, loop
}

func newCharacter(name, portrait string) *dcanvas.Character {
	ch := &dcanvas.Character{Name: name}
	portrait = strings.TrimSuffix(portrait, ".BMP")
	if portrait != "" && portrait != "None" {
		ch.Portrait = path.Join("portraits", portrait + ".png")
	}
	return ch
}

func (n *Node) ToUrl(baseUrl string) string {
	dialogName := strings.TrimSuffix(strings.ToLower(n.Dialog.Id.DlgName), ".dlg")
	dialogId := fmt.Sprintf("%s-%d", dialogName, n.Dialog.Id.Index)
	return fmt.Sprintf(
		"%s/dialog/%s#%s-%s-%d-",
		strings.TrimRight(baseUrl, "/"),
		dialogId,
		n.Type,
		strings.ToLower(n.Origin.DlgName),
		n.Origin.Index,
	)
}
