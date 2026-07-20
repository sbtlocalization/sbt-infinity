---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
blocked by:
  - "[[0002 - Decompile compiled scripts (.BCS) to display-family calls]]"
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

End-to-end **compiled script context**: StrRefs referenced by display-family actions in standalone `.BCS` resources gain context in the `text export` spreadsheet, surfaced in a new `SCRIPTS` section.

This wires the `.BCS` decompiler (task 0002) into `text export`. Scripts are listed and read through `InfinityFs` as the existing dialog/creature scans do. A new context category `SCRIPTS` is added, keyed by script file name and valued by the decoded call (e.g. `DisplayString(LastTrigger,14381)`), rendered as a key→lines list like the existing sections. No node link for this section (compiled scripts carry no node URL).

`FileType_BCS` is added both as a handled case in the `text export` flow and to the default context-type list, so `--context-from all` includes it and `--context-from bcs` scopes to it alone. At this stage `SCRIPTS` entries show the bare decoded call without a usage annotation — the annotation is added in task 0005.

# Acceptance criteria

- [x] Running `text export --context-from bcs` produces a `SCRIPTS` section keyed by script file, with the decoded call as the value.
- [x] For `CTFB.BCS`, StrRef `14381` shows `DisplayString(LastTrigger,14381)`.
- [x] Only StrRefs genuinely referenced by display-family actions appear; no noise from other actions.
- [x] `FileType_BCS` is included in `--context-from all` and selectable via `--context-from bcs`.
- [x] A malformed/unparseable script is skipped with a warning and does not abort the export (graceful degradation).
- [x] Collection-seam test: loading compiled script context into a `TextCollection` produces the expected `Context`/`Labels` entries and leaves untouched ids alone.
