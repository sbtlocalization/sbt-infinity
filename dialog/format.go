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
)

const (
	Width  = 400
	Height = 300
)

var (
	StateColor           = "3"
	TransitionColor      = "4"
	FinalTransitionColor = "2"
	LoopEdgeColor        = "6"
)

func (d *Dialog) ToJsonCanvas() *canvas.Canvas {
	c := canvas.NewCanvas()

	edges := make(map[string]*canvas.Edge)
	nodes := make(map[string]*canvas.Node)
	layoutEdges := make([][]string, 0)
	for dNodeOrigin, dNode := range d.All() {
		_ = dNodeOrigin
		cNode := newNode(dNode)
		if cNode != nil {
			c.AddNodes(cNode)
			nodes[cNode.ID] = cNode
		}

		if dNode.Parent != nil {
			cEdge := newEdge(dNode)
			if cEdge == nil {
				continue
			}

			if _, ok := edges[cEdge.ID]; !ok {
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
	)
	for _, n := range layout.Nodes {
		if cNode, ok := nodes[n.ID]; ok {
			cNode.X = int(n.X)
			cNode.Y = int(n.Y)
		} else {
			fmt.Printf("Warning: node %s not found in canvas nodes\n", n.ID)
		}
	}

	return c
}

func newNode(node *Node) *canvas.Node {
	cNode := canvas.Node{
		ID:     node.String(),
		X:      0,
		Y:      0,
		Width:  Width,
		Height: Height,
	}

	var sb strings.Builder
	name := node.Origin.String()
	if node.Type == TransitionNodeType {
		if node.Transition.IsDialogEnd {
			name = "Кінець діалогу " + name
		} else {
			name = "Відповідь " + name
		}
	}
	sb.WriteString(fmt.Sprintf("#### %s\n\n", name))
	switch node.Type {
	case StateNodeType:
		cNode.Color = &StateColor
		sb.WriteString(fmt.Sprintf("<small>Text **#%d**</small>\n\n", node.State.TextRef))
		sb.WriteString(node.State.Text)
	case TransitionNodeType:
		if node.Transition.IsDialogEnd {
			cNode.Color = &FinalTransitionColor
		} else {
			cNode.Color = &TransitionColor
		}

		if node.Transition.HasText {
			sb.WriteString(fmt.Sprintf("<small>Text **#%d**</small>\n\n", node.Transition.TextRef))
			sb.WriteString(node.Transition.Text)
		}
		if node.Transition.HasJournalText {
			if node.Transition.HasText {
				sb.WriteString("\n\n-----\n\n")
			}
			sb.WriteString(fmt.Sprintf("<small>Journal Text **#%d**</small>\n\n", node.Transition.TextRef))
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
	fromSide, toSide, toEnd := "bottom", "top", "arrow"
	var color string
	fromNode, toNode := node.Parent.String(), node.String()

	cEdge := &canvas.Edge{
		ID:       fmt.Sprintf("%s-%s", fromNode, toNode),
		FromNode: fromNode,
		FromSide: &fromSide,
		ToNode:   toNode,
		ToSide:   &toSide,
		ToEnd:    &toEnd,
		Color:    &color,
	}
	return cEdge
}
