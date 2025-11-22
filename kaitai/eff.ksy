# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: eff
  title: EFF v2.0
  file-extension: eff
  ks-version: "0.11"
  endian: le
  bit-endian: le
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/eff_v2.htm
seq:
  - id: magic
    contents: "EFF "
  - id: version
    contents: "V2.0"
  - id: body
    type: body_v2(false)
types:
  body_v2:
    params:
      - id: embedded
        type: bool
    seq:
      - id: magic
        contents: "EFF "
        if: not embedded
      - id: version
        contents: "V2.0"
        if: not embedded
      - id: magic2
        contents: [0, 0, 0, 0]
        if: embedded
      - id: version2
        contents: [0, 0, 0, 0]
        if: embedded
      - id: opcode
        type: u4
      - id: target_type
        type: u4
        enum: target_type
      - id: power
        type: u4
      - id: parameter1
        type: u4
      - id: parameter2
        type: u4
      - id: timing_mode
        type: u2
        enum: timing_mode
      - type: u2
      - id: duration
        type: u4
      - id: probability1
        type: u2
      - id: probability2
        type: u2
      - id: res_name
        type: strz
        size: 8
        encoding: ASCII
      - id: dice_thrown
        type: u4
      - id: dice_sides
        type: u4
      - id: saving_throw_type
        size: 4
        type: saving_throw_type
      - id: save_bonus
        type: u4
      - id: special
        type: u4
      - id: primary_type
        type: u4
        doc: See `MSCHOOL.2DA`.
      - id: reserved
        type: u4
      - id: parent_resource_lowest_affected_level
        type: u4
      - id: parent_resource_highest_affected_level
        type: u4
      - id: dispel_resistance
        size: 4
        type: dispel_resistance
      - id: parameter3
        type: u4
      - id: parameter4
        type: u4
      - id: parameter5
        type: u4
      - id: time_applied
        type: u4
      - id: res_name2
        type: strz
        size: 8
        encoding: ASCII
        doc: |
          `VCC` in many effects.
      - id: res_name3
        type: strz
        size: 8
        encoding: ASCII
      - id: caster_coordinate
        type: coord
      - id: target_coordinate
        type: coord
      - id: parent_resource_type
        type: u4
        enum: parent_resource_type
      - id: parent_resource
        type: strz
        size: 8
        encoding: ASCII
      - id: parent_resource_flags
        size: 4
        type: parent_resource_flags
      - id: projectile
        type: u4
      - id: parent_resource_slot
        type: u4
      - id: variable_name
        type: strz
        size: 32
        encoding: ASCII
      - id: caster_level
        type: u4
      - id: first_apply
        type: u4
      - id: secondary_type
        type: u4
        doc: See `MSECTYPE.2DA`.
      - size: 4 * 15
    types:
      coord:
        seq:
          - id: x
            type: u4
          - id: y
            type: u4
      saving_throw_type:
        seq:
          - id: spells
            type: b1
          - id: breath
            type: b1
          - id: paralyze_poison_death
            type: b1
          - id: wands
            type: b1
          - id: petrify_polymorph
            type: b1
          - id: ee_spells
            type: b1
          - id: ee_breath
            type: b1
          - id: ee_paralyze_poison_death
            type: b1
          - id: ee_wands
            type: b1
          - id: ee_petrify_polymorph
            type: b1
          - id: ignore_primary_target
            type: b1
          - id: ignore_secondary_target
            type: b1
          - type: b13
          - id: bypass_mirror_image
            type: b1
          - id: ignore_difficulty
            type: b1
          - id: reserved
            type: b1
      dispel_resistance:
        seq:
          - id: dispel
            type: b1
          - id: bypass_resistance
            type: b1
          - id: bypass_ee
            type: b1
          - id: self_targeted
            type: b1
          - type: b27
          - id: effect_applied_by_ite
            type: b1
      parent_resource_flags:
        seq:
          - type: b10
          - id: hostile
            type: b1
          - id: no_los_required
            type: b1
          - id: allow_spotting
            type: b1
          - id: outdoors_only
            type: b1
          - id: non_magical_ability
            type: b1
          - id: ignore_wild_surge
            type: b1
          - id: non_combat_ability
            type: b1
    enums:
      target_type:
        0: none
        1: self
        2: projectile_target
        3: party
        4: everyone
        5: everyone_but_party
        6: caster_group
        7: target_group
        8: everyone_but_self
        9: original_caster
      timing_mode:
        0: instant_limited
        1: instant_permanent_until_death
        2: instant_while_equipped
        3: delay_limited
        4: delay_permanent
        5: delay_while_equipped
        6: limited_after_duration
        7: permanent_after_duration
        8: equipped_after_duration
        9: instant_permanent
        10: instant_limited_ticks
        4096: absolute_duration
      parent_resource_type:
        0: none
        1: spell
        2: item
