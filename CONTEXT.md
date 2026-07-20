# SBT Infinity Tools — Context

The ubiquitous language behind the CLI toolkit that processes Infinity Engine
localization resources. This is the vocabulary; the code is the *how*.

> The dCanvas dialogue format has its own context in
> [`dcanvas/docs/CONTEXT.md`](./dcanvas/docs/CONTEXT.md).

## Language

**StrRef**:
A string reference — the numeric id of a TLK entry. Every translatable line is
addressed by its StrRef.
_Avoid_: string id, text id, tlk id.

**Context** (of a StrRef):
The set of places where a StrRef is used in the game, attached to its row in the
exported spreadsheet so a translator knows what the line *is* and where it shows
up. Gathered during `text export` from many resource types.
_Avoid_: usage, references.

**Context source**:
A resource type scanned to build context (dialogs, creatures, areas, items,
scripts, …). Selected with `--context-from`.

### Scripts

**Script context** (`SCRIPTS`):
Context contributed by a **compiled script** (a standalone `.BCS` resource) that
references a StrRef through a display-family action. Surfaced per script file.

**Dialog script context** (`DIALOG SCRIPTS`):
Context contributed by a **source-form action** embedded in a dialog transition
(`DLG`). Always marked as an `Action` to distinguish it from the line a
character actually speaks.

**Display-family action**:
An `ACTION.IDS` function whose signature carries an `I:StrRef*` parameter (e.g.
`DisplayString`, `DisplayStringHead`). These are the only actions that
contribute script context. Triggers never carry a StrRef.
_Avoid_: string action, text action.

**Compiled script** vs **source script**:
A `.BCS` resource stores actions in numeric form (action id + positional integer/
string/object/point slots, resolved against `ACTION.IDS`). A dialog stores the
same actions as already-decompiled source text. The two need different parsing
but yield the same script context.

**Script usage**:
Where a `.BCS` script is attached in the game — an area script, a trap/region
script, a door or container script (`ARE`), or one of a creature's script slots
(`CRE`). Built as a reverse index so a script file (e.g. `CTFB.BCS`) can be
labelled with a meaningful site (e.g. `trap @ AR1234`).
_Avoid_: script reference, caller.
