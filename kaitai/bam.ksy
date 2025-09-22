# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: bam
  title: BAM v2
  file-extension: bam
  ks-version: "0.11"
  endian: le
doc: |
  The BAM v2 file format is used to store animation data for the Infinity Engine games.
  Such files are used for animations (both creature animations, item and spell animations)
  and interactive GUI elements (e.g. buttons) and for logical collections of images (e.g. fonts).
  BAM files can contain multiple sequences of animations, up to a limit of 255.

  The BAM v2 format is used in games like Baldur's Gate, Icewind Dale, and Planescape: Torment.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/bam_v2.htm
seq:
  - id: header
    type: header
instances:
  frame_entries:
    pos: _root.header.frame_entries_offset
    type: frame_entries
  cycle_entries:
    pos: _root.header.cycle_entries_offset
    type: cycle_entries
  data_blocks:
    pos: _root.header.data_blocks_offset
    type: data_blocks
types:
  header:
    seq:
      - id: magic
        contents: "BAM "
      - id: version
        contents: "V2  "
      - id: frame_count
        type: u4
      - id: cycle_count
        type: u4
      - id: data_blocks_count
        type: u4
      - id: frame_entries_offset
        type: u4
      - id: cycle_entries_offset
        type: u4
      - id: data_blocks_offset
        type: u4
  frame_entry:
    seq:
      - id: width
        type: u2
      - id: height
        type: u2
      - id: center_x
        type: s2
      - id: center_y
        type: s2
      - id: data_blocks_start_index
        type: u2
      - id: data_blocks_count
        type: u2
  cycle_entry:
    seq:
      - id: frame_count
        type: u2
      - id: frame_entries_start_index
        type: u2
  data_block:
    seq:
      - id: prvz_page
        type: u4
      - id: source_x
        type: u4
      - id: source_y
        type: u4
      - id: width
        type: u4
      - id: height
        type: u4
      - id: target_x
        type: u4
      - id: target_y
        type: u4
  frame_entries:
    seq:
      - id: entry
        type: frame_entry
        repeat: expr
        repeat-expr: _root.header.frame_count
  cycle_entries:
    seq:
      - id: entry
        type: cycle_entry
        repeat: expr
        repeat-expr: _root.header.cycle_count
  data_blocks:
    seq:
      - id: block
        type: data_block
        repeat: expr
        repeat-expr: _root.header.data_blocks_count
