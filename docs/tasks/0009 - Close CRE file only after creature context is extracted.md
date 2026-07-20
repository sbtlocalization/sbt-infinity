---
epic:
parent:
status: Done
priority: 4
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

Regression in `text export`: the `EFFECTS` context section lost all `Creature <file>.CRE → effect <n>` entries, and StrRefs referenced only by creature effects lost the `creature` label.

The creature scan closes each CRE file right after the kaitai `Read()` call and only then extracts context from the parsed object. The effects list of a creature is a *lazy* kaitai instance — it seeks the underlying stream on first access, which happens inside the context extraction. By that time the file is closed, the read fails with `os.ErrInvalid`, and the failure is swallowed (see the companion bug on silent lazy-getter errors), so every creature silently contributes no effect context.

Close the CRE file only after `LoadContextFromCreature` returns, mirroring the ordering the shared `processFiles` helper already uses for ARE/ITM/SPL/STO. Do not revert to `defer` inside the loop — the file should live exactly as long as the unread lazy fields do, not until the whole scan finishes.

Introduced by commit `454b190` ("Skip unsupported versions of CRE during `text ex`"), where switching the loop from return-on-first-error to warn-and-continue moved the close before the extraction. Verified against IWD:EE: with the file kept open, `SWORD01.CRE` yields effect 27 → opcode 267 → StrRef 39752 ("Stricken by a foul disease"), matching pre-regression exports.

# Acceptance criteria

- [x] `text export --context-from all` (or `cre`) restores `Creature <file>.CRE → effect <n>` entries in the `EFFECTS` section and the `creature` label for StrRefs referenced by creature effects (verify on IWD:EE StrRef 39752: `SWORD01.CRE`, `BPARTY03.CRE`, `hparty03.CRE`).
- [x] The CRE file is closed after context extraction for every loop outcome (parse error, extraction error, success) — no descriptor leak on any path.
- [x] Existing export behavior for all other context sources is unchanged.
