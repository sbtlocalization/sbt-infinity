// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/sbtlocalization/dcanvas"
	"github.com/sbtlocalization/dcanvas/layout"
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

func validText(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return string([]rune(s))
}

func formatSound(name, prefix, suffix string) string {
	if name == "" {
		return ""
	}
	return validText(prefix + name + suffix)
}

func (d *Dialog) ToDCanvas(opts FormatOptions) (*dcanvas.Canvas, error) {
	c := &dcanvas.Canvas{Version: dcanvas.Version, Nodes: []*dcanvas.Node{}, Edges: []*dcanvas.Edge{}}
	stateColors := []string{"3", "6", "2", "5"}
	dlgNameToColor := make(map[string]string)
	errorNodes := make(map[string]bool)
	edgeIDs := make(map[string]bool)
	for _, dNode := range d.All() {
		if dNode.IsEmptyTransition() {
			continue
		}
		if dNode.Type == StateNodeType {
			if _, exists := dlgNameToColor[dNode.Origin.DlgName]; !exists {
				dlgNameToColor[dNode.Origin.DlgName] = stateColors[len(dlgNameToColor)%len(stateColors)]
			}
		}
		cNode, err := newNode(d, dNode, dlgNameToColor, opts)
		if err != nil {
			return nil, fmt.Errorf("format node %s: %w", dNode.Origin, err)
		}
		if cNode != nil && (dNode.Type != ErrorNodeType || !errorNodes[cNode.ID]) {
			c.Nodes = append(c.Nodes, cNode)
			if dNode.Type == ErrorNodeType {
				errorNodes[cNode.ID] = true
			}
		}
		if edge := newEdge(dNode); edge != nil {
			baseID := edge.ID
			for suffix := 2; edgeIDs[edge.ID]; suffix++ {
				edge.ID = fmt.Sprintf("%s#%d", baseID, suffix)
			}
			edgeIDs[edge.ID] = true
			c.Edges = append(c.Edges, edge)
		}
	}

	applyLayout(c, d.Id, func(c *dcanvas.Canvas) error { return layout.Layout(c) })
	return c, nil
}

func applyLayout(c *dcanvas.Canvas, id DialogID, layoutFn func(*dcanvas.Canvas) error) {
	positions := make([][2]int, len(c.Nodes))
	for i, n := range c.Nodes {
		positions[i] = [2]int{n.X, n.Y}
	}
	if err := layoutFn(c); err != nil {
		for i, n := range c.Nodes {
			n.X, n.Y = positions[i][0], positions[i][1]
		}
		fmt.Printf("warning(%s): layout algorithm failed: %v\nNodes will be placed at default positions\n", id, err)
	}
}

func newNode(d *Dialog, node *Node, colors map[string]string, opts FormatOptions) (*dcanvas.Node, error) {
	cNode := &dcanvas.Node{
		ID: validText(node.String()), Type: "text", Width: Width, Height: Height, DomainID: validText(node.Origin.String()),
	}
	if node.Type == StateNodeType || node.Type == TransitionNodeType {
		cNode.DomainID = node.domainID(d.Id)
	}
	switch node.Type {
	case StateNodeType:
		cNode.Kind = dcanvas.KindLine
		cNode.TextID = fmt.Sprintf("#%d", node.State.TextRef)
		cNode.Sound = formatSound(node.State.Sound, opts.SoundPrefix, opts.SoundSuffix)
		cNode.Condition = validText(strings.TrimSpace(node.State.Trigger))
		cNode.Text = validText(node.State.Text)
		creatures := slices.Clone(d.AllCreatures[node.Origin.DlgName])
		slices.SortFunc(creatures, func(a, b *Creature) int {
			return strings.Compare(strings.ToUpper(a.ResourceName), strings.ToUpper(b.ResourceName))
		})
		if len(creatures) > 0 {
			cNode.Character = newCharacter(creatures[0])
			if cNode.Character != nil {
				for _, cre := range creatures[1:] {
					if character := newCharacter(cre); character != nil {
						cNode.AlternativeCharacters = append(cNode.AlternativeCharacters, character)
					}
				}
			}
		}
		cNode.Color = colors[node.Origin.DlgName]
		if cNode.Color == "" {
			cNode.Color = StateColor
		}
	case TransitionNodeType:
		cNode.Kind = dcanvas.KindReply
		if node.Transition.HasText {
			cNode.TextID = fmt.Sprintf("#%d", node.Transition.TextRef)
			cNode.Text = validText(node.Transition.Text)
			cNode.Sound = formatSound(node.Transition.Sound, opts.SoundPrefix, opts.SoundSuffix)
		}
		if node.Transition.HasJournalText {
			for _, field := range []struct{ key, value string }{
				{"x-journalTextId", fmt.Sprintf("#%d", node.Transition.JournalTextRef)},
				{"x-journalText", validText(node.Transition.JournalText)},
				{"x-journalSound", formatSound(node.Transition.JournalSound, opts.SoundPrefix, opts.SoundSuffix)},
			} {
				if field.value != "" {
					if err := cNode.SetExtra(field.key, field.value); err != nil {
						return nil, err
					}
				}
			}
		}
		cNode.Action = validText(strings.TrimSpace(node.Transition.Action))
		cNode.Color = TransitionColor
		if node.Transition.IsDialogEnd {
			cNode.Role = "dialog-end"
			cNode.Color = FinalTransitionColor
		}
	case ErrorNodeType:
		cNode.Role = "error"
		cNode.Color = FinalTransitionColor
		cNode.Text = "**Error loading state**"
	case LoopNodeType, ConvergenceNodeType:
		return nil, nil
	}
	return cNode, nil
}

func newEdge(node *Node) *dcanvas.Edge {
	if node.Parent == nil {
		return nil
	}
	fromNode, toNode := validText(node.Parent.String()), validText(node.String())
	var condition string
	if node.Type == TransitionNodeType && node.Transition.HasTrigger {
		condition = validText(strings.TrimSpace(node.Transition.Trigger))
	}
	if node.Parent.IsEmptyTransition() && node.Parent.Parent != nil {
		fromNode = validText(node.Parent.Parent.String())
		if node.Parent.Transition.HasTrigger {
			condition = validText(strings.TrimSpace(node.Parent.Transition.Trigger))
		}
	}
	kind := dcanvas.KindNormal
	if node.Type == LoopNodeType {
		kind = dcanvas.KindLoop
	}
	return &dcanvas.Edge{
		ID: fmt.Sprintf("%s-%s", fromNode, toNode), FromNode: fromNode, FromSide: "bottom",
		ToNode: toNode, ToSide: "top", ToEnd: "arrow", Condition: condition, Kind: kind,
	}
}

func newCharacter(cre *Creature) *dcanvas.Character {
	name, ref, hasRef := cre.LongName, cre.LongNameId, cre.HasLongNameId
	if name == "" {
		name, ref, hasRef = cre.ShortName, cre.ShortNameId, cre.HasShortNameId
	}
	if name == "" {
		name, hasRef = strings.TrimSuffix(cre.ResourceName, ".CRE"), false
	}
	if name == "" {
		return nil
	}
	ch := &dcanvas.Character{Name: validText(name), Gender: validText(cre.Sex)}
	if hasRef {
		ch.TextID = fmt.Sprintf("#%d", ref)
	}
	portrait := strings.TrimSuffix(validText(cre.Portrait), ".BMP")
	if portrait != "" && portrait != "None" {
		ch.Portrait = path.Join("portraits", portrait+".png")
	}
	return ch
}

func (id NodeOrigin) CanvasStem() string {
	return validText(fmt.Sprintf("%s-%d", id.DlgName, id.Index))
}

func (n *Node) domainID(root DialogID) string {
	if strings.EqualFold(n.Origin.DlgName, root.DlgName) {
		return fmt.Sprintf("%s-%d", n.Type, n.Origin.Index)
	}
	return validText(fmt.Sprintf("%s-%s-%d", n.Type, n.Origin.DlgName, n.Origin.Index))
}

func (n *Node) ToUrl(baseUrl string) string {
	id := n.domainID(n.Dialog.Id)
	fragment := url.PathEscape(id)
	if fragment != id {
		// The viewer tries the raw fragment before decoding it; prevent an escaped
		// identifier from accidentally matching another node's literal identifier.
		fragment = strings.Replace(fragment, "-", "%2D", 1)
	}
	return fmt.Sprintf("%s/dialog/%s#%s", strings.TrimRight(baseUrl, "/"),
		url.PathEscape(n.Dialog.Id.CanvasStem()), fragment)
}
