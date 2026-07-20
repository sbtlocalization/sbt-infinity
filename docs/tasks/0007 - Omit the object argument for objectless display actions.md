---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
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

Fix the **compiled script context** (`SCRIPTS` section, task 0004): the decoded call always renders an object argument, even for display-family actions whose signature carries no object parameter. An all-zero object slot then resolves through `OBJECT.IDS` (`0 Nothing`), producing a phantom first argument — IWD's `EraseJournalEntry(I:STRREF*)` renders as `EraseJournalEntry(Nothing,26331)` instead of `EraseJournalEntry(26331)`.

The renderer must consult the action's parsed signature (already resolved from the game's `ACTION.IDS` when the action is matched) and include the object argument only when the signature carries an object-typed parameter. Actions that do carry one (e.g. `DisplayString(O:Object*,I:StrRef*)`) keep rendering as today.

The bug went unnoticed because the end-to-end validation from the epic (`CTFB.BCS` → `DisplayString(LastTrigger,14381)`) used an action *with* an object parameter; it only shows on objectless display-family actions.

# Acceptance criteria

- [x] A display-family action whose signature has no object parameter renders as `Name(<strref>)`, with no phantom object argument.
- [x] An action whose signature has an object parameter keeps rendering as `Name(<object>,<strref>)`.
- [x] Parser-seam test: decoding a compiled action for an objectless signature (the `EraseJournalEntry(I:STRREF*)` shape) yields a call without an object argument.
- [x] Existing decoded-call rendering for `DisplayString(LastTrigger,14381)` is unchanged.
