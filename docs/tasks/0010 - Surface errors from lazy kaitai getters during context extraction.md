---
epic:
parent:
status: Done
priority: 3
blocked by:
  - "[[0009 - Close CRE file only after creature context is extracted]]"
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

Context extraction in the `text` package guards fourteen lazy kaitai getters with `if x, err := ...; err == nil { ... }` and no error branch, so a failed read silently drops a whole context section. This is what kept the CRE effects regression invisible: the export completed "successfully" while quietly losing data. Affected loaders and getters:

- creature: effects
- area: actors, regions, containers, doors, automap notes, rest encounters
- item: global effects, extended headers, per-ability effects
- spell: effects, extended headers, per-ability effects
- store: drinks

Make these failures visible while keeping graceful degradation: accumulate the errors within one resource via `errors.Join` and return the combined error from the loader, keeping whatever context *was* successfully extracted. The scan loops already print a per-file warning in `--verbose` mode when a loader returns an error — a partially readable resource then both contributes its good sections and shows up in the warnings, instead of choosing between the two. UI-screen and world-map loaders already propagate errors this way and are the model to follow.

# Acceptance criteria

- [x] No `err == nil`-guarded lazy getter in the context loaders discards its error; each failure is joined into the loader's returned error.
- [x] A resource with one failing section still contributes context from its other sections.
- [x] `text export --verbose` prints a warning naming the file when any lazy getter fails; non-verbose output stays quiet.
- [x] A loader returning a joined error still lets the scan continue with the remaining files.
- [x] Collection-seam test: a loader fed an object whose lazy getter fails returns a non-nil error and still records the context that was readable.
