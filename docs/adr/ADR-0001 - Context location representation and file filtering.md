---
adr-status: Accepted
superseded-by:
status: Done
---

# ADR-0001 — Context location representation and file filtering

## Context

Each **StrRef** exported by `text export` carries its **Context** — the set of places the line is used — in a spreadsheet cell, so a translator understands what the line *is* and where it appears.

For **Context sources** that reference a StrRef deep inside a file (spell/item effects, area triggers, world-map areas, store drinks), the location was rendered as a full internal path: `Spell MISLEAD.SPL → ability 0 → effect 20`. Two problems followed:

1. **The cell overflows.** A single message reused across many resources — very common for effect messages like "Protection from spell" — produced dozens of near-identical lines, because the `ability`/`effect` indices differ per line and defeat de-duplication.
2. **The indices are useless to the reader.** A translator cannot map "ability 0 → effect 20" to anything in the running game; only the source *file* is a meaningful anchor.

Separately, the spreadsheet was not filterable by source file: only dialogs recorded their file name (as a label), so "show me every line from `MISLEAD.SPL`" was not answerable for other resource types.

## Decision

1. **Collapse location values to the source file name.** For the `file → index` sources (creature/item/spell effects, world maps, areas, stores), the Context value keeps only the bare file name; the internal indices and the `Spell`/`Item`/`Creature` prefixes are dropped. De-duplication then folds the repeats into the distinct set of source files.
2. **Keep full detail behind a `--long` flag** (default off). It restores the verbose per-index strings verbatim, as an escape hatch, without affecting labels or the tracking reformat.
3. **Add the source file name as a label** for every StrRef gathered from a concrete file. Because labels are a de-duplicated set, this makes the sheet filterable by file. The label is the raw filesystem spelling (with extension), so filtering is consistent regardless of how a context body happens to render the name.

Sources whose value is not a `file → index` path are left as-is: standalone `.EFF` (already a bare file name), `UI` (the file name is the key; the value is the in-file `window N → control N` coordinate), the tracking default-embed value, and all `2DA` values (row keys / slots / screen images that carry real meaning). The ranger Tracking area value is reformatted to `<rowKey>.ARE` unconditionally, since the row key already denotes an area resource.

## Consequences

- Effect-heavy Context cells shrink from dozens of lines to a short, sorted, de-duplicated file list — the intended readability win.
- The exact reference site inside a file is no longer visible by default; anyone who needs it (debugging, verification) uses `--long`.
- The `labels` column grows by the set of source files per StrRef, but since it is a set the growth is bounded by the (small) number of distinct files.
- Export output is now sorted and deterministic for these values, so regenerating the spreadsheet produces clean diffs.
