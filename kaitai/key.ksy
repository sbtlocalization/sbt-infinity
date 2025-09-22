# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: key
  file-extension: key
  endian: le
  bit-endian: le
doc: |
  This file format acts as a central reference point to locate files required
  by the game (in a BIFF file on a CD or in the override directory). The key
  file also maintains a mapping from an 8 byte resource name (refref) to a
  32 byte ID (using the lowest 12 bits to identify a resource). There is
  generally only one key file with each game (chitin.key).
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/key_v1.htm
seq:
  - id: magic
    contents: "KEY "
  - id: version
    contents: "V1  "
  - id: num_biff_entries
    type: u4
  - id: num_res_entries
    type: u4
  - id: ofs_biff_entries
    type: u4
  - id: ofs_res_entries
    type: u4
instances:
  biff_entries:
    type: biff_entry
    pos: _root.ofs_biff_entries
    repeat: expr
    repeat-expr: num_biff_entries
  res_entries:
    type: res_entry
    pos: _root.ofs_res_entries
    repeat: expr
    repeat-expr: num_res_entries
types:
  biff_entry:
    seq:
      - id: len_file
        type: u4
      - id: ofs_file_name
        type: u4
      - id: len_file_name
        type: u2
      - id: location_bits
        type: location
        size: 2
    instances:
      file_name_ext:
        type: strz
        encoding: ASCII
        io: _root._io
        pos: ofs_file_name
        size: len_file_name
    types:
      location:
        seq:
          - id: in_data
            type: b1
          - id: in_cache
            type: b1
          - id: cd1
            type: b1
          - id: cd2
            type: b1
          - id: cd3
            type: b1
          - id: cd4
            type: b1
          - id: cd5
            type: b1
          - id: cd6
            type: b1
  res_entry:
    seq:
      - id: name
        type: strz
        encoding: ASCII
        size: 8
      - id: type
        type: u2
      - id: locator
        type: locator
    types:
      locator:
        seq:
          - id: file_index
            type: b14
          - id: tileset_index
            type: b6
          - id: biff_file_index
            type: b12
        instances:
          biff_file:
            pos: _root.ofs_biff_entries + biff_file_index * 12
            type: biff_entry
            io: _root._io
