---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
blocked by:
kind: Task
mode: AFK
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

A parser that turns an `ACTION.IDS` signature string (the identifier returned by `ParseIds`, e.g. `DisplayString(O:Object*,I:StrRef*)`) into an ordered list of typed parameters, and from that identifies the **display-family actions** of the current game.

The parser is fed the game's own `ACTION.IDS` at run time (never a hard-coded set), so games that add or reorder actions — notably Planescape: Torment with its `FloatMessage*` variants and reordered `DisplayStringPoint*` — resolve correctly.

For every signature it records, for each `StrRef` parameter (type integer, name matched case-insensitively against `StrRef`/`STRREF`/`Strref`):

- its **positional index** among all parameters — used by the source-form (dialog) path to pick the bare integer argument; and
- its **integer-slot index** — the count of integer-typed parameters preceding it. Parameters of other types (`O`, `S`, `P`) do not advance the integer count. This is what keeps Torment's reordering correct for the compiled-script path.

An action is in the display-family set iff its signature carries at least one `StrRef` integer parameter. Triggers are parsed for completeness but never contribute a StrRef.

This task delivers the parser and its data model only; the dialog and compiled-script paths consume it in later tasks. Tests live at the parser seam, following the prior art of the existing 2DA parser tests (in-memory reader, assert on returned data).

# Acceptance criteria

- [x] A signature string resolves to the correct ordered parameter list (type + name per parameter).
- [x] For a `StrRef` parameter, both its positional index and its integer-slot index are reported, with non-integer parameters not advancing the integer count.
- [x] The Torment reordering case is covered: `I:Strref*` placed before vs. after a `P:Location*` parameter yields different integer-slot indices, proving the by-type counting.
- [x] `StrRef` name matching is case-insensitive (`StrRef`, `STRREF`, `Strref`).
- [x] The display-family set is derived from a supplied parsed `ACTION.IDS`, not hard-coded; an action with no `StrRef` parameter is excluded.
- [x] Parser-seam tests cover the above, mirroring the 2DA parser test style.
