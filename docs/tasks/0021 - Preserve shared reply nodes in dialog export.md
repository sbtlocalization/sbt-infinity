---
epic:
parent:
status: Backlog
priority: 1
blocked by:
kind: Bug
mode: AFK
tier: high
---

# Sub-tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - parent == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      value.asFile().properties.status != 'Done' || acc,
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon('square-check-big'),
      if(formula.is_blocked,
        icon('construction'), 
        icon('square')
      )
    )
  progress_sort: |-
    if(status == "Done", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == 'In progress', 
          5, 
          0)))
  progress_string: |-
    if(status == "Done", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon('tally-' + priority), null)
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - file.name
      - formula.priority
      - status
      - blocked by
      - epic
    sort:
      - property: formula.progress_sort
        direction: ASC
      - property: priority
        direction: DESC
      - property: file.basename
        direction: ASC
    summaries: {}
    columnSize:
      formula.progress_icon: -56
      note.status: 125
      note.blocked by: 332

```
<!-- end unchanged block -->

# What to build

Fix dialog export when multiple entry nodes refer to the same reply node. The issue predates [[0020]] and was discovered during its review. Today each entry produces another copy of that reply with the same canvas `id` and `d-id`, so structural validation fails; after loop edges are removed, the remaining graph can still contain a cycle. `dialog list --with-length` can also disagree with the number of exported nodes, and text export can repeat the reply's context.

Reproduce with a synthetic DLG whose entry 0 and entry 1 both refer to reply 0, and whose reply 0 leads to entry 1. The reachable graph is `entry 0 → reply 0 → entry 1 → reply 0`. The last edge closes the cycle. This is a format-level fixture, not a claim that the example occurs in a particular game.

Export one node for each reachable entry and retained reply, and preserve every incoming edge to shared nodes. Classify edges through a deterministic depth-first traversal over both entry and reply nodes: an edge to a node on the current path closes a cycle, while an edge to an already expanded node outside that path is convergence. Classification is a static graph and layout hint, not a simulation of conditions or actions. Keep the existing meaning of conditions, actions, journal text, dialog ends, skipped empty replies, and parallel edges. Removing loop edges must leave the reachable exported graph acyclic without losing nodes or non-loop edges.

# Acceptance criteria

- [ ] A fixture with `entry 0 → reply 0 → entry 1 → reply 0` exports each node once, preserves all three edges, marks `entry 1 → reply 0` as loop, and passes dCanvas structural validation.
- [ ] A fixture where two different entry nodes converge on the same reply outside the current DFS path exports one reply with both incoming edges normal and included in layout.
- [ ] Cycles involving shared entry or reply nodes are classified according to traversal order; removing all loop edges leaves the reachable exported graph acyclic without losing nodes or non-loop edges.
- [ ] Retained replies preserve their text, actions, journal information and dialog-end semantics; conditions remain attached to the appropriate incoming edges, including shared and parallel edges.
- [ ] Skipped empty replies preserve the existing direct-edge behavior and conditions, including convergence, cycles and parallel edges.
- [ ] Every retained node has a unique canvas `id` and `d-id`, and text export provides the correct canvas-local context URL for each entry, reply and addressable journal action without duplicating context for the same node.
- [ ] `dialog list --with-length` counts each retained node once and agrees with the number of exported canvas nodes.
- [ ] Fixture-based tests exercise dialog export and text export for shared replies, including a cycle and convergence, and require neither an installed game nor a sibling repository.
