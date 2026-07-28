---
adr-status: Accepted
superseded-by:
status: To do
---

# ADR-0002 — Key script context by the decoded call and move script usage into labels

## Context

Script context shipped keyed by *where* the call was found: `SCRIPTS` keyed each entry by the script file annotated with its usage (`DUADRAN.BCS [general @ TRADRAN.CRE]`) and put the decoded call in the value; `DIALOG SCRIPTS` keyed by dialog and put `Action: <call> @ <node-url>` in the value. Both shapes repeat the *same* call once per place, and the repetition is the common case, not the exception:

- A generic line floated by ten different creature scripts produced ten near-identical entries, each restating `DisplayStringHead(Myself,40811)`.
- Dialog entries fanned out a second time, because one `DLG` file yields one dialog tree per root state — `DKRESSEL[0]`, `[1]`, `[28]` are three renderings of the *same* transition, so a single `SetQuestDone` produced three lines with three URLs differing only in the page they point at.

Meanwhile the part a translator actually needs — that `CTFB.BCS` is a *trap* in `AR1234` — was buried inside a key it could not be filtered by, and capped at five sites with a `+N more` tail.

## Decision

1. **Both script sections are keyed by the decoded call**; the values list the places it comes from — script files for `SCRIPTS`, dialog names for `DIALOG SCRIPTS`. Repetition of an identical call collapses into one line. Values render through the shared inline/list helper, so `--long` widens the list and changes nothing else.
2. **Script usage moves out of the context body and into labels**: the resource where the script is attached (with its extension — `TRADRAN.CRE`, `AR9711.ARE`) plus the `ARE` slot it sits in (`trap`, `door`, `container`, `info point`, …), or `creature` for any of the five `CRE` slots, whose differences are engine evaluation order and carry nothing for a translator. The attachment is therefore filterable, which it never was inside a key.
3. **Script file names are dropped from labels** and live only in the context body. A `.BCS` name is opaque to a translator and useless as a filter, whereas the resources it is attached to are exactly what one filters by. This is a deliberate exception to ADR-0001 §3.
4. **Nothing about the resulting volume is specific to scripts.** Measured against BG2, the worst StrRef would gain 1038 usage-place labels, because `GPSHOUT.BCS` alone is attached to 408 places and shouts a handful of lines. But a shout script attached to 408 creatures produces plain `.CRE` labels, indistinguishable from the ones the creature scan already produces, and items and areas overflow the same way without any script involved. The ceiling therefore belongs to the general label mechanism of [[ADR-0003 - Cap file-name labels per resource family]], not here; this decision only feeds it more labels.
5. **`AddJournalEntry` is the one action whose context keeps node URLs**, matched by name. It is where a journal line *enters* the game, so the conversation branch that adds it is the situation being translated; every other action either merely displays a line (the dialog name locates it well enough) or references a line introduced elsewhere (`EraseJournalEntry`, `SetQuestDone`). Its entries sort ahead of the rest inside `DIALOG SCRIPTS`.
6. **A dialog tree is a dialog.** One `DLG` file yields one tree per root state, and those trees are treated as separate dialogs — so an action reachable from three entry points keeps all three node URLs, and the same fan-out in the spoken-line `DIALOGS` section is left exactly as it is. The repetition there is a faithful reflection of the model, not the defect this ADR fixes.
7. **Dialog names stay extension-less** in both context and labels, unlike every other resource. The whole dialog subsystem — node URLs, dcanvas pages, the existing dialog labels — has always addressed dialogs without `.DLG`, and that spelling is load-bearing for the published dialog site.

## Consequences

- A widely shared script now shows up as `*.CRE` rather than as a name, so the exported line says "heard all over the game" instead of naming one arbitrary creature. That marker is produced by the general guard, which this decision depends on: shipping the usage labels without it would put four-figure label cells into the spreadsheet.
- Where a script is attached is no longer visible in the context body at all, not even under `--long`. The labels column is the only place it lives, which is a real loss for anyone verifying a single line by hand.
- Naming `AddJournalEntry` in code is a deviation from EPIC-0001's "no hard-coded action set". That constraint stands for the display-family *set*, which is still derived from the game's own `ACTION.IDS`; it does not extend to a presentation rule about one action. The name is stable across BG1, BG2, IWD and PST (action id 173, identical signature).
- `AddMapNote` and `AddMapNoteColor` create a lasting player-facing record too, and were deliberately left out of the URL tier: map-note text is short and self-explanatory, so the node adds little.
