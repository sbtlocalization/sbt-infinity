# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: @definitelythehuman
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: bif
  file-extension: bif
  endian: le
  bit-endian: le
  imports:
    - key
doc: |
  This file format is a simple archive format, used mainly both to simplify organization of the files by grouping
  logically related files together (especially for areas). There is also a gain from having few large files rather than
  many small files, due to the wastage in the FAT and NTFS file systems. BIF files containing areas typically contain:
  - one or more WED files, detailing tiles and wallgroups;
  - one or more TIS files, containing the tileset itself;
  - one or more MOS files, containing the minimap graphic;
  - 3 or 4 bitmap files which contain one pixel for each tile needed to cover the region.

  The bitmaps are named `xxxxxxHT.BMP`, `xxxxxxLM.BMP`, `xxxxxxSR.BMP` and optionally `xxxxxxLN.BMP`.
  - `xxxxxxHT.BMP`: Height map, detailing altitude of each tile cell in the associated wed file.
  - `xxxxxxLM.BMP`: Light map, detailing the level and colour of illumination each tile cell on the map. Used during
  daytime.
  - `xxxxxxLN.BMP`: Light map, detailing the level and colour of illumination each tile cell on the map. Used during
  night-time.
  - `xxxxxxSR.BMP`: Search Map, detailing where characters cannot walk, and the footstep sounds.

  ## Overall structure

    - Header
    - File entries
    - Tileset entries
    - Data for the contained files, as described in the file and tileset entries

  Note that the data of the contained files might be after the header, and the file/tileset entries after the contained
  files. The offset to the file entries and the offset to each contained file should be used to know the exact location
  of each.

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/bif_v1.htm
seq:
  - id: magic
    contents: "BIFF"
  - id: version
    contents: "V1  "
  - id: num_file_entries
    type: u4
  - id: num_tileset_entries
    type: u4
  - id: ofs_file_entries
    type: u4

instances:
  file_entries:
    type: file_entry
    pos: ofs_file_entries
    repeat: expr
    repeat-expr: num_file_entries
  tileset_entries:
    type: tileset_entry
    pos: ofs_file_entries + num_file_entries * 16  # size of file_entry
    repeat: expr
    repeat-expr: num_tileset_entries

types:
  file_entry:
    doc-ref: https://gibberlings3.github.io/iesdp/file_formats/ie_formats/bif_v1.htm#bif_v1_FileEntry
    seq:
      - id: locator
        type: key::res_entry::locator(true)
      - id: ofs_data
        type: u4
      - id: len_data
        type: u4
      - id: res_type
        type: u2
        enum: key::res_type
      - id: reserved
        type: u2
    instances:
      data:
        pos: ofs_data
        size: len_data

  tileset_entry:
    doc-ref: https://gibberlings3.github.io/iesdp/file_formats/ie_formats/bif_v1.htm#bif_v1_TilesetEntry
    seq:
      - id: locator
        type: key::res_entry::locator(true)
      - id: ofs_data
        type: u4
      - id: num_tiles
        type: u4
      - id: len_tile
        type: u4
      - id: res_type
        type: u2
        enum: key::res_type
        valid: key::res_type::tis
      - id: reserved
        type: u2
    instances:
      data:
        pos: ofs_data
        size: num_tiles * len_tile
