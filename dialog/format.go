// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"strings"

	"github.com/nulab/autog"
	"github.com/nulab/autog/graph"
	"github.com/supersonicpineapple/go-jsoncanvas/canvas"
	"go.yaml.in/yaml/v3"
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

type Character struct {
	Name     string `yaml:"name"`
	Portrait string `yaml:"portrait"`
}

type FrontMatter struct {
	NodeId        string     `yaml:"nodeId"`
	Character     *Character `yaml:"character,omitempty"`
	TextID        string     `yaml:"textId,omitempty"`
	JournalTextID string     `yaml:"journalTextId,omitempty"`
	Trigger       string     `yaml:"trigger,omitempty"`
	Action        string     `yaml:"action,omitempty"`
}

func NewFrontMatter(nodeId string) *FrontMatter {
	return &FrontMatter{
		NodeId: nodeId,
	}
}

func (fm *FrontMatter) SetCharacter(name, portrait string) {
	if fm.Character == nil {
		fm.Character = &Character{}
	}
	fm.Character.Name = name
	fm.Character.Portrait = strings.Replace(portrait, ".BMP", ".png", 1)
}

func isEmptyTransitionNode(node *Node) bool {
	return node.Type == TransitionNodeType &&
		!node.Transition.IsDialogEnd &&
		!node.Transition.HasText &&
		!node.Transition.HasJournalText &&
		!node.Transition.HasTrigger &&
		!node.Transition.HasAction
}

func (d *Dialog) ToJsonCanvas() *canvas.Canvas {
	c := canvas.NewCanvas()

	// Create color mapping for unique DlgName values (built on-demand)
	stateColors := []string{"3", "6", "2", "5"}
	dlgNameToColor := make(map[string]string)
	colorIndex := 0

	edges := make(map[string]*canvas.Edge)
	nodes := make(map[string]*canvas.Node)
	layoutEdges := make([][]string, 0)
	for _, dNode := range d.All() {
		if isEmptyTransitionNode(dNode) {
			continue
		} else {
			// Build color mapping on-demand for StateNodeType
			if dNode.Type == StateNodeType {
				if _, exists := dlgNameToColor[dNode.Origin.DlgName]; !exists {
					dlgNameToColor[dNode.Origin.DlgName] = stateColors[colorIndex%len(stateColors)]
					colorIndex++
				}
			}

			cNode := newNode(d, dNode, dlgNameToColor)
			if cNode != nil {
				c.AddNodes(cNode)
				nodes[cNode.ID] = cNode
			}

			if dNode.Parent != nil {
				cEdge := newEdge(dNode)
				if cEdge == nil {
					continue
				}

				edges[cEdge.ID] = cEdge
				layoutEdges = append(layoutEdges, []string{cEdge.FromNode, cEdge.ToNode})
				c.AddEdges(cEdge)
			}
		}
	}

	src := graph.EdgeSlice(layoutEdges)
	layout := autog.Layout(
		src,
		autog.WithNodeFixedSize(Width, Height),
		autog.WithLayerSpacing(200),
	)
	for _, n := range layout.Nodes {
		if cNode, ok := nodes[n.ID]; ok {
			cNode.X = int(n.X)
			cNode.Y = int(n.Y)
		} else {
			fmt.Printf("warning: node %s not found in canvas nodes\n", n.ID)
		}
	}

	return c
}

func newNode(d *Dialog, node *Node, dlgNameToColor map[string]string) *canvas.Node {
	cNode := canvas.Node{
		ID:     node.String(),
		X:      0,
		Y:      0,
		Width:  Width,
		Height: Height,
	}

	var sb strings.Builder

	// front matter
	sb.WriteString("---\n")
	fm := NewFrontMatter(node.Origin.String())
	switch node.Type {
	case StateNodeType:
		fm.TextID = fmt.Sprintf("#%d", node.State.TextRef)
		fm.Trigger = strings.TrimSpace(node.State.Trigger)
		if cre, ok := d.AllCreatures[node.Origin.DlgName]; ok {
			fm.SetCharacter(cre.LongName, cre.Portrait)
		}
	case TransitionNodeType:
		if node.Transition.HasText {
			fm.TextID = fmt.Sprintf("#%d", node.Transition.TextRef)
		}
		if node.Transition.HasJournalText {
			fm.JournalTextID = fmt.Sprintf("#%d", node.Transition.JournalTextRef)
		}
		fm.Action = strings.TrimSpace(node.Transition.Action)
		if node.Transition.IsDialogEnd {
			fm.SetCharacter("End dialog", "")
		} else {
			fm.SetCharacter("Answer", "")
		}
	}

	fms, _ := yaml.Marshal(fm)
	sb.Write(fms)
	sb.WriteString("---\n")

	// content
	switch node.Type {
	case StateNodeType:
		if color, exists := dlgNameToColor[node.Origin.DlgName]; exists {
			cNode.Color = &color
		} else {
			cNode.Color = &StateColor
		}
		sb.WriteString(node.State.Text)
	case TransitionNodeType:
		if node.Transition.IsDialogEnd {
			cNode.Color = &FinalTransitionColor
		} else {
			cNode.Color = &TransitionColor
		}

		if node.Transition.HasText {
			sb.WriteString(node.Transition.Text)
		}

		if node.Transition.HasJournalText {
			sb.WriteString("\n\n>---- JOURNAL ----<\n\n")
			sb.WriteString(node.Transition.JournalText)
		}
	case ErrorNodeType:
		cNode.Color = &FinalTransitionColor
		sb.WriteString("**Error loading state**")
	case LoopNodeType:
		return nil
	}
	cNode.SetText(sb.String())

	return &cNode
}

func newEdge(node *Node) *canvas.Edge {
	if node.Parent == nil {
		return nil
	}

	fromNode, toNode := node.Parent.String(), node.String()
	fromSide, toSide, toEnd := "bottom", "top", "arrow"
	var color string

	if isEmptyTransitionNode(node.Parent) && node.Parent.Parent != nil {
		// skip empty transition nodes, connect parent state to next state directly
		fromNode = node.Parent.Parent.String()
	}

	cEdge := &canvas.Edge{
		ID:       fmt.Sprintf("%s-%s", fromNode, toNode),
		FromNode: fromNode,
		FromSide: &fromSide,
		ToNode:   toNode,
		ToSide:   &toSide,
		ToEnd:    &toEnd,
		Color:    &color,
	}

	var triggerText string
	if node.Type == TransitionNodeType && node.Transition.HasTrigger {
		triggerText = strings.TrimSpace(node.Transition.Trigger)
	}

	if triggerText != "" {
		cEdge.Label = &triggerText
	}

	return cEdge
}
