---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
blocked by:
  - "[[0001 - Parse ACTION.IDS action signatures]]"
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

End-to-end **dialog script context**: lines shown by a source-form action embedded in a dialog transition (`DLG`) gain context in the `text export` spreadsheet, surfaced in a new `DIALOG SCRIPTS` section.

The dialog transition action text is already stored decompiled in the dialog model (`Transition.Action`). This pass — folded into the existing dialog context pass that already runs during `text export` — is given the parsed `ACTION.IDS` (task 0001) so it can recognise display-family actions. For each transition action it matches the function name against the signature table; the StrRef is the bare integer argument at the StrRef parameter's positional index. Only integer-literal arguments are linked (string/tilde literals do not exist in stored DLG text), so e.g. `DisplayStringHead(Myself,14381)` links `14381` while `SetGlobal("v","GLOBAL",1)` links nothing.

A new context category `DIALOG SCRIPTS` is added, keyed by dialog name and valued as `Action: <call> @ <node-url>`, reusing the same node-URL scheme as the existing dialog context. The `Action:` prefix and separate section keep it from being confused with the spoken-line `DIALOGS` context.

# Acceptance criteria

- [x] Running `text export --context-from dlg` produces a `DIALOG SCRIPTS` section for StrRefs referenced by display-family actions in dialog transitions.
- [x] Each entry reads `Action: <decoded call> @ <node-url>` and links to the exact dialog node, reusing the existing node-URL scheme.
- [x] `DisplayStringHead(Myself,14381)` yields StrRef `14381`; `SetGlobal("v","GLOBAL",1)` yields none.
- [x] Only integer-literal StrRef arguments are linked.
- [x] The `DIALOG SCRIPTS` section is rendered distinctly from the spoken-line `DIALOGS` section.
- [x] Collection-seam test: loading dialog script context into a `TextCollection` produces the expected `Context`/`Labels` entries for the affected StrRefs and leaves untouched ids alone.
