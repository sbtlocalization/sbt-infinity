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
Context contributed by a **compiled script** (a standalone `.BCS` resource) that references a StrRef through a display-family action. Surfaced per decoded call, listing the scripts that make it.

**Dialog script context** (`DIALOG SCRIPTS`):
Context contributed by a **source-form action** embedded in a dialog transition (`DLG`). Surfaced per decoded call, listing the dialogs that make it, so it is never mistaken for the line a character actually speaks.

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
Where a `.BCS` script is attached in the game — an area script, a trap/region script, a door or container script (`ARE`), or one of a creature's script slots (`CRE`). Built as a reverse index and recorded as labels on every StrRef the script displays, so a line can be filtered by the resource it is attached to (e.g. `AR1234.ARE`, `trap`).
_Avoid_: script reference, caller, site.

### Labels

**Label**:
A short filterable tag on a StrRef, sitting beside the context in its own spreadsheet column. Either a word naming what the line *is* (`trap`, `creature`, `journal`, `script action`) or an identifier naming a resource it comes from (`AR9711.ARE`, `TRADRAN.CRE`).
_Avoid_: tag, marker, category.

**Label family**:
The group an identifier label is counted in when deciding whether there are too many of them — normally its file extension (`CRE`, `ARE`, `ITM`), occasionally named outright where the spelling does not give it away. Word labels and dialog names belong to no family and are always shown in full.

**Verbal constant**:
One of a creature's canned lines, indexed by `SNDSLOT.IDS` (a greeting, a battle cry, a reaction to hunger) rather than spoken in a dialog. Labelled with the creature's dialog so a translator can tell who the voice belongs to.
_Avoid_: sound slot, creature sound.
