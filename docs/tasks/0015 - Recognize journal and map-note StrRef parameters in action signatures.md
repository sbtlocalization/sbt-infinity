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

The display-family detector identifies a StrRef parameter only when an integer (`I:`) parameter is named exactly `StrRef` (case-insensitive). Several `ACTION.IDS` actions carry a StrRef under a different parameter name and are therefore silently dropped from the display-family set, so their StrRefs never reach `text export` context — neither from compiled `.BCS` scripts nor from dialog transition actions.

Missing actions (identical across BG1/BG2/IWD/PST):

- `AddJournalEntry(I:Entry*,I:Type*JourType)` — StrRef parameter named `Entry`.
- `AddMapNote(P:Position*,I:StringRef*)` — StrRef parameter named `StringRef`.
- `AddMapNoteColor(P:Position*,I:StringRef*,I:Color*Mapnotes)` — StrRef parameter named `StringRef`.

Extend the detector to match a **closed allowlist** of integer parameter names — `strref`, `stringref`, `entry` — instead of the single literal `strref`. The `I:` type gate stays, which is what keeps the string-typed `S:Entry*` parameters (`LeaveAreaLUAEntry`, `LeaveAreaLUAPanicEntry`, `ForceRandomEncounterEntry` — area/script names, not StrRefs) excluded.

The set is closed and complete: enumerating every distinct `I:` parameter name across BG1/BG2/IWD/PST yields exactly these three StrRef-plausible names; there is no `I:Entry` other than `AddJournalEntry` (unique per game) and no `I:StringRef` other than the two map-note actions. Rejected alternative: a per-action-name exception for `AddJournalEntry`. It would still need to know which integer slot holds the StrRef, and PST reorders parameters — so it is strictly more fragile than one extra allowlisted parameter name for no benefit.

The integer-slot index is already computed by parameter type, so the existing consumers (`bcs.go` via `StrRefIntSlot`, `builder.go` via `StrRefPositional`) pick up the fix without further change. For `AddJournalEntry`, `Entry` is the first integer parameter (slot 0), ahead of `I:Type`.

# Acceptance criteria

- [x] `AddJournalEntry(I:Entry*,I:Type*JourType)` resolves as a display-family action with its StrRef at integer slot 0 (positional index 0).
- [x] `AddMapNote(P:Position*,I:StringRef*)` and `AddMapNoteColor(P:Position*,I:StringRef*,I:Color*Mapnotes)` resolve as display-family actions with the `StringRef` parameter as their StrRef.
- [x] Existing `StrRef`/`STRREF`/`Strref` matching keeps working; the three names match case-insensitively.
- [x] String-typed `S:Entry*` parameters (e.g. `LeaveAreaLUAEntry`) are still NOT treated as StrRef parameters.
- [x] Parser-seam tests cover `AddJournalEntry` and `AddMapNote`, mirroring the existing action-signature test style.
- [x] A `text export` over data containing these actions surfaces their StrRefs in `SCRIPTS` / `DIALOG SCRIPTS` context. *(Proven at the compiled-script seam: `TestDecodeCompiledScript_AddJournalEntry` decodes the journal StrRef end-to-end through `DecodeCompiledScript`, which is the same `DisplayFamilyActions`-fed path `text export` uses. Not run against a live game archive.)*
