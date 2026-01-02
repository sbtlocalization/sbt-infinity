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
  - id: north_area_are
    type: strz
    size: 8
    encoding: ASCII
  - id: north_area_flags
    size: 4
    type: neighboring_area_flags
  - id: east_area_are
    type: strz
    size: 8
    encoding: ASCII
  - id: east_area_flags
    size: 4
    type: neighboring_area_flags
  - id: south_area_are
    type: strz
    size: 8
    encoding: ASCII
  - id: south_area_flags
    size: 4
    type: neighboring_area_flags
  - id: west_area_are
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
    type: bool2
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
  - id: area_script_bcs
    type: strz
    size: 8
    encoding: ASCII
  - id: num_explored_bitmask
    type: u4
  - id: ofs_explored_bitmask
    type: u4
  - id: num_doors
    type: u4
  - id: ofs_doors
    type: u4
  - id: num_animations
    type: u4
  - id: ofs_animations
    type: u4
  - id: num_tiled_objects
    type: u4
  - id: ofs_tiled_objects
    type: u4
  - id: ofs_songs
    type: u4
  - id: ofs_rest_encounters
    type: u4
  - id: other_offsets
    type: other_offsets
  - id: num_bg_projectile_traps
    type: u4
  - id: bg_rest_movie_day
    type: strz
    size: 8
    encoding: ASCII
  - id: bg_rest_movie_night
    type: strz
    size: 8
    encoding: ASCII
  - size: 56
instances:
  actors:
    pos: ofs_actors
    size: 0x110
    type: actor
    repeat: expr
    repeat-expr: num_actors
  regions:
    pos: ofs_regions
    size: 0xc4
    type: region
    repeat: expr
    repeat-expr: num_regions
  spawn_points:
    pos: ofs_spawn_points
    size: 0xc8
    type: spawn_point
    repeat: expr
    repeat-expr: num_spawn_points
  entrances:
    pos: ofs_entrances
    size: 0x68
    type: entrance
    repeat: expr
    repeat-expr: num_entrances
  containers:
    pos: ofs_containers
    size: 0xc0
    type: container
    repeat: expr
    repeat-expr: num_containers
  vertices:
    pos: ofs_vertices
    size: 4
    type: point
    repeat: expr
    repeat-expr: num_vertices
  items:
    pos: ofs_items
    size: 20
    type: item
    repeat: expr
    repeat-expr: num_items
  ambients:
    pos: ofs_ambients
    size: 0xd4
    type: ambient
    repeat: expr
    repeat-expr: num_ambients
  variables:
    pos: ofs_variables
    size: 0x54
    type: variable
    repeat: expr
    repeat-expr: num_variables
  explored_bitmask:
    pos: ofs_explored_bitmask
    repeat: expr
    repeat-expr: num_explored_bitmask
    type: b1
  doors:
    pos: ofs_doors
    size: 0xc8
    type: door
    repeat: expr
    repeat-expr: num_doors
  animations:
    pos: ofs_animations
    size: 0x4c
    type: animation
    repeat: expr
    repeat-expr: num_animations
  bg_automap_notes:
    pos: other_offsets.ofs_bg_automap_notes
    size: 0x34
    type: bg_automap_note
    repeat: expr
    repeat-expr: other_offsets.num_bg_automap_notes
  pst_automap_notes:
    pos: other_offsets.ofs_pst_automap_notes
    size: 0x214
    type: pst_automap_note
    repeat: expr
    repeat-expr: other_offsets.num_pst_automap_notes
  tiled_objects:
    pos: ofs_tiled_objects
    size: 0x68
    type: tiled_object
    repeat: expr
    repeat-expr: num_tiled_objects
  bg_projectile_traps:
    pos: other_offsets.ofs_bg_projectile_traps
    size: 0x1c
    type: bg_projectile_trap
    repeat: expr
    repeat-expr: num_bg_projectile_traps
  songs:
    pos: ofs_songs
    size: 0x90
    type: songs
  rest_encounters:
    pos: ofs_rest_encounters
    size: 0xe4
    type: rest_encounters
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
      - id: bit
        type: b1
        repeat: expr
        repeat-expr: 11
    instances:
      bg_outdoor:
        value: bit[0]
      bg_day_night:
        value: bit[1]
      bg_weather:
        value: bit[2]
      bg_city:
        value: bit[3]
      bg_forest:
        value: bit[4]
      bg_dungeon:
        value: bit[5]
      bg_extended_night:
        value: bit[6]
      bg_can_rest_indoors:
        value: bit[7]
      pst_hive:
        value: bit[0]
      pst_clerks_ward:
        value: bit[2]
      pst_lower_ward:
        value: bit[3]
      pst_ravels_maze:
        value: bit[4]
      pst_baator:
        value: bit[5]
      pst_rubikon:
        value: bit[6]
      pst_fortress_of_regrets:
        value: bit[7]
      pst_curst:
        value: bit[8]
      pst_carceri:
        value: bit[9]
      pst_outdoors:
        value: bit[10]
  other_offsets:
    seq:
      - id: first
        type: u4
      - id: second
        type: u4
      - id: third
        type: u4
    instances:
      ofs_bg_automap_notes:
        value: first
      num_bg_automap_notes:
        value: second
      ofs_bg_projectile_traps:
        value: third
      ofs_pst_automap_notes:
        value: second
      num_pst_automap_notes:
        value: third
  actor:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: current_x
        type: u2
      - id: current_y
        type: u2
      - id: destination_x
        type: u2
      - id: destination_y
        type: u2
      - id: flags
        type: flags
        size: 4
      - id: spawnable
        type: bool2
      - id: first_cre_letter
        type: u1
      - size: 1
      - id: actor_animation
        type: u4
      - id: actor_orientation
        type: u2
        enum: orientation
      - size: 2
      - id: expiry_time
        type: u4
      - id: wander_distance
        type: u2
      - id: follow_distance
        type: u2
      - id: appearance_schedule
        type: schedule
      - id: num_times_talked_to
        type: u4
      - id: dialog_dlg
        type: strz
        size: 8
        encoding: ASCII
      - id: override_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: general_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: class_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: race_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: default_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: specific_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: cre
        type: strz
        size: 8
        encoding: ASCII
      - id: ofs_cre_structure
        type: u4
      - id: len_cre_structure
        type: u4
      - size: 128
    types:
      flags:
        seq:
          - id: cre_attached
            type: b1
          - id: has_seen_party
            type: b1
          - id: invulnerable
            type: b1
          - id: override_script_name
            type: b1
  region:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: region_type
        enum: region_type
        type: u2
      - id: bounding_box
        type: bounding_box
      - id: num_vertices
        type: u2
      - id: first_vertex_index
        type: u4
      - id: trigger_value
        type: u4
      - id: cursor_index
        doc: cursors.bam
        type: u4
      - id: destination_area_are
        type: strz
        size: 8
        encoding: ASCII
      - id: entrance_name
        type: strz
        size: 32
        encoding: UTF-8
      - id: flags
        type: flags
        size: 4
      - id: info_ref
        type: u4
      - id: trap_detection_difficulty
        type: u2
      - id: trap_removal_difficulty
        type: u2
      - id: is_trapped
        type: bool2
      - id: is_trap_detected
        type: bool2
      - id: trap_launch_location
        type: point
      - id: key_itm
        type: strz
        size: 8
        encoding: ASCII
      - id: script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: alternative_point
        type: point
      - size: 4
      - size: 32
      - id: pst_sound
        type: strz
        size: 8
        encoding: ASCII
      - id: pst_talk_location
        type: point
      - id: pst_speaker_name_ref
        type: u4
      - id: pst_dialog_dlg
        type: strz
        size: 8
        encoding: ASCII
    instances:
      vertices:
        pos: _root.ofs_vertices + first_vertex_index * 4
        io: _root._io
        type: point
        size: 4
        repeat: expr
        repeat-expr: num_vertices
    types:
      flags:
        seq:
          - id: key_required
            type: b1
          - id: reset_trap
            type: b1
          - id: party_required
            type: b1
          - id: detectable
            type: b1
          - id: npc_activates
            type: b1
          - id: active_in_tutorial_area_only
            type: b1
          - id: anyone_activates
            type: b1
          - id: silent
            type: b1
          - id: deactivated
            type: b1
          - id: party_only
            type: b1
          - id: use_alternative_point
            type: b1
          - id: connected_to_door
            type: b1
    enums:
      region_type:
        0: proximity_trigger
        1: info_point
        2: travel_region
  spawn_point:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: coord
        type: point
      - id: creature_cre
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 10
      - id: num_creatures
        type: u2
      - id: encounter_difficulty
        type: u2
      - id: spawn_rate
        type: u2
      - id: spawn_method
        type: spawn_method
        size: 2
      - id: expiry_time
        type: u4
      - id: wander_distance
        type: u2
      - id: follow_distance
        type: u2
      - id: maximum_num_creatures
        type: u2
      - id: enabled
        type: bool2
      - id: appearance_schedule
        type: schedule
      - id: probability_day
        type: u2
      - id: probability_night
        type: u2
      - id: spawn_frequency
        type: u4
      - id: countdown
        type: u4
      - id: spawn_weight
        type: u1
        repeat: expr
        repeat-expr: 10
      - size: 38
    types:
      spawn_method:
        seq:
          - id: spawn_until_paused
            type: b1
          - id: single_shot
            type: b1
          - id: spawn_paused
            type: b1
  entrance:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: coord
        type: point
      - id: orientation
        type: u2
        enum: orientation
      - size: 66
  container:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: coord
        type: point
      - id: container_type
        type: u2
        enum: container_type
      - id: lock_difficulty
        type: u2
      - id: flags
        type: flags
        size: 4
      - id: trap_detection_difficulty
        type: u2
      - id: trap_removal_difficulty
        type: u2
      - id: is_trapped
        type: bool2
      - id: is_trap_detected
        type: bool2
      - id: trap_launch_coord
        type: point
      - id: bounding_box
        type: bounding_box
      - id: first_item_index
        type: u4
      - id: num_items
        type: u4
      - id: trap_script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: first_vertex_index
        type: u4
      - id: num_vertices
        type: u2
      - id: trigger_range
        type: u2
      - id: owner
        type: strz
        size: 32
        encoding: UTF-8
      - id: key_itm
        type: strz
        size: 8
        encoding: ASCII
      - id: break_difficulty
        type: u4
      - id: lockpick_ref
        type: u4
      - size: 56
    instances:
      vertices:
        pos: _root.ofs_vertices + first_vertex_index * 4
        io: _root._io
        type: point
        size: 4
        repeat: expr
        repeat-expr: num_vertices
      items:
        pos: _root.ofs_items + first_item_index * 20
        io: _root._io
        type: item
        size: 20
        repeat: expr
        repeat-expr: num_items
    types:
      flags:
        seq:
          - id: locked
            type: b1
          - id: disable_if_no_owner
            type: b1
          - id: magically_locked
            type: b1
          - id: trap_resets
            type: b1
          - id: remove_only
            type: b1
          - id: disabled
            type: b1
          - id: dont_clear
            type: b1
    enums:
      container_type:
        0x00: n_a
        0x01: bag
        0x02: chest
        0x03: drawer
        0x04: pile
        0x05: table
        0x06: shelf
        0x07: altar
        0x08: non_visible
        0x09: spellbook
        0x0a: body
        0x0b: barrel
        0x0c: crate
  item:
    seq:
      - id: item_itm
        type: strz
        size: 8
        encoding: ASCII
      - id: expiry_time
        type: u2
      - id: quantity_charges
        type: u2
        repeat: expr
        repeat-expr: 3
      - id: flags
        type: flags
        size: 4
    types:
      flags:
        seq:
          - id: identified
            type: b1
          - id: unstealable
            type: b1
          - id: stolen
            type: b1
          - id: undroppable
            type: b1
  ambient:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: ASCII
      - id: coord
        type: point
      - id: radius
        type: u2
      - id: height
        type: u2
      - id: pitch_variance
        type: u4
      - id: volume_variance
        type: u2
      - id: volume
        type: u2
      - id: sound_wav
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 10
      - id: num_sounds
        type: u2
      - size: 2
      - id: base_interval
        type: u4
      - id: interval_variation
        type: u4
      - id: play_schedule
        type: schedule
      - id: flags
        type: flags
        size: 4
      - size: 64
    types:
      flags:
        seq:
          - id: enabled
            type: b1
          - id: loop
            type: b1
          - id: ignore_radius
            type: b1
          - id: random_order
            type: b1
          - id: high_memory_ambient
            type: b1
  bounding_box:
    seq:
      - id: left
        type: u2
      - id: top
        type: u2
      - id: right
        type: u2
      - id: bottom
        type: u2
  point:
    seq:
      - id: x
        type: u2
      - id: y
        type: u2
  bool2:
    seq:
      - id: value
        type: b1
      - type: b15
  schedule:
    seq:
      - id: from_0030_till_0129
        type: b1
      - id: from_0130_till_0229
        type: b1
      - id: from_0230_till_0329
        type: b1
      - id: from_0330_till_0429
        type: b1
      - id: from_0430_till_0529
        type: b1
      - id: from_0530_till_0629
        type: b1
      - id: from_0630_till_0729
        type: b1
      - id: from_0730_till_0829
        type: b1
      - id: from_0830_till_0929
        type: b1
      - id: from_0930_till_1029
        type: b1
      - id: from_1030_till_1129
        type: b1
      - id: from_1130_till_1229
        type: b1
      - id: from_1230_till_1329
        type: b1
      - id: from_1330_till_1429
        type: b1
      - id: from_1430_till_1529
        type: b1
      - id: from_1530_till_1629
        type: b1
      - id: from_1630_till_1729
        type: b1
      - id: from_1730_till_1829
        type: b1
      - id: from_1830_till_1929
        type: b1
      - id: from_1930_till_2029
        type: b1
      - id: from_2030_till_2129
        type: b1
      - id: from_2130_till_2229
        type: b1
      - id: from_2230_till_2329
        type: b1
      - id: from_2330_till_0029
        type: b1
      - size: 1
  variable:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: ASCII
      - id: var_type
        type: u2
        enum: var_type
      - id: ref_value
        type: u2
      - id: dword_value
        type: u4
      - id: int_value
        type: s4
      - id: double_value
        type: f8
      - id: script_name_value
        type: strz
        size: 32
        encoding: UTF-8
    enums:
      var_type:
        0: integer
        1: float
        2: script_name
        3: res_ref
        4: str_ref
        5: dword
  door:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: door_id
        type: strz
        size: 8
        encoding: ASCII
      - id: flags
        type: flags
        size: 4
      - id: first_vertex_index_open_door
        type: u4
      - id: num_vertices_open_door
        type: u2
      - id: num_vertices_closed_door
        type: u2
      - id: first_vertex_index_closed_door
        type: u4
      - id: minimum_bb_open_door
        type: bounding_box
      - id: minimmum_bb_closed_door
        type: bounding_box
      - id: first_impeded_cell_index_open_door
        type: u4
      - id: num_impeded_cells_open_door
        type: u2
      - id: num_impeded_cells_closed_door
        type: u2
      - id: first_impeded_cell_index_closed_door
        type: u4
      - id: hit_points
        type: u2
      - id: armor_class
        type: u2
      - id: door_open_sound_wav
        type: strz
        size: 8
        encoding: ASCII
      - id: door_close_sound_wav
        type: strz
        size: 8
        encoding: ASCII
      - id: cursor_index
        type: u4
        doc: See `cursors.bam`
      - id: trap_detection_difficulty
        type: u2
      - id: trap_removal_difficulty
        type: u2
      - id: is_trapped
        type: bool2
      - id: trap_detected
        type: bool2
      - id: trap_launch_point
        type: point
      - id: key_itm
        type: strz
        size: 8
        encoding: ASCII
      - id: script_bcs
        type: strz
        size: 8
        encoding: ASCII
      - id: detection_difficulty
        type: u4
      - id: lock_difficulty
        type: u4
      - id: open_location
        type: point
      - id: close_location
        type: point
      - id: unlock_message_ref
        type: u4
      - id: travel_trigger_name
        type: strz
        size: 24
        encoding: ASCII
      - id: speaker_name_ref
        type: u4
      - id: dialog_dlg
        type: strz
        size: 8
        encoding: ASCII
      - size: 8
    types:
      flags:
        seq:
          - id: door_open
            type: b1
          - id: door_locked
            type: b1
          - id: reset_trap
            type: b1
          - id: trap_detectable
            type: b1
          - id: broken
            type: b1
          - id: cant_close
            type: b1
          - id: linked
            type: b1
          - id: door_hidden
            type: b1
          - id: door_found
            type: b1
          - id: can_be_looked_through
            type: b1
          - id: consumes_key
            type: b1
          - id: ignore_obstacles_when_closing
            type: b1
  animation:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: coordinate
        type: point
      - id: appearance_schedule
        type: schedule
      - id: animation_bam
        type: strz
        size: 8
        encoding: ASCII
      - id: bam_sequence
        type: u2
      - id: bam_frame
        type: u2
      - id: flags
        size: 4
        type: flags
      - id: height
        type: u2
      - id: transparency
        type: u2
      - id: start_range
        type: u2
      - id: loop_probability
        type: u1
      - id: start_delay
        type: u1
      - id: palette_bmp
        type: strz
        size: 8
        encoding: ASCII
      - id: movie_width
        type: u2
      - id: movie_height
        type: u2
    types:
      flags:
        seq:
          - id: enabled
            type: b1
          - id: black_is_transparent
            type: b1
          - id: not_light_source
            type: b1
          - id: partial
            type: b1
          - id: synchronized_draw
            type: b1
          - id: random_start_frame
            type: b1
          - id: not_covered_by_wall
            type: b1
          - id: disable_on_slow_machines
            type: b1
          - id: draw_as_background
            type: b1
          - id: play_all_frames
            type: b1
          - id: use_palette
            type: b1
          - id: mirror_y_axis
            type: b1
          - id: show_in_combat
            type: b1
          - id: use_wbm
            type: b1
          - id: draw_stenciled
            type: b1
          - id: use_pvrz
            type: b1
          - id: pst_cover_animations
            type: b1
        instances:
          pst_alt_blending_mode:
            value: draw_as_background
  bg_automap_note:
    seq:
      - id: coordinate
        type: point
      - id: note_ref
        type: u4
      - id: note_ref_is_internal
        doc: |
          Internal means it's in `dialog.tlk`,
          external means it's overridden with TOH/TOT.
        type: bool2
      - id: marker_color
        type: u2
        enum: marker_color
      - id: control_id
        type: u4
      - size: 36
    enums:
      marker_color:
        0: gray
        1: violet
        2: green
        3: orange
        4: red
        5: blue
        6: dark_blue
        7: light_gray
  pst_automap_note:
    seq:
      - id: x
        type: u4
      - id: y
        type: u4
      - id: text
        type: strz
        size: 500
        encoding: ASCII
      - id: note_color
        type: u4
        enum: note_color
      - size: 20
    enums:
      note_color:
        0: blue_user_note
        1: red_game_note
  tiled_object:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: tile_id
        type: strz
        size: 8
        encoding: ASCII
      - id: flags
        size: 4
        type: flags
      - id: ofs_open_search_squares
        type: u4
      - id: num_open_search_squares
        type: u2
      - id: num_closed_search_squares
        type: u2
      - id: ofs_closed_search_squares
        type: u4
      - size: 48
    types:
      flags:
        seq:
          - id: in_secondary_state
            type: b1
          - id: can_be_looked_through
            type: b1
  bg_projectile_trap:
    seq:
      - id: projectile_pro
        type: strz
        size: 8
        encoding: ASCII
      - id: effect_block_offset
        type: u4
      - id: effect_block_size
        type: u2
      - id: missile_id
        type: u2
        doc: See `missile.ids`
      - id: ticks_until_next_trigger_check
        type: u2
      - id: triggers_remaining
        type: u2
      - id: x
        type: u2
      - id: y
        type: u2
      - id: z
        type: u2
      - id: enemy_ally_targeting
        type: u1
      - id: party_member_index
        type: u1
  songs:
    doc: See `songlist.2ds`/`musiclis.ids`/`songs.ids`.
    seq:
      - id: day_song
        type: u4
      - id: night_song
        type: u4
      - id: win_song
        type: u4
      - id: battle_song
        type: u4
      - id: lose_song
        type: u4
      - id: alt_music
        type: u4
        repeat: expr
        repeat-expr: 5
      - id: main_day_ambient_wav
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 2
      - id: main_day_ambient_volume
        type: u4
      - id: main_night_ambient_wav
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 2
      - id: main_night_ambient_volume
        type: u4
      - id: reverb
        type: u4
        doc: See `REVERB.IDS`/`REVERB.2DA` if exist.
      - size: 60
  rest_encounters:
    seq:
      - id: name
        type: strz
        size: 32
        encoding: UTF-8
      - id: creature_text_ref
        type: u4
        repeat: expr
        repeat-expr: 10
      - id: creature_to_spawn
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 10
        doc: See `spawngrp.2da`.
      - id: num_creatures
        type: u2
      - id: difficulty
        type: u2
      - id: expiry_time
        type: u4
      - id: creature_wander_distance
        type: u2
      - id: creature_follow_distance
        type: u2
      - id: maximum_creatures
        type: u2
      - id: active
        type: bool2
      - id: hourly_probability_day
        type: u2
      - id: hourly_probability_night
        type: u2
      - size: 56
enums:
  orientation:
    0: south
    1: south_south_west
    2: south_west
    3: west_south_west
    4: west
    5: west_north_west
    6: north_west
    7: north_north_west
    8: north
    9: north_north_east
    10: north_east
    11: east_north_east
    12: east
    13: east_south_east
    14: south_east
    15: south_south_east
