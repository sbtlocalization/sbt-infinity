// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"iter"
)

type nodeId struct {
	Type   NodeType
	Origin NodeOrigin
}

func (d *Dialog) All() iter.Seq2[NodeOrigin, *Node] {
	return func(yield func(NodeOrigin, *Node) bool) {
		if d.RootState == nil {
			return
		}

		// Stack for DFS traversal
		stack := []*Node{d.RootState}
		visited := make(map[nodeId]bool)

		for len(stack) > 0 {
			// Pop from stack
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			currentId := nodeId{Type: current.Type, Origin: current.Origin}

			if visited[currentId] {
				continue
			}
			visited[currentId] = true

			if !yield(current.Origin, current) {
				return
			}

			// Add children to stack in reverse order for left-to-right DFS
			for i := len(current.Children) - 1; i >= 0; i-- {
				nextId := nodeId{Type: current.Children[i].Type, Origin: current.Children[i].Origin}
				if !visited[nextId] {
					stack = append(stack, current.Children[i])
				}
			}
		}
	}
}
