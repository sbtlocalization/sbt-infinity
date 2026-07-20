---
epic: "[[EPIC-0001 - Script context extraction for text export]]"
parent:
status: Done
priority:
blocked by:
  - "[[0001 - Parse ACTION.IDS action signatures]]"
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

A compiled-script parser/decompiler for standalone `.BCS` resources. Given the bytes of a script it walks the block structure (`SC`/`CR`/`CO`/`TR`/`RS`/`RE`/`AC`/`OB`), tolerant of the format's glued markers (`100AC`, `151OB`).

For each action it resolves the action id against the signature table (task 0001). When the action is in the display-family set it reads the StrRef from the **mapped integer slot**: the scalar slot layout of an action block is `int0`, point(x,y), `int1`, `int2`, two strings, and logical integer slots map to physical fields `int0 → 0`, `int1 → 3`, `int2 → 4`.

It renders a **decoded call** for the action: the action name plus its parameters. Object decoding is pragmatic — resolve the `OBJECT.IDS` identifier (e.g. `LastTrigger`, `Myself`); honour the convention that an action's first object slot (`o1`) is reserved for the `ActionOverride` target, so a normal single object lives in `o2`; show an object name string when present; fall back to a compact form for exotic EA/RACE/… combinations rather than emitting a full `[EA.GENERAL…]` spec. No trailing `// …` comments.

Decoding degrades gracefully: a malformed action/block is skipped rather than aborting the whole script.

This task delivers the parser only — yielding the `{strref, decoded call}` pairs for a script. Wiring into `text export` is task 0004. Validate end-to-end against `CTFB.BCS` (BG1), which decodes to StrRef `14381` ("Trap Sprung") via `DisplayString(LastTrigger,14381)`, confirming the integer-slot mapping and the `o1`-reserved object convention against the Near Infinity decompile.

Tests sit at the parser seam (the primary test target), in the 2DA-parser style: a small BCS source of the `CTFB` shape yields the expected pairs and ignores non-display-family actions.

# Acceptance criteria

- [x] A small in-memory BCS source of the `CTFB` shape yields `{14381, "DisplayString(LastTrigger,14381)"}`.
- [x] Actions that are not display-family contribute no StrRef pairs.
- [x] The integer-slot mapping (`int0→0`, `int1→3`, `int2→4`) is exercised so a StrRef in a non-zero slot is read from the correct field.
- [x] The decoded call honours the `o1`=ActionOverride convention (a normal single object renders from `o2`) and resolves `OBJECT.IDS` identifiers such as `LastTrigger`/`Myself`.
- [x] Glued block markers (`100AC`, `151OB`) parse correctly.
- [x] A malformed action/block is skipped without aborting the rest of the script.
- [x] `.BS` and other non-`.BCS` formats are out of scope and not parsed.
