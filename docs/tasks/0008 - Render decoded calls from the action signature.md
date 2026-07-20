---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority: 5
blocked by:
kind: Feature
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

Replace the pragmatic decoded-call renderer in the **compiled script context** (`parser/bcs.go`, `renderCall`) with a fully signature-driven one. Today the renderer hardcodes the `Name(<object>,<strref>)` / `Name(<strref>)` shapes: every parameter other than the object and the StrRef is silently dropped (`DisplayStringWait(O:Object*,I:Time*,I:StrRef*)` renders without its `Time`), and the argument order is canonical rather than the signature's actual order (a Torment-reordered signature would render StrRef in the wrong position).

The new renderer walks `sig.Parameters` in signature order and renders each parameter from its typed slot pool in the compiled action block:

- `O:` — object slots in order: o2 first, then o3 (o1 stays reserved for the ActionOverride target). Render as today: quoted name string wins, then an `OBJECT.IDS` identifier resolved by value.
- `I:` — integer slots in order: logical slots 0, 1, 2 map to physical fields 0, 3, 4 (the existing `intSlotField` mapping). Render the raw decimal value; symbolic IDS lookup for `I:…*IdsFile` parameters is out of scope.
- `P:` — the point from physical fields 1–2, rendered as `[x.y]`. The block carries only one point.
- `S:` — the two string parameters, rendered quoted. They are currently read and **discarded** in `readAction`; keep them on `bcsAction`.

**The hard rule: nothing may branch on an action's name or on an expected signature shape.** No special-casing of `DisplayString`, `ActionOverride`, or anyone else. Whatever the renderer cannot map from the block, it renders as `?`:

- `A:` and `T:` parameters (nested action/trigger — they have no slot in a compiled block; the compiler folds `ActionOverride` into o1 + the inner action's id).
- Overflow: a 3rd/4th `S:` parameter (the engine concatenates Area+Name pairs for those actions, and telling which requires hardcoding action ids — exactly what this task forbids), a 2nd `P:`, a 4th `I:`, a 3rd `O:`.
- Any parameter type the renderer does not recognise.
- An exotic object specification (non-zero EA/GENERAL/RACE/… fields) — currently `[object]`; unify to `?`.

Unchanged: the display-family gate (only actions whose signature carries a StrRef contribute script context) and the StrRef extraction via `StrRefIntSlot`.

One parser fix along the way: `parseActionParameter` trims only a trailing `*`, so the `I:Value*Time` form (where the part after `*` names the IDS file constraining the value) yields the parameter name `Value*Time`. Cut the name at the first `*` instead; the IDS-file tag may be kept or dropped, but it must not leak into the name.

# Acceptance criteria

- [x] The end-to-end anchor still renders `DisplayString(LastTrigger,14381)`, and the objectless anchor still renders `EraseJournalEntry(26331)`.
- [x] All integer parameters render in signature order: `DisplayStringWait(O:Object*,I:Time*,I:StrRef*)` with Time=7 renders as `DisplayStringWait(Myself,7,14381)`.
- [x] A Torment-reordered signature renders its arguments in the signature's actual order — the StrRef appears at its declared position, not always last.
- [x] `S:` parameters render as quoted strings from the block; `P:` renders as `[x.y]`.
- [x] An unmappable parameter (`A:`, `T:`, unknown type, or slot-pool overflow) renders as `?`; exotic object specifications render as `?` too.
- [x] The renderer contains no action-name literals and no branching on expected signature shapes.
