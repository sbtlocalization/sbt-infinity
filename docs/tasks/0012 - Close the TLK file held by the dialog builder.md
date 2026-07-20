---
epic:
parent:
status: Done
priority: 1
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

The dialog builder opens the TLK file when loading dialogs and never closes it — the handle lives for the rest of the process. Today this is only a descriptor leak (dialog text lookups keep working precisely *because* nothing closes the file), but it is an unmanaged lifetime of a lazily-read kaitai resource, the same shape that produced the CRE effects regression.

Give the TLK handle an explicit lifetime: the builder closes it once dialog loading is finished, or exposes a close method its callers invoke when done. Mind the laziness of TLK string entries — the file must stay open until every text lookup the builder serves has been materialized (or lookups must read through while the file is provably open).

# Acceptance criteria

- [x] The TLK file opened by the dialog builder is closed by the time its owner is done with it, on both success and error paths.
- [x] All dialog text and journal lookups still return correct strings (`text export` dialog context and `dialog ex` output unchanged).
- [x] No lazy TLK read can occur after the close (eager materialization or lifetime documented at the seam).
