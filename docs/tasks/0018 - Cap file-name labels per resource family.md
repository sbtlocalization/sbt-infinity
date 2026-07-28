---
epic: "[[EPIC-0002 - Context and label readability in text export]]"
parent:
status: Done
priority: 3
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

Stop the labels column from turning into a wall of file names. Decided in [[ADR-0003 - Cap file-name labels per resource family]].

A label belongs to a **family**, and past **ten** labels of one family the whole family is replaced by a single marker: `*.CRE`, `*.ARE`, `*.ITM`, and `many speakers` for the `SPEAKER` family. All-or-nothing, one threshold for every family, so a filter is either exact or visibly absent — never quietly short. In Baldur's Gate II one row otherwise carries 403 area names, another 258 items, another 159 creatures.

The family of a label is its uppercased file extension, so nothing has to be enumerated: `TRADRAN.CRE` → `CRE`, `AR9711.ARE` → `ARE`. Labels with no extension have no family and are never capped — that covers every word label (`trap`, `creature`, `journal`, `script action`, …) and every dialog name, which is exempt on purpose.

The one exception has to be recorded rather than guessed. A creature's **verbal constants** are labelled with the creature's dialog resref, so a translator knows who says a bare "Yes?"; those labels are spelled exactly like dialog names, and cannot be told apart by looking at them. So the label store keeps the family alongside the label, and that one call site in the creature scan names the family `SPEAKER` explicitly. Everywhere else keeps adding labels as it does today.

The cap applies at render time only — the full set stays in the model, so nothing downstream loses information it currently has.

# Acceptance criteria

- [x] A StrRef with more than ten labels of one extension renders that family as `*.EXT` and nothing else of that family; families under the threshold are untouched.
- [x] The threshold is per family: a row with 400 `.CRE` labels and 2 `.ARE` labels renders `*.CRE` plus both area names.
- [x] Word labels and dialog names are never replaced, however many there are.
- [x] Verbal-constant labels belong to the `SPEAKER` family and collapse to `many speakers`, while dialog names added by the dialog scan stay intact on the same row.
- [x] Exactly ten labels of a family still render in full; eleven collapse.
- [x] Label output stays sorted and de-duplicated, with markers sorting among the rest.
- [x] `--long` does not change the cap.
- [x] `go test ./...` passes.
