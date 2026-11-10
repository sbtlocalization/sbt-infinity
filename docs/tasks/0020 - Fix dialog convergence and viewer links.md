---
epic:
parent:
status: Done
priority: 1
blocked by:
kind: Bug
mode: AFK
tier: mid
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

Fix `dialog export` so converging dialog paths are not mistaken for loops, and make the links in `text export` point to the corresponding exported nodes in the current `sbt-dialog-viewer` `v1.0` branch. The viewer itself must not be changed.

Today, a state is treated as a loop whenever it has already been visited anywhere in the dialog. For example, with `A → B → D` and `A → C → D`, the second incoming edge to `D` is incorrectly marked as a loop and excluded from layout. A loop is instead an edge to a state on the current depth-first traversal path; a later edge to an already expanded state outside that path is convergence. Keep one exported node for the state and preserve all incoming edges. Loop classification is a static graph and layout hint, not a simulation of the player's history: do not evaluate conditions or actions to determine whether a cycle is playable. Which edge closes a cycle can depend on the deterministic traversal order.

The `text export` links currently use lowercased dialog slugs and legacy node fragments, whereas the viewer matches the canvas file stem and `d-id` literally. Follow the shared-address approach used by `sbt-mele`: generate the fragment from the same identifier written as the target node's `d-id`, and generate the slug from the same canonical resource spelling used for the exported canvas filename. State and transition indices occupy different namespaces and must remain distinguishable. Within `FOO-0.d.canvas`, use `state-0` and `transition-0` for nodes from `FOO.DLG`, and `state-BAR-0` and `transition-BAR-0` for nodes from `BAR.DLG`; use the catalog's spelling for the other resource. Identifiers are scoped to a canvas, so the same `BAR.DLG` state may have a different identifier when `BAR.DLG` is the root.

Identify DLG resources case-insensitively while retaining their canonical catalog spelling for emitted names and cross-resource identifiers. An explicitly requested resource with different casing must produce the same filename and addressable graph as the catalog spelling. This deliberately replaces the old link and `d-id` formats for newly generated output; compatibility with older viewer versions, migration of already exported files or spreadsheets, and changes to the viewer are out of scope. Preserve the existing canvas-local IDs and other export metadata apart from the canonical spelling needed to resolve mixed-case aliases.

# Acceptance criteria

- [x] A fixture with `A → B → D` and `A → C → D` exports one `D` and both incoming edges as `d-kind: "normal"`; convergence contributes to layout and the canvas passes dCanvas structural validation.
- [x] Self-loops, ancestor cycles, and cycles reached through converging branches produce loop edges only for edges that close cycles in the chosen depth-first traversal; these edges remain in the document but do not influence layout.
- [x] Changing the order in which the branches are traversed may change which edge is marked as a loop, but removing all loop edges leaves the reachable exported graph acyclic; no transition or reachable state is lost.
- [x] Classification uses the graph structure without evaluating dialog conditions or actions; retained transitions, skipped empty transitions, parallel edges, and their conditions keep their existing meaning.
- [x] DLG names are matched case-insensitively for state identity: a mixed-case reference back to an already encountered state is recognized as the same state, while equally numbered states in different DLG resources remain distinct.
- [x] Each exported line and retained reply has an unambiguous canvas-local `d-id`: root-resource nodes use `state-<index>` or `transition-<index>`, and nodes from another DLG include its canonical resource name, as in `state-BAR-0` and `transition-BAR-0`.
- [x] Every generated dialog-context URL for a line, reply, or addressable journal action uses the exact exported canvas file stem as its slug and a fragment that resolves to the intended node's `d-id` under the `v1.0` viewer's literal, percent-decoding fragment rules.
- [x] Exporting a resource by an explicitly supplied name with different casing produces the same catalog-cased canvas filename and matching text-context links as exporting it by its catalog name.
- [x] `dialog list --with-length` still counts each retained node once, and `text export` retains the same dialog context and labels except for the agreed address changes and correction of mixed-case duplicate states.
- [x] Fixture-based end-to-end tests exercise both dialog and text exports, including a cross-resource dialog, a state and transition with the same index, journal links, mixed-case references, convergence, and loops; tests require neither an installed game nor a sibling repository.
- [x] Existing URL formats and exported files are not migrated, no backward-compatibility mode is added, and the viewer is not modified.
