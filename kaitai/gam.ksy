# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: @definitelythehuman
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: gam
  file-extension: gam
  endian: le
  bit-endian: le
doc: |
  This file format is used to hold game information in save games.
  The GAM file does not store area, creature or item information, instead, it stores information on the party members and the global variables which affect party members.

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/gam_v2.0.htm
seq:
  - id: magic
    contents: "GAME"
  - id: version
    contents: "V2.0"
  - id: game_time
    type: u4
  - id: selected_formation
    type: u2
  - id: formation_button
    type: u2
    repeat: expr
    repeat-expr: 5
  - id: party_gold
    type: u4
  - id: use_active_area
    type: s2
    enum: party_member
  - id: weather_bitfield
    type: weather_flags
    size: 2
  - id: ofs_party_members
    type: u4
  - id: num_party_members
    type: u4
  - id: ofs_party_inventory
    type: u4
  - id: num_party_inventory
    type: u4
  - id: ofs_non_party_members
    type: u4
  - id: num_non_party_members
    type: u4
  - id: ofs_global_variables
    type: u4
  - id: num_global_variables
    type: u4
  - id: main_area
    type: strz
    encoding: ASCII
    size: 8
  - id: ofs_familiar_extra
    type: u4
  - id: num_journal_entries
    type: u4
  - id: ofs_journal_entries
    type: u4
  - id: party_reputation
    type: u4
  - id: current_area
    type: strz
    encoding: ASCII
    size: 8
  - id: gui_flags
    type: gui_flags
    size: 4
  - id: loading_progress
    type: u4
    enum: loading_progress
  - id: ofs_familiar_info
    type: u4
  - id: ofs_stored_locations
    type: u4
  - id: num_stored_locations
    type: u4
  - id: game_time_seconds
    type: u4
  - id: ofs_pocket_plane_locations
    type: u4
  - id: num_pocket_plane_locations
    type: u4
  - id: zoom_level
    type: u4
  - id: random_encounter_area
    type: strz
    encoding: ASCII
    size: 8
  - id: current_worldmap
    type: strz
    encoding: ASCII
    size: 8
  - id: current_campagin
    type: strz
    encoding: UTF-8
    size: 8
  - id: familiar_owner
    type: u4
    enum: party_member
  - id: random_encounter_entry
    type: strz
    encoding: ASCII
    size: 20

instances:
  party_members:
    type: npcs
    pos: ofs_party_members
    repeat: expr
    repeat-expr: num_party_members
  #party inventory is not reversed yet
  #party_inventory:
    #type: inventory
    #pos: ofs_party_inventory
    #repeat: expr
    #repeat-expr: num_party_inventory
  non_party_members:
    type: npcs
    pos: ofs_non_party_members
    repeat: expr
    repeat-expr: num_non_party_members
  global_variables:
    type: global_var
    pos: ofs_global_variables
    repeat: expr
    repeat-expr: num_global_variables
  familiar_extra:
    type: strz
    encoding: ASCII
    size: 8
    if: ofs_familiar_extra != 0xFFFFFFFF
    pos: ofs_familiar_extra
  journal_entries:
    type: journal_entry
    pos: ofs_journal_entries
    repeat: expr
    repeat-expr: num_journal_entries
  familiar_info:
    type: familiar_info
    if: ofs_familiar_info != 0xFFFFFFFF
    pos: ofs_familiar_info
  stored_locations:
    type: stored_locations_info
    pos: ofs_stored_locations
    repeat: expr
    repeat-expr: num_stored_locations
  pocket_plane_locations:
    type: pocket_plane_info
    pos: ofs_pocket_plane_locations
    repeat: expr
    repeat-expr: num_pocket_plane_locations

types:
  weather_flags:
    seq:
    - id: rain
      type: b1
    - id: snow
      type: b1
    - id: light_rain
      type: b1
    - id: medium_rain
      type: b1
    - id: light_wind
      type: b1
    - id: medium_wind
      type: b1
    - id: rare_lightning
      type: b1
    - id: lightning
      type: b1
    - id: storm_increasing
      type: b1

  gui_flags:
    seq:
    - id: party_ai_enabled
      type: b1
    - id: text_window_size
      type: b2
      enum: text_window_size
    - type: b1
    - id: hide_gui
      type: b1
    - id: hide_options
      type: b1
    - id: hide_portraits
      type: b1
    - id: show_automap_notes
      type: b1

  npcs:
    seq:
    - id: character_selection
      type: u2
      enum: character_selection
    - id: party_order
      type: u2
    - id: ofs_cre_data
      type: u4
    - id: size_cre_data
      type: u4
    - id: character_name
      type: strz
      encoding: UTF-8
      size: 8
    - id: character_orientation
      type: u4
      # [TODO] @GooRoo: use orientation enum from ARE parser
    - id: character_cur_area
      type: strz
      encoding: ASCII
      size: 8
    - id: character_x
      type: u2
    - id: character_y
      type: u2
    - id: view_rect_x
      type: u2
    - id: view_rect_y
      type: u2
    - id: modal_action
      type: u2
    - id: happiness
      type: u2
    - id: count_interacted_npc
      type: u4
      repeat: expr
      repeat-expr: 24
    - id: quick_weapon_slot
      doc: See `slots.ids`
      type: u2
      repeat: expr
      repeat-expr: 4
    - id: quick_weapon_ability
      type: u2
      repeat: expr
      repeat-expr: 4
    - id: quick_spell_spl
      type: strz
      encoding: ASCII
      size: 8
      repeat: expr
      repeat-expr: 3
    - id: quick_item_slot
      doc: See `slots.ids`
      type: u2
      repeat: expr
      repeat-expr: 3
    - id: quick_item_ability
      type: u2
      repeat: expr
      repeat-expr: 3
    - id: name
      type: strz
      encoding: UTF-8
      size: 32
    - id: talkcount
      type: u4
    - id: character_stats
      type: char_stats
    - id: voice_set
      size: 8

  char_stats:
    seq:
    - id: mpv_name_ref
      type: u4
    - id: mpv_xp_reward
      type: u4
    - id: time_in_party
      type: u4
    - id: time_joined
      type: u4
    - id: is_party_member
      type: u1
    - size: 2
    - id: first_letter_cre
      type: u1
    - id: kills_xp_chapter
      type: u4
    - id: kills_count_chapter
      type: u4
    - id: kills_xp
      type: u4
    - id: kills_count
      type: u4
    - id: favorite_spell_spl
      type: strz
      encoding: ASCII
      size: 8
      repeat: expr
      repeat-expr: 4
    - id: favorite_spell_count
      type: u2
      repeat: expr
      repeat-expr: 4
    - id: favorite_weapon_itm
      type: strz
      encoding: ASCII
      size: 8
      repeat: expr
      repeat-expr: 4
    - id: favorite_weapon_time
      type: u2
      repeat: expr
      repeat-expr: 4

  # [TODO] @GooRoo: use variable type from ARE parser
  global_var:
    seq:
    - id: var_name
      type: strz
      encoding: ASCII
      size: 32
    - id: var_type
      type: u2
    - id: ref_val
      type: u2
    - id: dword_val
      type: u4
    - id: int_val
      type: s4
    - id: double_val
      type: u8
    - id: script_name
      type: strz
      encoding: ASCII
      size: 32

  journal_entry:
    seq:
    - id: journal_text_ref
      type: u4
    - id: time_seconds
      type: u4
    - id: current_chapter_num
      type: u1
    - id: read_by_char_x
      type: u1
    - id: journal_section
      type: u1 #bitfield with weird state when no bits set
    - id: location
      type: u1
      enum: location

  familiar_info:
    seq:
    - id: lawful_good_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: lawful_neutral_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: lawful_evil_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: neutral_good_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: neutral_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: neutral_evil_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: chaotic_good_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: chaotic_neutral_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: chaotic_evil_familiar_cre
      type: strz
      encoding: ASCII
      size: 8
    - id: ofs_familiar_res #not reversed
      type: u4
    - id: num_familiar_lg_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_ln_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_cg_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_ng_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_tn_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_ne_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_le_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_cn_level
      type: u4
      repeat: expr
      repeat-expr: 9
    - id: num_familiar_ce_level
      type: u4
      repeat: expr
      repeat-expr: 9

  stored_locations_info:
    seq:
    - id: area_are
      type: strz
      encoding: ASCII
      size: 8
    - id: x_coord
      type: u2
    - id: y_coord
      type: u2

  pocket_plane_info:
    seq:
    - id: area_are
      type: strz
      encoding: ASCII
      size: 8
    - id: x_coord
      type: u2
    - id: y_coord
      type: u2

enums:
  party_member:
    0: player_1
    1: player_2
    2: player_3
    3: player_4
    4: player_5
    5: player_6
    0xffff: not_in_party

  loading_progress:
    0: restrict_xp_bg1
    1: restrict_xp_totsc
    2: restrict_xp_soa
    3: processing_xnewarea_2da
    4: complete_xnewarea_2da
    5: tob_active

  character_selection:
    0x0: not_selected
    0x1: selected
    0x8000: dead

  text_window_size:
    0: small
    1: medium
    2: unused
    3: large

  location:
    0x1F: external_tot_toh
    0xFF: internal_tlk
