// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"iter"
)

func (d *Dialog) All() iter.Seq2[NodeOrigin, *Node] {
	return func(yield func(NodeOrigin, *Node) bool) {
		if d.RootState == nil {
			return
		}

		var dfs func(*Node) bool
		dfs = func(node *Node) bool {
			if !yield(node.Origin, node) {
				return false
			}

			for _, child := range node.Children {
				if !dfs(child) {
					return false
				}
			}

			return true
		}

		dfs(d.RootState)
	}
}
