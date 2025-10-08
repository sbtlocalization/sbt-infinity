# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: wmp
  title: WMAP v1
  file-extension: wmp
  ks-version: "0.11"
  endian: le
  bit-endian: le
doc: |
  This file format describes the top-level map structure of the game. It details the x/y coordinate
  location of areas, the graphics used to represent the area on the map (both MOS and BAM ) and
  stores flag information used to decide how the map icon is displayed (visable, reachable, already
  visited etc.)

  ## Engine specific notes:
  Areas may be also displayed on the WorldMap in ToB using 2DA files:
  - `XNEWAREA.2DA` (Area entries section of wmp )
  - 2DA file specified in `XNEWAREA.2DA` (Area links section) for example `XL3000.2DA`

  NB. A WMP file must have at least one area entry, and one area link to be considered valid.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/wmap_v1.htm
seq:
  - id: magic
    contents: "WMAP"
  - id: version
    contents: "V1.0"
  - id: num_worldmap_entries
    type: u4
  - id: ofs_worldmap_entries
    type: u4
instances:
  worldmap_entries:
    pos: ofs_worldmap_entries
    type: worldmap_entry
    repeat: expr
    repeat-expr: num_worldmap_entries
types:
  worldmap_entry:
    seq:
      - id: background_image_mos
        type: strz
        size: 8
        encoding: ASCII
      - id: width
        type: u4
      - id: height
        type: u4
      - id: map_number
        type: u4
      - id: area_name_ref
        type: u4
      - id: center_x
        type: u4
      - id: center_y
        type: u4
      - id: num_area_entries
        type: u4
      - id: ofs_area_entries
        type: u4
      - id: ofs_area_link_entries
        type: u4
      - id: num_area_link_entries
        type: u4
      - id: map_icons_bam
        type: strz
        size: 8
        encoding: ASCII
      - id: flags
        size: 4
        type: flags
      - id: reserved
        size: 124
    instances:
      area_entries:
        pos: ofs_area_entries
        type: area_entry
        repeat: expr
        repeat-expr: num_area_entries
    types:
      flags:
        seq:
          - id: colored_icons
            type: b1
          - id: ignore_palette
            type: b1
  area_entry:
    seq:
      - id: area
        type: strz
        size: 8
        encoding: ASCII
      - id: short_name
        type: strz
        size: 8
        encoding: UTF-8
      - id: long_name
        type: strz
        size: 32
        encoding: UTF-8
      - id: status
        type: status
        size: 4
      - id: icons_bam_sequence
        type: u4
      - id: x
        type: u4
      - id: y
        type: u4
      - id: caption_ref
        type: u4
      - id: tooltip_ref
        type: u4
      - id: loading_screen_mos
        type: strz
        size: 8
        encoding: ASCII
      - id: north_link_index
        type: u4
      - id: num_north_links
        type: u4
      - id: west_link_index
        type: u4
      - id: num_west_links
        type: u4
      - id: south_link_index
        type: u4
      - id: num_south_links
        type: u4
      - id: east_link_index
        type: u4
      - id: num_east_links
        type: u4
      - id: reserved
        size: 128
    instances:
      north_links:
        pos: _parent.ofs_area_link_entries + north_link_index * 216 # size of area_link_entry
        type: area_link_entry
        repeat: expr
        repeat-expr: num_north_links
      south_links:
        pos: _parent.ofs_area_link_entries + south_link_index * 216 # size of area_link_entry
        type: area_link_entry
        repeat: expr
        repeat-expr: num_south_links
      east_links:
        pos: _parent.ofs_area_link_entries + east_link_index * 216 # size of area_link_entry
        type: area_link_entry
        repeat: expr
        repeat-expr: num_east_links
      west_links:
        pos: _parent.ofs_area_link_entries + west_link_index * 216 # size of area_link_entry
        type: area_link_entry
        repeat: expr
        repeat-expr: num_west_links
    types:
      status:
        seq:
          - id: visible
            type: b1
          - id: visible_from_adjacent
            type: b1
          - id: reachable
            type: b1
          - id: visited
            type: b1
  area_link_entry:
    seq:
      - id: destination_area_index
        type: u4
      - id: entry_point
        type: strz
        size: 32
        encoding: ASCII
      - id: travel_time
        type: u4
      - id: default_entrance
        type: u4
        enum: entrance
      - id: random_encounter_area_1
        type: strz
        size: 8
        encoding: ASCII
      - id: random_encounter_area_2
        type: strz
        size: 8
        encoding: ASCII
      - id: random_encounter_area_3
        type: strz
        size: 8
        encoding: ASCII
      - id: random_encounter_area_4
        type: strz
        size: 8
        encoding: ASCII
      - id: random_encounter_area_5
        type: strz
        size: 8
        encoding: ASCII
      - id: random_encounter_probability
        type: u4
      - id: reserved
        size: 128
    enums:
      entrance:
        0x01: north
        0x02: east
        0x04: south
        0x08: west
