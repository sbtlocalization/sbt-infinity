---
epic:
parent:
status: Done
priority: 1
blocked by:
kind: Task
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

Make `dialog export` generate dCanvas 1.1 documents using `github.com/sbtlocalization/dcanvas v1.1.2` instead of the bundled format library. The bundled format historically called dCanvas 2.0 is the obsolete internal draft now archived as 0.2 in the master format repository. This is a clean switch for newly generated files, not a compatibility layer: existing exports will be regenerated from the game's DLG and TLK resources.

Raise the project's minimum Go version to 1.27 and align release builds with that requirement. Remove the bundled format implementation and delegate encoding and graph layout to the external library. Keep Infinity-specific conversion in this project.

## Field mapping

| Legacy field | dCanvas 1.1 output |
|---|---|
| `x-dCanvasVersion: "2.0"` | `d-version: "1.1"` |
| `x-nodeRole: "state"` | `d-kind: "line"` |
| `x-nodeRole: "transition"` | `d-kind: "reply"`, including transitions without text |
| `x-nodeId` | `d-id`, retaining the existing `DLG[index]` value |
| `x-textId` | `d-textId`, retaining the existing `#StrRef` convention |
| `x-trigger` on a node | `d-condition` on that node |
| `x-condition` on an edge | `d-condition` on that edge |
| `x-action` | `d-action` |
| `x-sound` | `d-sound` |
| `x-character` | `d-character` for real speakers only |
| `x-journalText`, `x-journalTextId`, `x-journalSound` | Unchanged field names, emitted as project extensions through the library's `SetExtra` API |

Do not copy `state` or `transition` into `d-role`. A terminal transition keeps `d-kind: "reply"` and gains `d-role: "dialog-end"` instead of the synthetic character named `End dialog`. A state-loading error remains a red text node with the existing error message, has `d-role: "error"`, and carries no `d-kind`. Remove the synthetic character named `Answer` from ordinary replies as well.

Keep visible text exclusively in the existing nodes. Infinity does not need `edge.label`; do not generate edge labels or edge text IDs. Preserve canvas-local IDs, edge endpoints, colors, node sizes, sound prefix/suffix formatting, file names, and dialog node URLs. Domain IDs remain separate from the canvas-local IDs referenced by edges.

## Speakers

Preserve real speakers' names, portrait references, and gender. Include `d-character.textId` using the `#StrRef` convention when the speaker name's StrRef is available; omit it when unavailable or unset. Use the reference corresponding to the exported name, not a different name variant.

Retain all CRE speakers associated with a DLG instead of overwriting earlier speakers during scanning. Select the first CRE by resource name as the displayed speaker, independent of scan order or localized character names. Emit the remaining speakers in deterministic resource-name order as `d-alternativeCharacters`, with the same available metadata as the displayed speaker. Omit the alternatives field when there is only one speaker, and never emit alternatives without a displayed speaker. Include portrait assets for alternative speakers as well as the displayed speaker. Multiple-speaker support is required even if the real game data used during implementation has no such example; a fixture can establish the behavior.

## Graph and layout behavior

Keep the existing omission of empty nonterminal transitions that have no text, journal entry, or action, including transfer of their conditions to the resulting direct edges. Retain transitions with actions, journal entries, or dialog termination even when their text is empty.

Mark edges currently treated as loops with `d-kind: "loop"` and forward edges with `d-kind: "normal"`. Use the library's default `LoopCut` layout strategy so loop edges remain in the document but do not influence positioning. Remove the direct layout-engine call. Layout remains best-effort: report a layout failure and continue exporting with the initial coordinates rather than aborting the export or silently claiming that positioning succeeded.

## Verification and scope

Test the observable exported document, preferably by exporting representative resource fixtures through the existing command and decoding the result. Verify conformance to the canonical dCanvas 1.1 JSON Schema as well as the library's structural validation; these are complementary checks. Cover metadata preservation through encode/decode/re-encode, terminal and error nodes, condition transfer through skipped transitions, loops, speaker name references, and deterministic alternative speakers. Tests must not require an installed game or a sibling checkout of the master format repository.

Existing command flags, output naming, sound/StrRef reports, overlap scanning, and node URLs remain unchanged. No old-file reader, migration command, legacy format mode, reader changes, new layout flags, or epic is part of this task.

# Acceptance criteria

- [x] The project requires Go 1.27, release builds use a compatible toolchain, and the dependency is `github.com/sbtlocalization/dcanvas v1.1.2`; no bundled dCanvas implementation or direct layout-engine call remains.
- [x] `dialog export` produces `.d.canvas` files stamped `d-version: "1.1"` that decode with the external library and pass both canonical dCanvas 1.1 JSON Schema conformance and structural validation for representative fixtures.
- [x] State nodes use `d-kind: "line"`; all retained transition nodes use `d-kind: "reply"`, including transitions with empty text. Ordinary nodes do not carry `d-role: "state"` or `d-role: "transition"`.
- [x] Terminal transitions use `d-role: "dialog-end"`; error nodes use `d-role: "error"` without `d-kind` and retain the existing error text and red color. No synthetic `Answer` or `End dialog` character is emitted.
- [x] Domain IDs, text StrRefs, node and edge conditions, actions, sounds, and speaker metadata follow the agreed field mapping, with no obsolete dialogue field names left in new output.
- [x] Journal text, StrRefs, and sound references remain `x-` project extensions and survive an encode/decode/re-encode cycle without loss.
- [x] Player reply text remains in the reply node; the export adds neither `edge.label` nor an edge `d-textId`.
- [x] Real speakers retain their names, portrait references, and gender; name StrRefs are emitted as `d-character.textId` when available and omitted when unavailable or unset.
- [x] A multiple-CRE fixture proves that the first speaker by CRE resource name is displayed and the others are exported as alternatives with their available metadata and portrait assets; scan order and localization do not change the selection. A single-speaker fixture has no alternatives field.
- [x] Skipped empty transitions still transfer their conditions to direct edges, while action-only, journal-only, and terminal transitions remain represented as reply nodes.
- [x] Forward and loop edges carry `d-kind: "normal"` and `d-kind: "loop"` respectively; library layout excludes loop edges from positioning without removing them from the exported graph. A layout failure produces a warning and does not prevent export.
- [x] Existing flags, file names, canvas-local IDs, node URLs, colors, sizes, sound formatting, reports, and overlap scanning retain their behavior apart from the agreed format and metadata changes; old-file migration is not introduced.
- [x] Fixture-based tests cover the export and round-trip behavior without an installed game or sibling repository, and `go build` and `go test ./...` pass on Go 1.27.
