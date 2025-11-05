# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: are
  title: ARE v1
  file-extension: are
  ks-version: "0.11"
  endian: le
  bit-endian: le
doc: |
  ## Area overview
  An area in the Infinity Engine is made up of several files; ARE, TIS, WED, MOS, BMP and, depending
  on the complexity of the area, references to BCS, CRE, ITM, BAM and ACM files. The role of each
  file type is outlined below, based on area AR0001 and standard naming conventions:
   - `AR0001.TIS` contains the area graphics
   - `AR0001.WED` contains the area region information (wall groups, alternate tiles, doors etc.)
   - `AR0001.MOS` contains the area minimap graphics (unused in BGEE)
   - `AR0001.ARE` contains the area definition (area type, animations, containers etc.)
   - `AR0001SR.BMP` contains the area search map
   - `AR0001LM.BMP` contains the area light map
   - `AR0001HT.BMP` contains the area height map
   - `AR0001.BCS` contains the area script

  ## General Description
  The ARE file format describes the content of an area (rather than its visual representation).
  ARE files contain the list of actors, items, entrances and exits, spawn points and other
  area-associated info. The ARE file may contain references to other files, e.g. the list of items
  in a container is stored in the ARE file, however the files themselves are not embedded in the
  ARE file.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/are_v1.htm
seq:
  - id: magic
    contents: "AREA"
  - id: version
    contents: "V1.0"
  - id: area
    doc: resref to WED
    type: strz
    size: 8
    encoding: ASCII
  - id: last_saved
    doc: seconds, real time
    type: u4
  - id: flags
    type: flags
    size: 4
  - id: north_area
    type: strz
    size: 8
    encoding: ASCII
  - id: north_area_flags
    size: 4
    type: neighboring_area_flags
  - id: east_area
    type: strz
    size: 8
    encoding: ASCII
  - id: east_area_flags
    size: 4
    type: neighboring_area_flags
  - id: south_area
    type: strz
    size: 8
    encoding: ASCII
  - id: south_area_flags
    size: 4
    type: neighboring_area_flags
  - id: west_area
    type: strz
    size: 8
    encoding: ASCII
  - id: west_area_flags
    size: 4
    type: neighboring_area_flags
  - id: area_type_flags
    type: area_type_flags
    size: 2
  - id: rain_probability
    type: u2
  - id: snow_probability
    type: u2
  - id: fog_probability
    type: u2
  - id: lightning_probability
    type: u2
  - id: overlay_transparency
    type: b1
  - size: 1
  - id: ofs_actors
    type: u4
  - id: num_actors
    type: u2
  - id: num_regions
    type: u2
  - id: ofs_regions
    type: u4
  - id: ofs_spawn_points
    type: u4
  - id: num_spawn_points
    type: u4
  - id: ofs_entrances
    type: u4
  - id: num_entrances
    type: u4
  - id: ofs_containers
    type: u4
  - id: num_containers
    type: u2
  - id: num_items
    type: u2
  - id: ofs_items
    type: u4
  - id: ofs_vertices
    type: u4
  - id: num_vertices
    type: u2
  - id: num_ambients
    type: u2
  - id: ofs_ambients
    type: u4
  - id: ofs_variables
    type: u4
  - id: num_variables
    type: u4
  - id: ofs_tiled_object_flags
    type: u2
  - id: num_tiled_object_flags
    type: u2
  - id: area_script
    type: strz
    size: 8
    encoding: ASCII
types:
  flags:
    seq:
      - id: save_not_allowed
        type: b1
  neighboring_area_flags:
    seq:
      - id: party_required
        type: b1
      - id: party_enabled
        type: b1
  area_type_flags:
    seq:
      - id: bit0
        type: b1
      - id: bit1
        type: b1
      - id: bit2
        type: b1
      - id: bit3
        type: b1
      - id: bit4
        type: b1
      - id: bit5
        type: b1
      - id: bit6
        type: b1
      - id: bit7
        type: b1
      - id: bit8
        type: b1
      - id: bit9
        type: b1
      - id: bit10
        type: b1
    instances:
      bg_outdoor:
        value: bit0
      bg_day_night:
        value: bit1
      bg_weather:
        value: bit2
      bg_city:
        value: bit3
      bg_forest:
        value: bit4
      bg_dungeon:
        value: bit5
      bg_extended_night:
        value: bit6
      bg_can_rest_indoors:
        value: bit7
      pst_hive:
        value: bit0
      pst_clerks_ward:
        value: bit2
      pst_lower_ward:
        value: bit3
      pst_ravels_maze:
        value: bit4
      pst_baator:
        value: bit5
      pst_rubikon:
        value: bit6
      pst_fortress_of_regrets:
        value: bit7
      pst_curst:
        value: bit8
      pst_carceri:
        value: bit9
      pst_outdoors:
        value: bit10
