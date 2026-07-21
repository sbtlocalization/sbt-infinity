<!--
SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# SBT Infinity Tools

[☑️ Українська](./README.md) | **✅ English**

Here you can find our auxiliary tools for working with Infinity Engine-based games, such as Baldur’s Gate, Baldur’s Gate II, Planescape: Torment, and others.

## Where to Get It

The latest version is available for download [here](https://github.com/sbtlocalization/sbt-infinity/releases/latest).

## Commands and Features

### Working with packed game files

All game resources are packed into `BIF` files on disk. To work with them, we have the following commands:

- `bif types` — list of all supported resource types.
- `bif list` — list of all resources (and which `BIF` file they belong to).
- `bif extract` — extract resources to disk (can be limited by types and names).

### Working with dialogs

- `dialog list` — list all available dialogs.
- `dialog export` — save dialogs in JSON Canvas format.

### Working with text strings

- `text list` — list textual strings (filterable).
- `text export` — save as `.xlsx`.
- `text import` — import translations from `XLSX` back into `TLK` files.
- `text convert` — convert an `XLSX` table to JSON.

### WeiDU format support

- `tra export` — export a `TRA` file to `XLSX` format.
- `tra import` — create a `TRA` file from an `XLSX` table.
- `tra update` — update strings of an existing `TRA` file from `CSV`.

### Working with sound

- `sound export` — export game audio files as WAV or FLAC.

### Other

- `2da show` — view `2DA` tables.
- `ids show` — view `IDS` tables.
- `csv diff` — generate a `CSV` with the difference between two other `CSV` files.
- `extract-bam` and `update-bam` — working [with textures](docs/BAM-update.md).

## License

Without open source contributions from many people, localizing games would be much harder if possible at all. By releasing the code of our small tool, we want to give something back to the community and hope it will be useful to someone. So, all code in this repository is available under the [GPL 3.0](./LICENSES/GPL-3.0-only.txt) license.

<a href="https://github.com/sbtlocalization/sbt-infinity">SBT Infinity Tools</a> © 2025 by <a href="https://sbt.localization.com.ua">SBT Localization</a>
