---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority: 2
blocked by:
kind: Bug
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

The DLG helper methods that fetch trigger and action text of dialog states and transitions handle read failures inconsistently, and two of the three ways are wrong:

1. A transition whose flags promise a trigger or an action but whose lazy read fails hits `log.Fatalln` — one malformed transition kills the entire export instead of skipping the one dialog.
2. A failed read of a *state* trigger is silently treated as "no trigger", which misclassifies the state as a root state — no log line, wrong dialog structure.

Both must degrade the way the rest of the toolkit does: log a warning naming the dialog resource and carry on, treating the unreadable text as absent. No process termination from library code.

# Acceptance criteria

- [x] No `log.Fatalln` remains in the DLG text helpers; a malformed transition logs a warning and is skipped, and the export completes.
- [x] A failed state-trigger read logs a warning instead of silently classifying the state as root.
- [x] Well-formed dialogs are processed identically to before.
