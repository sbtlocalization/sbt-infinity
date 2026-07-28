---
adr-status: Accepted
superseded-by:
status: To do
---

# ADR-0003 — Cap file-name labels per resource family

## Context

Labels are the filterable column of the export: a flat set of strings per StrRef, mixing two very different things. Some are **words** the builder chooses — `trap`, `door`, `creature`, `journal`, `script action`, `dialog action` — and there are never many of them (eight at most across all four games). The rest are **identifiers**: resource file names, added with their extension by [[ADR-0001 - Context location representation and file filtering]] §3, plus dialog names, which by long-standing convention carry no extension.

Identifiers have no ceiling, and the tail is heavy. Measured against Baldur's Gate II, counting rows that carry more than ten labels of one kind, and the worst single row:

| kind | rows > 10 | worst row |
| --- | --- | --- |
| `.CRE` | 587 | 159 |
| bare resrefs from verbal constants | 387 | 129 |
| `.ITM` | 155 | 258 |
| `dialog N @ X` | 42 | 17 |
| `.BCS` | 34 | 149 |
| `.SPL` | 30 | 36 |
| `.STO` | 12 | 62 |
| `.ARE` | 4 | 403 |
| word labels | 0 | 8 |

A cell with 403 area names is not a filter aid; it is a wall. And [[ADR-0002 - Key script context by the decoded call and move script usage into labels]] makes it worse on purpose — it moves script attachment into labels, which alone would put 1038 labels on one StrRef.

The second-largest group needs explaining, because it is the one that made a naive fix impossible. A creature carries **verbal constants** — canned lines indexed by `SNDSLOT.IDS` (`SELECT_COMMON`, `MORALE`, `DIALOG_HUNGRY`, …). When the creature scan records them, it also adds the creature's *dialog* resref as a label, so that a translator seeing a bare "Yes?" knows which character says it. That is real value for a named character and pure noise for a generic town guard shared by hundreds of creature files. But those labels are spelled exactly like the dialog names the dialog scan adds — a bare resref, no extension — so nothing about the string itself says where it came from. Cutting one and keeping the other cannot be done by looking at the label.

## Decision

1. **A label may belong to a family**, recorded when the label is added rather than inferred afterwards. Everything that exists today keeps being added as a plain label; only the verbal-constant call site names a family explicitly.
2. **The default family is the file extension**, uppercased: `TRADRAN.CRE` → `CRE`, `AR9711.ARE` → `ARE`. A label with no extension and no explicitly named family — every word label, and every dialog name from the dialog scan — has no family.
3. **Verbal constants use the explicit family `SPEAKER`.** This is what makes the family a stored property instead of a guess: the same string is a cappable speaker hint when the creature scan adds it and an uncappable dialog name when the dialog scan does.
4. **Past ten labels of one family, the whole family is replaced by a marker** — `*.CRE`, `*.ARE`, `*.ITM`, and `many speakers` for `SPEAKER`. All-or-nothing, one threshold for every family.
5. **Labels with no family are never capped.** Word labels cannot overflow. Dialog names are the exception on purpose: a dialog is the primary unit of translatable text in this toolkit, seventeen of them is not a wall, and a translator filtering by dialog expects an exact answer.

## Consequences

- Slicing off a first-N would silently break "show me every line from `AR9711.ARE`" — the filter would return fewer rows with no hint that it lied. Replacing the whole family says out loud that the answer is unavailable, and the marker is information in itself: `*.CRE` means the line is heard all over the game.
- The threshold is a single number with no per-family tuning. Ten is comfortably above what a hand-authored resource produces and comfortably below the generic-content tail; the table above shows no family clustering near the boundary.
- `many speakers` breaks the `*.EXT` pattern because `SPEAKER` is not an extension. Both markers are unambiguous — no real label looks like either.
- The marker is not filterable back to its members. Somebody looking for every line spoken by one generic guard cannot get there through labels, and the context body is where they must go instead.
- Labels stop being a plain set of strings in the model. That is one map value type and one new adder; the alternative — sniffing the origin at render time — is not merely uglier, it is impossible for point 3.
