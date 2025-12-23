# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: @definitelythehuman
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: itm
  file-extension: itm
  endian: le
  bit-endian: le
  imports:
    - eff
    - sto

doc: |
  This file format describes an "item". Items include weapons, armor, books, scrolls, rings and more.
  Items can have attached abilties, occuring either when a target creature it hit, or when the item is equipped.
  ITM files have a similar structure to SPL files.

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/itm_v1.htm
seq:
  - id: magic
    contents: "ITM "
  - id: version
    contents: "V1  "
  - id: unidentified_name_ref
    type: u4
  - id: identified_name_ref
    type: u4
  - id: item_or_sound
    type: item_or_sound
    size: 8
  - id: flags
    type: header_flags
    size: 4
  - id: item_type
    type: u2
    enum: sto::purchased_item_type
  - id: usability_bitmask
    type: usability_flags
    size: 4
  - id: item_animation
    # [TODO] : Wrap table https://gibberlings3.github.io/iesdp/file_formats/ie_formats/itm_v1.htm#Header_Animation
    type: str
    encoding: ASCII
    size: 2
  - id: min_level
    type: u2
  - id: min_strength
    type: u2
  - id: min_strength_bonus
    type: u1
  - id: kit_usability_1
    type: kit_usability_1
    size: 1
  - id: min_intelligence
    type: u1
  - id: kit_usability_2
    type: kit_usability_2
    size: 1
  - id: min_dexterity
    type: u1
  - id: kit_usability_3
    type: kit_usability_3
    size: 1
  - id: min_wisdom
    type: u1
  - id: kit_usability_4
    type: kit_usability_4
    size: 1
  - id: min_constitution
    type: u1
  - id: weapon_proficiency
    type: u1
    enum: weapon_proficiency
  - id: min_charisma
    type: u2
  - id: price
    type: u4
  - id: stack_amount
    type: u2
  - id: inventory_icon_bam
    type: strz
    encoding: ASCII
    size: 8
  - id: lore_to_id
    type: u2
  - id: ground_icon_bam
    type: strz
    encoding: ASCII
    size: 8
  - id: weight
    type: u4
  - id: unidentified_description_ref
    type: u4
  - id: identified_description_ref
    type: u4
  - id: description_icon_bam
    type: strz
    encoding: ASCII
    size: 8
  - id: enchantment
    type: u4
  - id: ofs_extended_headers
    type: u4
  - id: num_extended_headers
    type: u2
  - id: ofs_feature_blocks
    type: u4
  - id: idx_equipping_feature_blocks
    type: u2
  - id: num_equipping_feature_blocks
    type: u2
instances:
  extended_headers:
    type: extended_header
    pos: ofs_extended_headers
    repeat: expr
    repeat-expr: num_extended_headers
  feature_blocks:
    type: eff::header_v1
    pos: ofs_feature_blocks + idx_equipping_feature_blocks * 48  # feature block size
    repeat: expr
    repeat-expr: num_equipping_feature_blocks

types:
  item_or_sound:
    seq:
      - id: bg_replacement_item_itm
        type: strz
        encoding: ASCII
    instances:
      pstee_drop_sound:
        value: bg_replacement_item_itm

  header_flags:
    seq:
      - id: unsellable
        type: b1
      - id: two_handed
        type: b1
      - id: movable_droppable
        type: b1
      - id: displayable
        type: b1
      - id: cursed
        type: b1
      - id: cannot_scribe_to_spellbook
        type: b1
      - id: magical
        type: b1
      - id: left_handed
        type: b1
      - id: silver
        type: b1
      - id: cold_iron
        type: b1
      - id: stolen_off_handed
        type: b1
      - id: conversable_unsellable
        type: b1
      - id: fake_two_handded
        type: b1
      - id: forbid_off_hand_weapon
        type: b1
      - id: usable_in_inventory
        type: b1
      - id: adamantine
        type: b1
      - type: b1
        repeat: expr
        repeat-expr: 8
      - id: undispellable
        type: b1
      - id: toggle_critical_hit_aversion
        type: b1
      - type: b1
        repeat: expr
        repeat-expr: 6

  usability_flags:
    seq:
      - id: chaotic_prefix
        type: b1
      - id: evil
        type: b1
      - id: good
        type: b1
      - id: neutral
        type: b1
      - id: lawful_prefix
        type: b1
      - id: neutral_prefix
        type: b1
      - id: bard
        type: b1
      - id: cleric
        type: b1
      - id: cleric_mage
        type: b1
      - id: cleric_thief
        type: b1
      - id: cleric_ranger
        type: b1
      - id: fighter
        type: b1
      - id: fighter_druid
        type: b1
      - id: fighter_mage
        type: b1
      - id: fighter_cleric
        type: b1
      - id: fighter_mage_cleric
        type: b1
      - id: fighter_mage_thief
        type: b1
      - id: fighter_thief
        type: b1
      - id: mage
        type: b1
      - id: mage_thief
        type: b1
      - id: paladin
        type: b1
      - id: ranger
        type: b1
      - id: thief
        type: b1
      - id: elf
        type: b1
      - id: dwarf
        type: b1
      - id: half_elf
        type: b1
      - id: halfling
        type: b1
      - id: human
        type: b1
      - id: gnome
        type: b1
      - id: monk
        type: b1
      - id: druid
        type: b1
      - id: half_orc
        type: b1

  kit_usability_1:
    seq:
      - id: cleric_of_talos
        type: b1
      - id: cleric_of_helm
        type: b1
      - id: cleric_of_lathlander
        type: b1
      - id: totemic_druid
        type: b1
      - id: shapeshifter_druid
        type: b1
      - id: avenger_druid
        type: b1
      - id: barbarian
        type: b1
      - id: wildmage
        type: b1

  kit_usability_2:
    seq:
      - id: stalker_ranger
        type: b1
      - id: beastmaster_ranger
        type: b1
      - id: assassin_thief
        type: b1
      - id: bounty_hunter_thief
        type: b1
      - id: swashbuckler_thief
        type: b1
      - id: blade_bard
        type: b1
      - id: jester_bard
        type: b1
      - id: skald_bard
        type: b1

  kit_usability_3:
    seq:
      - id: diviner
        type: b1
      - id: enchanter
        type: b1
      - id: illusionist
        type: b1
      - id: invoker
        type: b1
      - id: necromancer
        type: b1
      - id: transmuter
        type: b1
      - id: all
        type: b1
      - id: ferlain
        type: b1

  kit_usability_4:
    seq:
      - id: beserker_fighter
        type: b1
      - id: wizardslayer_fighter
        type: b1
      - id: kensai_fighter
        type: b1
      - id: cavalier_paladin
        type: b1
      - id: inquisiter_paladin
        type: b1
      - id: undead_hunter_paladin
        type: b1
      - id: abjurer
        type: b1
      - id: conjurer
        type: b1

  extended_header:
    seq:
      - id: attack_type
        type: u1
        enum: attack_type
      - id: id_requirement
        type: id_requirement
        size: 1
      - id: location
        type: u1
        enum: location
      - id: alternative_dice_sides
        type: u1
      - id: use_icon
        type: strz
        encoding: ASCII
        size: 8
      - id: target_type
        type: u1
        enum: target_type
      - id: target_count
        type: u1
      - id: range
        type: u2
      - id: launcher_required
        type: u1
        enum: launcher_required
      - id: alternative_dice_thrown
        type: u1
      - id: speed_factor
        type: u1
      - id: alternative_damage_bonus
        type: u1
      - id: thac0_bonus
        type: u2
      - id: dice_sides
        type: u1
      - id: primary_type #mschool.2da
        type: u1
      - id: dice_thrown
        type: u1
      - id: secondary_type #msectype.2da
        type: u1
      - id: damage_bonus
        type: u2
      - id: damage_type
        type: u2
        enum: damage_type
      - id: num_feature_blocks
        type: u2
      - id: idx_feature_blocks
        type: u2
      - id: max_charges
        type: u2
      - id: charge_depletion_behavior
        type: u2
        enum: charge_depletion_behavior
      - id: flags
        type: extension_header_flags
        size: 4
      - id: projectile_animation
        # [TODO] : link projectl.ids/missile.ids
        type: u2
      - id: melee_animation
        # [TODO] : wrap https://gibberlings3.github.io/iesdp/file_formats/ie_formats/itm_v1.htm#ExtendedHeader_MeleeAnimation
        size: 6
      - id: arrow_qualifier
        type: u2
        enum: qualifier
      - id: bolt_qualifier
        type: u2
        enum: qualifier
      - id: bullet_qualifier
        type: u2
        enum: qualifier
    instances:
      feature_blocks:
        type: eff::header_v1
        pos: _root.ofs_feature_blocks + idx_feature_blocks * 48  # feature block size
        repeat: expr
        repeat-expr: num_feature_blocks

  id_requirement:
    seq:
      - id: id_required
        type: b1
      - id: non_id_required
        type: b1

  extension_header_flags:
    seq:
      - id: add_strength_bonus
        type: b1
      - id: breakable
        type: b1
      - id: damage_strength_bonus
        type: b1
      - id: thac0_strength_bonus
        type: b1
      - type: b1
        repeat: expr
        repeat-expr: 5
      - id: breaks_sanctuary_invisibility
        type: b1
      - id: hostile
        type: b1
      - id: recharge_after_resting
        type: b1
      - type: b1
        repeat: expr
        repeat-expr: 13
      - id: tobex_toggle_backstab
        type: b1
      - id: ee_tobex_cannot_target_invisible
        type: b1
      - type: b1
        repeat: expr
        repeat-expr: 5

enums:
  weapon_proficiency:
    0x00: none
    0x59: bastard_sword
    0x5A: long_sword
    0x5B: short_sword
    0x5C: axe
    0x5D: two_handed_sword
    0x5E: katana
    0x5F: scimitar_wakizashi_ninja_to
    0x60: dagger
    0x61: war_hammer
    0x62: spear
    0x63: halberd
    0x64: flail_morningstar
    0x65: mace
    0x66: quarterstaff
    0x67: crossbow
    0x68: long_bow
    0x69: short_bow
    0x6A: darts
    0x6B: sling
    0x6C: blackjack
    0x6D: gun
    0x6E: martial_arts
    0x6F: two_handed_weapon_skill
    0x70: sword_and_shield_skill
    0x71: single_weapon_skill
    0x72: two_weapon_skill
    0x73: club
    0x74: extra_proficiency_2
    0x75: extra_proficiency_3
    0x76: extra_proficiency_4
    0x77: extra_proficiency_5
    0x78: extra_proficiency_6
    0x79: extra_proficiency_7
    0x7A: extra_proficiency_8
    0x7B: extra_proficiency_9
    0x7C: extra_proficiency_10
    0x7D: extra_proficiency_11
    0x7E: extra_proficiency_12
    0x7F: extra_proficiency_13
    0x80: extra_proficiency_14
    0x81: extra_proficiency_15
    0x82: extra_proficiency_16
    0x83: extra_proficiency_17
    0x84: extra_proficiency_18
    0x85: extra_proficiency_19
    0x86: extra_proficiency_20

  attack_type:
    0: none
    1: melee
    2: ranged
    3: magical
    4: launcher

  location:
    0: none
    1: weapon
    2: spell
    3: equipment_item
    4: innate

  target_type:
    0: invalid
    1: living_actor
    2: inventory
    3: dead_actor
    4: any_point_within_range
    5: caster
    6: crash
    7: caster_ee_only

  launcher_required:
    0: none
    1: bow
    2: crossbow
    3: sling
    40: spear
    100: throwing_axe

  damage_type:
    0: none
    1: piercing
    2: crushing
    3: slashing
    4: missile
    5: fist
    6: piercing_crushing_better
    7: piercing_slashing_better
    8: crushing_slashing_worse
    9: blunt_missile

  charge_depletion_behavior:
    0: item_remains
    1: item_vanishes
    2: replace_with_used_up
    3: item_recharges

  qualifier:
    0: no
    1: yes

