<!--
SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# SBT Infinity Tools

[☑️ Українська](./README.md) | **✅ English**

Here you can find our auxiliary tools for working with Infinity Engine-based games, such as Baldur’s Gate, Baldur’s Gate II, Planescape: Torment, and others.

Read about texture localization in Planescape: Torment in our [comprehensive article](https://sbt.localization.com.ua/en/article/texture-localization-in-planescape-torment/).

## Building the Project

```
go build
```
🙂

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

## License

Without open source contributions from many people, localizing games would be much harder if possible at all. By releasing the code of our small tool, we want to give something back to the community and hope it will be useful to someone.

So, all code in this repository is available under the [GPL 3.0](./LICENSES/GPL-3.0-only.txt) license, and all graphic assets (except original game resources) are licensed under <a href="https://creativecommons.org/licenses/by-sa/4.0/">CC BY-SA 4.0</a>&nbsp;<img src="https://mirrors.creativecommons.org/presskit/icons/cc.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/by.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/sa.svg" alt="" width="16px" height="16px">

<a href="https://github.com/sbtlocalization/sbt-infinity">SBT Infinity Tools</a> © 2025 by <a href="https://sbt.localization.com.ua">SBT Localization</a>
