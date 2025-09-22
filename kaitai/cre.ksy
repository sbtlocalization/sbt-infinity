# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: cre
  title: CRE
  file-extension: cre
  ks-version: "0.10"
  endian: le
  bit-endian: le
  imports:
    - eff
doc: |
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/cre_v1.htm
seq:
  - id: magic
    contents: "CRE "
  - id: version
    type: str
    encoding: ASCII
    size: 4
  - id: long_name
    type: u4
  - id: short_name
    type: u4
  - id: body
    type:
      switch-on: version
      cases:
        '"V1.0"': body_v1
enums:
  spell_type:
    0: priest
    1: wizard
    2: innate
types:
  body_v1:
    seq:
      - id: header
        type: header
    instances:
      known_spells:
        pos: header.ofs_known_spells
        type: known_spell
        repeat: expr
        repeat-expr: header.num_known_spells
      spell_memorization_infos:
        pos: header.ofs_spell_memorization_info
        type: spell_memorization_info
        repeat: expr
        repeat-expr: header.num_spell_memorization_info
        doc: |
          This section details how many spells the creature can memorize, and how many it has memorized. It consists of an array of entries.
      memorized_spells:
        pos: header.ofs_memorized_spells
        type: memorized_spell
        repeat: expr
        repeat-expr: header.num_memorized_spells
        doc: |
          This section details which spells the character has memorized. It consists of an array of entries formatted as follows.
      effects:
        pos: header.ofs_effects
        type:
          switch-on: header.eff_version
          cases:
            header::eff_version::version2: eff::body_v2(true)
        repeat: expr
        repeat-expr: header.num_effects
      items:
        pos: header.ofs_items
        type: item
        repeat: expr
        repeat-expr: header.num_items
      item_slots:
        pos: header.ofs_item_slots
        type: item_slot
    types:
      header:
        seq:
          - id: flags
            type: flags
          - id: xp_for_killing
            type: u4
          - id: xp_power_level
            type: u4
          - id: gold_carried
            type: u4
          - id: status
            type: status_flags
          - id: current_hp
            type: u2
          - id: maximum_hp
            type: u2
          - id: animation_id
            type: u4
          - id: metal_color_index
            type: u1
          - id: minor_color_index
            type: u1
          - id: major_color_index
            type: u1
          - id: skin_color_index
            type: u1
          - id: leather_color_index
            type: u1
          - id: armor_color_index
            type: u1
          - id: hair_color_index
            type: u1
          - id: eff_version
            type: u1
            enum: eff_version
          - id: small_portrait
            type: strz
            size: 8
            encoding: ASCII
          - id: large_portrait
            type: strz
            size: 8
            encoding: ASCII
            doc: |
              PSTEE: BAM
              Other games: BMP
          - id: reputation
            type: s1
          - id: hide_in_shadows
            type: u1
          - id: natural_ac
            type: s2
          - id: effective_ac
            type: s2
          - id: crushing_ac_modifier
            type: s2
          - id: missile_ac_modifier
            type: s2
          - id: piercing_ac_modifier
            type: s2
          - id: slashing_ac_modifier
            type: s2
          - id: thac0
            type: u1
            valid:
              min: 1
              max: 25
          - id: num_attacks
            type: u1
            valid:
              min: 0
              max: 10
            doc: |
              Number of attacks per round
          - id: save_vs_death
            type: u1
            valid:
              min: 0
              max: 20
          - id: save_vs_wands
            type: u1
            valid:
              min: 0
              max: 20
          - id: save_vs_polymorph
            type: u1
            valid:
              min: 0
              max: 20
          - id: save_vs_breath_attacks
            type: u1
            valid:
              min: 0
              max: 20
          - id: save_vs_spells
            type: u1
            valid:
              min: 0
              max: 20
          - id: resist_fire
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_cold
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_electricity
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_acid
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_magic
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_magic_fire
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_magic_cold
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_slashing
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_crushing
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_piercing
            type: u1
            valid:
              min: 0
              max: 100
          - id: resist_missile
            type: u1
            valid:
              min: 0
              max: 100
          - id: detect_illusion
            type: u1
            doc: Minimum value – 0
          - id: set_traps
            type: u1
          - id: lore
            type: u1
            valid:
              min: 0
              max: 100
            doc: |
              Lore is calculated as ((level * rate) + int_bonus + wis_bonus). Intelligence and wisdom
              bonuses are from LOREBON.2DA and the rate is the lookup value in LORE.2DA, based on class.
              For multiclass characters, (level * rate) is calculated for both classes separately and
              the higher of the two values is used - they are not cumulative.
          - id: lockpicking
            type: u1
            doc: Minimum value – 0
          - id: move_silently
            type: u1
            doc: Minimum value – 0
          - id: find_disarm_traps
            type: u1
            doc: Minimum value – 0
          - id: pick_pockets
            type: u1
            doc: Minimum value – 0
          - id: fatigue
            type: u1
            valid:
              min: 0
              max: 100
          - id: intoxication
            type: u1
            valid:
              min: 0
              max: 100
          - id: luck
            type: u1
          - id: large_swords_proficiency
            type: u1
          - id: small_swords_proficiency
            type: u1
          - id: bows_proficiency
            type: u1
          - id: spears_proficiency
            type: u1
          - id: blunt_proficiency
            type: u1
          - id: spiked_proficiency
            type: u1
          - id: axe_proficiency
            type: u1
          - id: missile_proficiency
            type: u1
          - id: reserved_proficiency1
            type: u1
          - id: reserved_proficiency2
            type: u1
          - id: reserved_proficiency3
            type: u1
          - id: reserved_proficiency4
            type: u1
          - id: reserved_proficiency5
            type: u1
          - id: unspent_proficiencies
            type: u1
          - id: num_available_inventory_slots
            type: u1
          - id: nightmare_mode_modifiers
            type: u1
          - id: translucency
            type: u1
          - id: reputation_gain_loss_when_killed_or_murder_variable_increment
            type: u1
          - id: reputation_gain_loss_when_joining_party
            type: u1
          - id: reputation_gain_loss_when_leaving_party
            type: u1
          - id: turn_undead_level
            type: u1
          - id: tracking_skill
            type: u1
          - id: tracking_target_or_pstee_flags
            size: 32
          - id: str_refs
            type: u4
            repeat: expr
            repeat-expr: 100
          - id: level_first_class
            type: u1
          - id: level_second_class
            type: u1
          - id: level_third_class
            type: u1
          - id: sex
            type: u1
            doc: |
              Sex (`GENDER.IDS`) - checkable via the SEX stat.

              EE only: determines casting sound prefix.
              Known values include:
              - 1 (Male), default – `CHA_M*.WAV`
              - 2 (Female) – `CHA_F*.WAV`
              - 3 (Other), 4 (Niether) – `CHA_S*.WAV`
          - id: strength
            type: u1
            valid:
              min: 1
              max: 25
          - id: strength_bonus
            type: u1
            valid:
              min: 0
              max: 100
            doc: 0..100%
          - id: intelligence
            type: u1
            valid:
              min: 1
              max: 25
          - id: wisdom
            type: u1
            valid:
              min: 1
              max: 25
          - id: dexterity
            type: u1
            valid:
              min: 1
              max: 25
          - id: constitution
            type: u1
            valid:
              min: 1
              max: 25
          - id: charisma
            type: u1
            valid:
              min: 1
              max: 25
          - id: morale
            type: u1
            valid:
              min: 0
              max: 20
            doc: default – 10
          - id: morale_break
            type: u1
          - id: racial_enemy
            type: u1
            doc: See `RACE.IDS`
          - id: morale_recovery_time
            type: u2
          - type: u4
          - id: override_script
            type: strz
            size: 8
            encoding: ASCII
          - id: class_script
            type: strz
            size: 8
            encoding: ASCII
          - id: race_script
            type: strz
            size: 8
            encoding: ASCII
          - id: general_script
            type: strz
            size: 8
            encoding: ASCII
          - id: default_script
            type: strz
            size: 8
            encoding: ASCII
          - id: enemy_ally
            type: u1
            doc: See `EA.IDS`
          - id: general
            type: u1
            doc: See `GENERAL.IDS`
          - id: race
            type: u1
            doc: See `RACE.IDS`
          - id: class
            type: u1
            doc: See `CLASS.IDS`
          - id: specific
            type: u1
            doc: See `SPECIFIC.IDS`
          - id: gender
            type: u1
            doc: See `GENDER.IDS`
          - id: object_ids_refs
            type: u1
            repeat: expr
            repeat-expr: 5
            doc: See `OBJECT.IDS`
          - id: alignment
            type: u1
            doc: See `ALIGNMEN.IDS`
          - id: global_actor_enum_value
            type: u2
          - id: local_actor_enum_value
            type: u2
          - id: death_variable
            type: strz
            size: 32
            encoding: ASCII
          - id: ofs_known_spells
            type: u4
          - id: num_known_spells
            type: u4
          - id: ofs_spell_memorization_info
            type: u4
          - id: num_spell_memorization_info
            type: u4
          - id: ofs_memorized_spells
            type: u4
          - id: num_memorized_spells
            type: u4
          - id: ofs_item_slots
            type: u4
          - id: ofs_items
            type: u4
          - id: num_items
            type: u4
          - id: ofs_effects
            type: u4
          - id: num_effects
            type: u4
          - id: dialog
            type: strz
            size: 8
            encoding: ASCII
        enums:
          eff_version:
            0: version1
            1: version2
        types:
          flags:
            seq:
              - id: show_longname_in_tooltip
                type: b1
              - id: no_corpse
                type: b1
              - id: keep_corpse
                type: b1
              - id: original_class_fighter
                type: b1
              - id: original_class_mage
                type: b1
              - id: original_class_cleric
                type: b1
              - id: original_class_thief
                type: b1
              - id: original_class_druid
                type: b1
              - id: original_class_ranger
                type: b1
              - id: fallen_paladin
                type: b1
              - id: fallen_ranger
                type: b1
              - id: exportable
                type: b1
              - id: hide_injury_status
                type: b1
              - id: affected_by_alternative_damage
                type: b1
              - id: moving_between_areas
                type: b1
              - id: been_in_party
                type: b1
              - id: restore_item_in_hand
                type: b1
              - id: reset_restoring_item_in_hand
                type: b1
              - id: reserved1
                type: b1
              - id: reserved2
                type: b1
              - id: prevent_exploding_death
                type: b1
              - id: reserved3
                type: b1
              - id: ignore_nightmare_modifiers
                type: b1
              - id: no_tooltip
                type: b1
              - id: allegiance_tracking
                type: b1
              - id: general_tracking
                type: b1
              - id: race_tracking
                type: b1
              - id: class_tracking
                type: b1
              - id: specific_tracking
                type: b1
              - id: gender_tracking
                type: b1
              - id: alignment_tracking
                type: b1
              - id: uninterruptable
                type: b1
          status_flags:
            seq:
              - id: sleeping
                type: b1
              - id: berserk
                type: b1
              - id: panic
                type: b1
              - id: stunned
                type: b1
              - id: invisible
                type: b1
              - id: helpless
                type: b1
              - id: frozen_death
                type: b1
              - id: stone_death
                type: b1
              - id: exploding_death
                type: b1
              - id: flame_death
                type: b1
              - id: acid_death
                type: b1
              - id: dead
                type: b1
              - id: silenced
                type: b1
              - id: charmed
                type: b1
              - id: poisoned
                type: b1
              - id: hasted
                type: b1
              - id: slowed
                type: b1
              - id: infravision
                type: b1
              - id: blind
                type: b1
              - id: diseased_or_deactivated
                type: b1
              - id: feeble_minded
                type: b1
              - id: non_detection
                type: b1
              - id: improved_visibility
                type: b1
              - id: bless
                type: b1
              - id: chant
                type: b1
              - id: draw_upon_holy_might
                type: b1
              - id: luck
                type: b1
              - id: aid
                type: b1
              - id: chantbad
                type: b1
              - id: blur
                type: b1
              - id: mirror_image
                type: b1
              - id: confused
                type: b1
      known_spell:
        seq:
          - id: res_name
            type: strz
            size: 8
            encoding: ASCII
            doc: Resource name of the SPL file
          - id: spell_level
            type: u2
            doc: Spell level minus 1
          - id: spell_type
            type: u2
            enum: spell_type
      spell_memorization_info:
        seq:
          - id: spell_level
            type: u2
          - id: num_memorizable_spells
            type: u2
          - id: num_memorizable_spells_after_effects
            type: u2
          - id: spell_type
            type: u2
            enum: spell_type
          - id: first_memorized_spell
            type: u4
          - id: num_memorized_spell
            type: u4
      memorized_spell:
        seq:
          - id: res_name
            type: strz
            size: 8
            encoding: ASCII
            doc: Resource name of the SPL file.
          - id: flags
            size: 4
            type: memorization_flags
        types:
          memorization_flags:
            seq:
              - id: memorized
                type: b1
              - id: disabled
                type: b1
      item:
        seq:
          - id: res_name
            type: strz
            size: 8
            encoding: ASCII
            doc: Resource name of the ITM file.
          - id: duration
            type: u2
          - id: quantity_charges1
            type: u2
          - id: quantity_charges2
            type: u2
          - id: quantity_charges3
            type: u2
          - id: flags
            size: 4
            type: item_flags
        types:
          item_flags:
            seq:
              - id: identified
                type: b1
              - id: unstealable
                type: b1
              - id: stolen
                type: b1
              - id: undroppable
                type: b1
      item_slot: {}
  spell_v22:
    seq:
      - id: spell_index
        type: u4
        doc: |
          Index into the relevant 2da file: listspll, listdomm, listinnt, listshap, listsong.
      - id: amount_memorized
        type: u4
      - id: amount_remaining
        type: u4
      - type: u4
