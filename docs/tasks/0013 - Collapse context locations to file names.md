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

Shorten how Context locations are rendered in the exported spreadsheet. Today a StrRef reused across many resources produces a long per-location list where each line carries internal indices the translator can never map back to the game (e.g. a single "Protection from spell" message expands to dozens of `Spell MISLEAD.SPL → ability 0 → effect 20` lines). Collapse each location value to the bare source file name, letting the existing de-duplication fold the repeats away.

Scope of the collapse — the location value keeps only the file name (drop the `Spell`/`Item`/`Creature` word and every `→ … N` index):

- **EFFECTS** — creature, item and spell effect contributions → `<file>.CRE` / `<file>.ITM` / `<file>.SPL`.
- **WORLD MAPS** — world map name, area caption, area tooltip → `<file>.WMP`.
- **AREAS** — info point, trigger speaker, container lockpick, door unlock, door speaker → `<file>.ARE`.
- **STORES** — drink name → `<file>.STO`.

Left unchanged (not a `file → index` shape): standalone `.EFF` contributions (already a bare file name), the `UI` context (file name is the key, value is the in-file `window N → control N` coordinate), the tracking default-embed value, and every `2DA` context value.

Tracking is a related but separate reformat, applied **always** and independent of the flag below: the ranger Tracking area value becomes `<rowKey>.ARE` instead of `Area <rowKey>` / bare `<rowKey>` — the row key already is an area resource name.

Add a global `--long` flag to `text export` (default off) that restores the current verbose location strings verbatim (with prefixes and indices) for the four collapsed sections. It does not affect the tracking `.ARE` reformat, nor labels. Carry the flag as a field on the text collection, set once in the command handler, so the many context builders can read it without threading a parameter through every signature.

Fix the inline-vs-list heuristic so the collapsed output renders cleanly: de-duplicate the values *before* the length threshold decides between a comma-joined inline string and a bullet list, and sort the unique values so the export is deterministic across runs (no noisy diffs from file-scan order).

# Acceptance criteria

- [x] `text export` (default) renders EFFECTS / WORLD MAPS / AREAS / STORES location values as de-duplicated, sorted bare file names — no `Spell`/`Item`/`Creature` prefix, no `→ … N` indices.
- [x] A message shared across many resources collapses to the distinct source files only (e.g. `MISLEAD.SPL, PROJIMAG.SPL, SIMULACR.SPL` instead of the former ~24-line index list).
- [x] Ranger Tracking area values are `<rowKey>.ARE` in both modes.
- [x] `--long` restores the exact current verbose strings for the four collapsed sections; tracking `.ARE` and labels are unaffected.
- [x] Inline-vs-list choice is based on the count of *unique* values; a single unique file renders inline, not as a one-item bullet list.
- [x] Location values are sorted; re-running the export yields identical output.
- [x] `.EFF`, `UI`, tracking-default, and `2DA` context values are unchanged.
