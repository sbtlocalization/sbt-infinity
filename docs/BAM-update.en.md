<!--
SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# Texture localization

[☑️ Українська](./BAM-update.md) | **✅ English**

Read about texture localization in Planescape: Torment in our [comprehensive article](https://sbt.localization.com.ua/en/article/texture-localization-in-planescape-torment/).

## Extracting frames from BAM v2 animation files

Animations in the Enhanced Editions of the aforementioned games are stored in the BAM v2 format. This is a binary format that contains frame data, where each frame consists of one or more blocks. These blocks are stored in separate `mosXXXX.PVRZ` files, where `XXXX` is a four-digit file index.

These PVRZ files are often built as texture atlases in which a single final image is split into many pieces for more efficient storage.

To export all individual frames from a BAM animation as PNG files, use the extract-bam command:
```
./sbt-inf extract-bam path/to/config.toml
```

### Configuration

The `config.toml` file contains information about which input files to process and where to save the results. For example:
```toml
[Input]
bam = "animation.BAM"

[InputMos]
1000 = "mos2000.PNG"
1001 = "mos2001.PNG"

[Output]
extract = "output"
```

The `bam` line points to the original animation, and the `InputMos` section lists all PVRZ files which BAM file depend to, exported as PNGs. To extract both, it’s best to use a tool called [Near Infinity](https://github.com/NearInfinityBrowser/NearInfinity/wiki).

The `extract` value is the path to the folder where the extracted frames will be saved.

All paths can be either absolute or relative (to the location of the TOML file itself).

## Updating texture atlases for animations

Despite its name, the `update-bam` command does not update the animation itself, and instead updates the textures associated with it. (If you have a better name in mind, feel free to make a PR 😉).

For this, you also need to specify a list of updated (redrawn) frames in the configuration file:

```toml
[Input]
bam = "animation.BAM"

[InputMos]
1000 = "mos2000.PNG"
1001 = "mos2001.PNG"

[NewFrames]
1 = "Frame1.png"
6 = "Frame6.png"
7 = "Frame7.png"

[Output]
update = "override"
```

Note that it is not necessary to specify all frames. If some are omitted, the program will instead use blocks from the “original” textures specified in InputMos.

Usage of the command is similar to the previous one:

```
./sbt-inf update-bam path/to/config.toml
```

## Examples

You can find our configuration files and redrawn frames for **Planescape: Torment: Enhanced Edition** in the [`inputs`](./inputs/) folder.
