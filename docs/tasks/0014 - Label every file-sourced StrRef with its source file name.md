---
epic:
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

Make the exported spreadsheet filterable by source file. The `labels` column is a de-duplicated set, so a translator can filter the sheet down to every line coming from one resource — but today only dialogs carry their file name (the `DlgName` label). Extend this so every StrRef gathered from a concrete game file also gets that file's name as a label.

Add a source-file label at every Context source that scans a real file: spells, items, creatures, effects, stores, areas, world maps, projectiles, UI, compiled scripts (`.BCS`), and the `2DA` sources. Dialogs already do this — leave them untouched.

The label is the **raw file name as it comes from the filesystem, plus its extension** — no case normalization. This deliberately differs from how some context bodies render the name (e.g. creature bodies lower-case it); the label must stay the raw file-system spelling so filtering is consistent. For the ranger Tracking source, where the file name is derived from a 2DA row key, the label is `<rowKey>.ARE`.

Labels are added **always**, independent of the `--long` flag from the collapse task — filtering is their whole purpose.

Note: several `2DA` loaders currently discard their file-name argument (`_ string`); they will start using it to emit the label.

# Acceptance criteria

- [x] Every StrRef sourced from a spell, item, creature, effect, store, area, world map, projectile, UI, compiled script, or 2DA file carries that file's name (with extension) as a label.
- [x] The label is the raw filesystem spelling — no forced upper/lower casing — even where the context body renders the name in a different case.
- [x] Ranger Tracking StrRefs are labelled `<rowKey>.ARE`.
- [x] Dialog labels are unchanged.
- [x] Labels are present regardless of `--long`.
- [x] Filtering the `labels` column by a file name (e.g. `MISLEAD.SPL`) returns exactly the StrRefs that file contributes.
