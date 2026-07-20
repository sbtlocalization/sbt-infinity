---
status: Done  # Enum: Backlog, To do, In progress, Done
priority:        # Number: from 1 to 5, or empty
blocked by:      # YAML list of refs: ["[[blocking task 1]]", "[[blocking task 2]]"]
kind: PRD        # Possible values: for now only PRD
---
# Tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - epic == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      value.asFile().properties.status != "Done" || acc,
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon("square-check-big"),
      if(formula.is_blocked,
        icon("construction"), 
        icon("square")
      )
    )
  progress_sort: |-
    if(status == "Done", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == "In progress", 
          5, 
          0)))
  progress_string: |-
    if(status == "Done", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon("tally-" + priority), null)
  kind_icon: |-
    if(kind == "Bug",
      icon('bug'),
      ''
    )
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
  formula.kind_icon:
    displayName: " "
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - formula.kind_icon
      - file.name
      - formula.priority
      - status
      - blocked by
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
      formula.kind_icon: -1
      file.name: 467
      note.status: 125

```
<!-- end unchanged block -->

# Problem Statement

When translating a StrRef in the spreadsheet produced by `text export`, the translator needs to know what the line *is* and where it appears. Today the export gathers context from dialogs, creatures, areas, items, spells, stores and more — but **not from scripts**. Yet many lines are shown to the player purely through a script action (a fireball trap displays "Trap Sprung!", an NPC floats a remark over its head, a quest journal entry is set). Those lines arrive with no context at all, leaving the translator blind: they cannot tell whether the line is even used, in what situation it fires, or what tone fits.

Even when a translator learns a line lives in some script, the bare script name is opaque. `CTFB.BCS` means nothing on its own — the translator has no way to know it is the script of a trap region in a particular area.

# Solution

Extend `text export` to gather **script context** from two new sources, surfaced as two new sections in the `context` cell next to the existing ones:

- **`SCRIPTS`** — StrRefs referenced by display-family actions in **compiled scripts** (standalone `.BCS` resources). Each entry shows the script file and the decoded action call, e.g. `DisplayString(LastTrigger,14381)`.
- **`DIALOG SCRIPTS`** — StrRefs referenced by **source-form actions** embedded in dialog transitions (`DLG`). Each entry shows the action call and a link to the exact node, explicitly marked as an `Action` so it is never confused with the line a character speaks.

To make a compiled script intelligible, `SCRIPTS` entries are annotated with the script's **usage** — a reverse index built while scanning areas and creatures tells the translator that `CTFB.BCS` is a `trap @ AR1234`.

A companion `ids show` command (mirroring `2da show`) is added so IDS files such as `ACTION.IDS` can be inspected directly.

# User Stories

1. As a translator, I want lines shown by a script action to carry context, so that I am not translating blind text whose purpose I cannot see.
2. As a translator, I want to know which compiled script (`.BCS`) displays a given line, so that I can find the situation it belongs to.
3. As a translator, I want to see the decoded action call (e.g. `DisplayString(LastTrigger,14381)`) for a scripted line, so that I understand how and to whom the line is shown.
4. As a translator, I want a compiled script annotated with where it is attached (e.g. `trap @ AR1234`, `override @ CHICKEN.CRE`), so that an opaque name like `CTFB.BCS` becomes meaningful.
5. As a translator, I want lines shown by an action inside a dialog transition to carry context, so that floating remarks and journal updates inside dialogs are not missed.
6. As a translator, I want dialog-script context marked as an `Action` and separated from spoken-line context, so that I never confuse a scripted display with the reply a character actually says.
7. As a translator, I want a link to the exact dialog node where a scripted action fires, so that I can jump to its place in the conversation.
8. As a translator, I want script context to appear only for lines genuinely referenced by display-family actions, so that the context column stays trustworthy and not noisy.
9. As a translator, I want script context gathered the same way as all other context — by running `text export --context-from all`, so that I do not change my workflow.
10. As a translator, I want to opt into script context alone via `--context-from bcs`, so that I can scope a run.
11. As a localization engineer, I want an `ids show` command that prints an IDS file (raw or as JSONL), so that I can inspect `ACTION.IDS` and verify which functions reference StrRefs.
12. As a localization engineer, I want the StrRef-bearing actions resolved from the current game's own `ACTION.IDS`, so that games that add or reorder actions (e.g. Planescape: Torment) are handled correctly.
13. As a localization engineer, I want decoding to degrade gracefully on scripts it cannot fully parse, so that one malformed resource does not abort the whole export.
14. As a maintainer, I want the StrRef-extraction logic covered by parser-level tests, so that the slot/signature mapping cannot silently regress.

# Implementation Decisions

## New `ids show` command
- A new command group `ids` with a `show` leaf, mirroring `2da show`: raw output by default, `--json` for JSONL. Reuses the existing `ParseIds` parser and `InfinityFs` with an IDS type filter. (Already scaffolded.)

## Identifying StrRef-bearing actions
- A **display-family action** is any `ACTION.IDS` entry whose signature carries a parameter of type integer named `StrRef`, matched **case-insensitively** (`StrRef`, `STRREF`, `Strref`). Triggers are parsed for completeness but never carry a StrRef.
- The action set is resolved from the **current game's** `ACTION.IDS` at run time, not hard-coded — games differ (Torment adds `FloatMessage*` and reorders parameters).
- A signature parser turns the identifier string returned by `ParseIds` (e.g. `DisplayString(O:Object*,I:StrRef*)`) into an ordered parameter list and records, for each StrRef parameter, its **integer-slot index** (the count of integer-typed parameters preceding it). Parameters of other types (`O`, `S`, `P`) do not advance the integer count — this keeps Torment's reordering correct.

## Compiled scripts (`.BCS`)
- A new compiled-script parser/decompiler walks the block structure (`SC`/`CR`/`CO`/`TR`/`RS`/`RE`/`AC`/`OB`), tolerant of the format's glued markers (`100AC`, `151OB`).
- For each action it resolves the action id against the signature table; if the action is in the display-family set it reads the StrRef from the mapped integer slot. The scalar slot layout of an action block is `int0`, point(x,y), `int1`, `int2`, two strings; logical integer slots map to physical fields `int0 → 0`, `int1 → 3`, `int2 → 4`.
- It renders a **decoded call** for that action: action name plus parameters. Object decoding is **pragmatic** — resolve the `OBJECT.IDS` identifier (e.g. `LastTrigger`, `Myself`); honour the convention that an action's first object slot (`o1`) is reserved for the `ActionOverride` target, so a normal single object lives in `o2`; show an object name string when present; fall back to a compact form for exotic EA/RACE/… combinations rather than emitting a full `[EA.GENERAL…]` spec. No trailing `// …` comments.
- Scripts are listed and read through `InfinityFs` (as the existing dialog/creature scans do). `.BS` and other formats are not scanned.

## Source-form actions in dialogs (`DLG`)
- Dialog transition action text is already stored decompiled; it is taken as-is from the existing dialog model (`Transition.Action`).
- The action's function **name** is matched against the signature table; the StrRef is the bare integer argument at the StrRef parameter's positional index. Only integer-literal arguments are linked (string/tilde literals do not exist in stored DLG text).
- This is folded into the existing dialog context pass, which already runs during `text export`. The pass is given the parsed `ACTION.IDS` so it can recognise display-family actions.

## Context model
- Two new context categories are added alongside the existing ones, plus one new label for filtering:
  - **`SCRIPTS`** — keyed by script file name, annotated with usage (e.g. `CTFB.BCS [trap @ AR1234]`), valued by the decoded call. Rendered as a key→lines list, like the existing sections.
  - **`DIALOG SCRIPTS`** — keyed by dialog name, valued by `Action: <call> @ <node-url>`, reusing the same node-URL scheme as the existing dialog context. The `Action:` prefix distinguishes it from the spoken-line `DIALOGS` section.

## Script usage reverse index
- While scanning **areas** (`ARE`: area script, trap/region scripts, door and container scripts, embedded actor script slots) and **creatures** (`CRE`: the five script slots), a reverse index maps each script resref to its usage sites (e.g. `trap @ AR1234`, `override @ CHICKEN.CRE`). These are the only resource types that carry `.BCS` references; items, spells, projectiles, effects, stores, world maps and 2DA do not.
- The index is populated from whatever area/creature scans the selected `--context-from` runs; it is not forced on. With `--context-from all` (the common case) the annotation is always present; with `--context-from bcs` alone the `SCRIPTS` entries show the bare decoded call without a usage annotation.
- **Ordering dependency:** because the `SCRIPTS` value embeds usage at the moment it is created, compiled scripts must be processed **after** areas and creatures. This ordering is enforced in the export flow and called out with a code comment.

## Wiring into `text export`
- `FileType_BCS` is added both as a handled case and to the **default context-type list** so that `--context-from all` includes it; script processing runs last in that list.

# Testing Decisions

A good test here asserts on **external behaviour at the highest stable seam** — given input bytes/text, the extracted StrRefs, decoded calls and resulting context — not on internal parser bookkeeping. The dominant risk is the signature→slot mapping, so the weight of testing sits at the parser seam.

- **Parser seam (primary), following the prior art of the existing 2DA parser tests** (table-style, parsing from an in-memory reader, asserting on the returned data):
  - Signature parsing: an identifier string resolves to the correct parameter list and StrRef integer-slot index, including cases with intervening object/point/string parameters and the Torment reordering (`I:Strref*` before vs after `P:Location*`), proving the by-type counting.
  - Compiled-script extraction: a small BCS source (the `CTFB` shape) yields the expected `{strref, decoded call}` pairs — `14381` with `DisplayString(LastTrigger,14381)` — and ignores actions that are not display-family.
  - Source-action parsing: dialog action text such as `DisplayStringHead(Myself,14381)` yields the expected name and StrRef, and `SetGlobal("v","GLOBAL",1)` yields none.
- **Collection seam:** loading script context into a `TextCollection` produces the expected `Context` and `Labels` entries for the affected StrRefs and leaves untouched ids alone. This is a new test area for the `text` package; the collection is a clean in-memory object that can be populated directly.
- **Usage formatting:** the function that renders a usage site (`trap @ AR1234`, `override @ CHICKEN.CRE`) is unit-tested directly. The full area/creature scan that feeds the index is not mocked.

# Out of Scope

- Inline source scripts in any resource other than `DLG` (none exist — areas/creatures reference scripts by resref or compiled blocks, covered by the `.BCS` scan).
- `.BS` (character) scripts.
- Full Near Infinity-fidelity decompilation of arbitrary object specifications (full `[EA.GENERAL.RACE…]` expansion).
- Extracting context from triggers (no trigger carries a StrRef).
- Decoding string/point/non-StrRef parameters for any purpose beyond rendering the decoded call line.
- Hard-coding the StrRef action set; it is always read from the game's `ACTION.IDS`.
- Clickable node URLs for the `SCRIPTS` (compiled) section — only `DIALOG SCRIPTS` carries a node link.

# Further Notes

- Validated end-to-end against real data: `CTFB.BCS` (BG1) decodes to StrRef `14381` = "Trap Sprung", confirming both the integer-slot mapping and the `o1`-reserved object convention against the Near Infinity decompile.
- `ACTION.IDS` across BG1/BG2/IWD/PST shares the same StrRef actions; PST additionally has `FloatMessage`/`FloatMessageFixed` and reorders `DisplayStringPoint`/`DisplayStringPointLog`, which is why the action set and parameter order are read per game.
- Domain vocabulary for this feature is captured in the root `CONTEXT.md` (StrRef, context source, script context, dialog script context, display-family action, compiled vs source script, script usage).
- No ADR was warranted: the pragmatic-decode and two-section choices are cheaply reversible, and the slot conventions are facts of the format rather than decisions.
