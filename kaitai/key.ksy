# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: @definitelythehuman
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: key
  file-extension: key
  endian: le
  bit-endian: le
doc: |
  This file format acts as a central reference point to locate files required by the game (in a BIFF file on a CD or in
  the `override` directory). The key file also maintains a mapping from an 8 byte resource name (refref) to a 32 byte ID
  (using the lowest 12 bits to identify a resource). There is generally only one key file with each game (`chitin.key`).

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
      - id: ofs_file_path
        type: u4
      - id: len_file_path
        type: u2
      - id: location_bits
        type: location
        size: 2
    instances:
      file_path:
        type: strz
        encoding: ASCII
        io: _root._io
        pos: ofs_file_path
        size: len_file_path
    types:
      location:
        seq:
          - id: in_data
            type: b1
          - id: in_cache
            type: b1
          - id: cd
            type: b1
            repeat: expr
            repeat-expr: 6
  res_entry:
    seq:
      - id: name
        type: strz
        encoding: ASCII
        size: 8
      - id: type
        type: u2
        enum: res_type
      - id: locator
        type: locator(false)
    types:
      locator:
        params:
          - id: in_biff
            type: bool
        seq:
          - id: file_index
            type: b14
          - id: tileset_index
            type: b6
          - id: biff_file_index
            type: b12
        instances:
          biff_file:
            pos: _root.ofs_biff_entries.as<u8> + biff_file_index * 12  # size of biff_entry
            type: biff_entry
            io: _root._io
            if: not in_biff
enums:
  res_type:
    0x001: bmp
    0x002: mve
    0x004: wav
    0x005: wfx
    0x006: plt
    0x3b8: tga
    0x3e8: bam
    0x3e9: wed
    0x3ea: chu
    0x3eb: tis
    0x3ec: mos
    0x3ed: itm
    0x3ee: spl
    0x3ef: bcs
    0x3f0: ids
    0x3f1: cre
    0x3f2: are
    0x3f3: dlg
    0x3f4: two_da
    0x3f5: gam
    0x3f6: sto
    0x3f7: wmp
    0x3f8: eff
    0x3f9: bs
    0x3fa: chr
    0x3fb: vvc
    0x3fc: vef
    0x3fd: pro
    0x3fe: bio
    0x3ff: wbm
    0x400: fnt
    0x402: gui
    0x403: sql
    0x404: pvrz
    0x405: glsl
    0x406: tot
    0x407: toh
    0x408: menu
    0x409: lua
    0x40a: ttf
    0x40b: png
    0x44c: bah
    0x802: ini
    0x803: src
    0x804: maze
    0xffe: mus
    0xfff: acm
