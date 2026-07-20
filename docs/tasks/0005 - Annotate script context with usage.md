---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
blocked by:
  - "[[0004 - Gather compiled script context (SCRIPTS)]]"
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

Make opaque compiled-script names intelligible by annotating `SCRIPTS` entries with the script's **usage** — where it is attached in the game, e.g. `CTFB.BCS [trap @ AR1234]`, `CHICKEN.CRE [override @ CHICKEN.CRE]`.

While scanning **areas** (`ARE`: area script, trap/region scripts, door and container scripts, embedded actor script slots) and **creatures** (`CRE`: the five script slots), build a reverse index mapping each script resref to its usage sites. These are the only resource types that carry `.BCS` references; items, spells, projectiles, effects, stores, world maps and 2DA do not. The index is populated from whatever area/creature scans the selected `--context-from` runs — it is not forced on.

When a `SCRIPTS` entry is created, its key embeds the usage drawn from this index. Because of that, compiled scripts must be processed **after** areas and creatures; this ordering is enforced in the export flow and called out with a code comment. With `--context-from all` (the common case) the annotation is always present; with `--context-from bcs` alone the entries show the bare decoded call without annotation.

The function that renders a single usage site (`trap @ AR1234`, `override @ CHICKEN.CRE`) is unit-tested directly; the full area/creature scan that feeds the index is not mocked.

# Acceptance criteria

- [x] Running `text export --context-from all` annotates `SCRIPTS` keys with usage, e.g. `CTFB.BCS [trap @ AR1234]`.
- [x] The reverse index is built from ARE (area/trap/region/door/container/actor script slots) and CRE (five script slots) scans.
- [x] Compiled scripts are processed after areas and creatures, with the ordering enforced and a code comment explaining why.
- [x] With `--context-from bcs` alone, `SCRIPTS` entries show the bare decoded call with no usage annotation.
- [x] The usage-formatting function is unit-tested directly for cases like `trap @ AR1234` and `override @ CHICKEN.CRE`.
