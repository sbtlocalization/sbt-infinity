---
epic: "[[EPIC-0002 - Context and label readability in text export]]"
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

Turn the `DIALOG SCRIPTS` section inside out the same way as the compiled-script section, and give journal entries the only node links in it. Decided in [[ADR-0002 - Key script context by the decoded call and move script usage into labels]].

Today the section is keyed by dialog and every value repeats the call together with a node URL, which fans out twice over: once per dialog that fires the action, and again because a single `DLG` file yields one dialog tree per root state, so the *same* transition surfaces as `DKRESSEL[0]`, `[1]`, `[28]` with three URLs differing only in the page they point at. A quest line touched by `SetQuestDone` and `EraseJournalEntry` therefore fills the cell with links no translator needs.

Key each entry by the **decoded call** and list the dialogs that make it, de-duplicated and sorted, inline through the shared helper. Dialog names stay without the `.DLG` extension, matching how the dialog subsystem and its published pages have always addressed them, and matching the existing dialog labels. The `Action:` prefix goes away: a function call with parentheses cannot be mistaken for a spoken reply, and the section heading plus the `dialog action` label already say so twice.

`AddJournalEntry` is the exception and keeps its node URLs — it is where a journal line enters the game, so the branch of the conversation that grants it is the situation being translated. All of its URLs are kept, since each root state is formally a separate dialog, and they render one per line rather than crammed inline. Its entries sort ahead of every other call in the section, so the thing worth reading comes first. Every other display-family action — including ones outside the handful seen in Icewind Dale, such as `AddMapNoteColor`, `SetPlayerSound`, `SetName` or Torment's `FloatMessage` — falls into the grouped, URL-less form without needing to be enumerated anywhere.

Labels gain the action's own name for every dialog action, and `journal` for `AddJournalEntry`; the existing `dialog action` label and the dialog name label stay. Since the section now needs two renderings in one block, it gets its own join function, following the precedent of the spoken-dialog section, which already has one.

# Acceptance criteria

- [x] `SetQuestDone` and `EraseJournalEntry` collapse to one entry per call, listing the dialogs that fire it — no repetition per dialog tree and no node URLs.
- [x] `AddJournalEntry` entries keep every node URL, rendered one per line, and sort ahead of all other calls in the section.
- [x] Entries with different object arguments stay separate, so it remains visible whom a floating line is shown over.
- [x] The `Action:` prefix is gone from every value.
- [x] Dialog names carry no `.DLG` extension, in both the context body and the labels.
- [x] Every dialog action adds its action name as a label; `AddJournalEntry` also adds `journal`; `dialog action` and the dialog name label are unchanged.
- [x] Actions outside the display family, and display-family calls whose StrRef argument is not an integer literal, are still ignored.
- [x] `go test ./...` passes.
