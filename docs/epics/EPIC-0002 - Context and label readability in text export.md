---
status: In progress  # Enum: Backlog, To do, In progress, Done
priority:        # Number: from 1 to 5, or empty
blocked by:      # YAML list of refs: ["[[blocking task 1]]", "[[blocking task 2]]"]
kind: PRD        # Possible values: for now only PRD
---
# Tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - epic == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      value.asFile().properties.status != "Done" || acc,
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon("square-check-big"),
      if(formula.is_blocked,
        icon("construction"), 
        icon("square")
      )
    )
  progress_sort: |-
    if(status == "Done", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == "In progress", 
          5, 
          0)))
  progress_string: |-
    if(status == "Done", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon("tally-" + priority), null)
  kind_icon: |-
    if(kind == "Bug",
      icon('bug'),
      ''
    )
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
  formula.kind_icon:
    displayName: " "
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - formula.kind_icon
      - file.name
      - formula.priority
      - status
      - blocked by
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
      formula.kind_icon: -1
      file.name: 467
      note.status: 125

```
<!-- end unchanged block -->

# Problem Statement

The `context` and `labels` columns of the spreadsheet produced by `text export` exist to answer two questions for a translator: *what is this line and where does it appear*, and *how do I narrow the sheet down to the lines I care about*. As more context sources were added, both columns outgrew the answers.

Context bodies restate internal structure the translator cannot map back to the game. A single "Protection from spell" message expands into dozens of `Spell MISLEAD.SPL → ability 0 → effect 20` lines that differ only in indices. A generic remark floated by ten scripts produces ten near-identical entries repeating the same decoded call. A quest line touched inside dialogs fills its cell with node URLs that differ only in which root state they point at.

Labels have the opposite failure. Only dialogs contributed their file name, so a translator could not filter by "everything coming from this creature" at all — and where labels *were* added generously, one Baldur's Gate II row ended up carrying 403 area names, another 258 items. A filter that is either missing or a wall of names is no filter.

# Solution

Treat the two columns as products in their own right and make them short, de-duplicated and filterable.

- **Collapse the body to what is addressable.** Location values keep the bare source file name and drop prefixes and internal indices; a `--long` flag restores the verbose form for anyone who needs it.
- **Key context by the thing that repeats least.** Script sections are turned inside out — keyed by the decoded call, listing the files that make it — instead of keyed by file and repeating the call. Node links survive only where they carry situation, namely `AddJournalEntry`.
- **Move identity into labels.** Every StrRef gathered from a real file is labelled with that file's name, and script usage (the resource a script is attached to, plus its `ARE` slot) becomes labels rather than an annotation buried in the body.
- **Cap labels per resource family.** Past ten labels of one family the whole family is replaced by a single marker (`*.CRE`, `*.ARE`, `many speakers`), all-or-nothing, so a filter is either exact or visibly absent — never quietly short.

The decisions behind this are recorded in [[ADR-0001 - Context location representation and file filtering]], [[ADR-0002 - Key script context by the decoded call and move script usage into labels]] and [[ADR-0003 - Cap file-name labels per resource family]].

# Out of Scope

- Adding new context *sources* — that is [[EPIC-0001 - Script context extraction for text export]] and its successors. This epic only reshapes what is already gathered.
- Resource-lifetime and swallowed-error bugs in the extraction pass (tasks 0009, 0010, 0012); they affect whether context is gathered at all, not how it reads.
- Any change to the spreadsheet's column set or to the import side of the round trip.
