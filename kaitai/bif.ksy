# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: bif
  file-extension: bif
  endian: le
  bit-endian: le
doc: |
  General Description
  This file format is a simple archive format, used mainly both to simplify organization of the files by grouping logically related files together (especially for areas). There is also a gain from having few large files rather than many small files, due to the wastage in the FAT and NTFS file systems. BIF files containing areas typically contain:
  * one or more WED files, detailing tiles and wallgroups
  * one or more TIS files, containing the tileset itself
  * one or more MOS files, containing the minimap graphic
  * 3 or 4 bitmap files which contain one pixel for each tile needed to cover the region

  The bitmaps are named xxxxxxHT.BMP, xxxxxxLM.BMP, xxxxxxSR.BMP and optionally xxxxxxLN.BMP.
  * xxxxxxHT.BMP: Height map, detailing altitude of each tile cell in the associated wed file
  * xxxxxxLM.BMP: Light map, detailing the level and colour of illumination each tile cell on the map. Used during daytime
  * xxxxxxLN.BMP: Light map, detailing the level and colour of illumination each tile cell on the map. Used during night-time
  * xxxxxxSR.BMP: Search Map, detailing where characters cannot walk, and the footstep sounds

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
    pos: ofs_file_entries
    repeat: expr
    repeat-expr: num_tileset_entries

    
types:
  file_entry:
    seq:
      - id: res_locator
        type: u4
      - id: res_offset
        type: u4
      - id: len_res_blob
        type: u4
      - id: res_type
        type: u2
      - id: unknown
        type: u2
    instances:
      res_blob:
        pos: res_offset
        size: len_res_blob
      file_extension:
        pos: res_offset
        type: str
        encoding: ASCII
        terminator: 0
        size: 4

        
  tileset_entry:
    seq:
      - id: tls_locator
        type: u4
      - id: tls_offset
        type: u4
      - id: tls_count
        type: u4
      - id: tls_len
        type: u4
      - id: tls_type
        type: u2
      - id: unknown
        type: u2
