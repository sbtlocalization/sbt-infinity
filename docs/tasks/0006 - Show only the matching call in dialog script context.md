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

Fix the **dialog script context** (`DIALOG SCRIPTS` section, task 0003): each entry currently embeds the *entire* transition action text, even when that block contains several statements. A translator looking up a StrRef sees unrelated calls (`SetGlobal(...)`, `GiveItemCreate(...)`) glued to the one display-family call that actually shows the line.

Instead, each entry must render **only the statement that references the StrRef** — the original source text of that single call, trimmed of surrounding whitespace but otherwise untouched (no reconstruction or normalisation; whatever is stored in the DLG is what the translator sees). The `Action: <call> @ <node-url>` shape of the entry stays the same.

The statement splitter already isolates individual calls when scanning a transition's action text; it just does not retain the original text of each statement. Retain it alongside the parsed name/arguments and use it as the rendered call. When one action block contains several display-family calls, each affected StrRef gets its own entry carrying only its own call.

# Acceptance criteria

- [x] For a transition action block with several statements, the `DIALOG SCRIPTS` entry for a StrRef contains only the call that references it, not the whole block.
- [x] The rendered call is the original statement text, trimmed at the edges only.
- [x] The entry keeps the `Action: <call> @ <node-url>` format and node link.
- [x] Two display-family calls in one block produce two entries, each with its own call.
- [x] Parser-seam test: splitting a multi-statement action block yields each statement's original text.
- [x] Collection-seam test: a block mixing `SetGlobal("v","GLOBAL",1)` and `DisplayStringHead(Myself,14381)` yields context for `14381` containing only the `DisplayStringHead` call.
