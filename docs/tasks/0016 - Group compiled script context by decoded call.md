---
epic: "[[EPIC-0002 - Context and label readability in text export]]"
parent:
status: Done
priority: 3
blocked by:
  - "[[0018 - Cap file-name labels per resource family]]"
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

Turn the `SCRIPTS` section of the exported context inside out, and move **script usage** from the context body into labels. Decided in [[ADR-0002 - Key script context by the decoded call and move script usage into labels]].

Today the section is keyed by the script file annotated with its usage and carries the decoded call as the value, so one generic line floated by ten scripts produces ten near-identical entries that restate the same call. Key each entry by the **decoded call** instead and list the scripts that make it, de-duplicated and sorted, through the same inline/list helper the other sections use — so the default export caps the inline list with a `(+N more)` tail and `--long` widens it. Nothing else about `--long` changes: the attachment of a script never returns to the context body, in either mode.

Script usage becomes labels on every StrRef the script displays: the resource the script is attached to, spelled with its extension, plus the `ARE` slot it sits in (`trap`, `door`, `container`, `info point`, `travel`, `region`, `area`, `actor`). The five creature script slots differ only in engine evaluation order, so they all collapse to the single `creature` label. Area attachments currently lose their extension — they must read `AR9711.ARE`, like every other resource label. The action's own name (as spelled in `ACTION.IDS`) becomes a label too; `script action` stays; the `.BCS` file name stops being a label, since it is opaque to a translator and useless as a filter.

Generic scripts need a ceiling — `GPSHOUT.BCS` alone is attached to 408 places, and the worst StrRef in Baldur's Gate II would otherwise gain 1038 labels. Nothing about that is script-specific, though: the labels are ordinary `.CRE` and `.ARE` names, and [[0018 - Cap file-name labels per resource family]] already caps those. This task must land after it, and adds no threshold of its own.

Consequently the usage index has to keep the slot and the file as separate parts rather than one pre-joined string, and the helpers that only existed to build the old annotated key retire along with their tests. The existing ordering constraint stands and now serves the labels: compiled scripts must be scanned after areas and creatures, or the index is empty when the labels are written.

# Acceptance criteria

- [x] A line floated by several scripts renders as a single entry keyed by the decoded call, listing the scripts that make it; identical calls no longer repeat per file.
- [x] Entries with different object arguments stay separate, so it remains visible whether a line floats over `Myself` or over `LastTrigger`.
- [x] Script lists are de-duplicated and sorted; the default export caps the inline list and `--long` shows every script.
- [x] No usage annotation appears in the context body in either mode.
- [x] Every StrRef displayed by a script is labelled with the resources the script is attached to, each carrying its extension — areas included (`AR9711.ARE`).
- [x] `ARE` slots appear as labels under their own names; all five creature script slots produce `creature` and nothing more.
- [x] The action name is added as a label alongside `script action`; `.BCS` file names are no longer labels.
- [x] A widely attached script produces the ordinary per-family marker (`*.CRE`) through the existing label cap, with no script-specific rule anywhere in this change.
- [x] Running with `--context-from bcs` alone still works, producing calls and action-name labels without any usage labels.
- [x] The retired usage-key helpers and their tests are gone, while the usage index still ignores empty and `NONE`/`*`/`0` script references and still names region slots correctly.
- [x] `go test ./...` passes.
