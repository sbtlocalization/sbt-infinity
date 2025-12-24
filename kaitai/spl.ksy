# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: spl
  title: SPL v1
  file-extension: spl
  endian: le
  bit-endian: le
  imports:
    - itm
    - eff

doc: |
  This file format describes a "spell". Spells include mage spells, priest spells, innate abilities, special abilities and effects used for game advancement (e.g. animation effects, custom spells).

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/spl_v1.htm
seq:
  - id: magic
    contents: "SPL "
  - id: version
    contents: "V1  "
  - id: unidentified_name_ref
    type: u4
  - id: identified_name_ref
    type: u4
    doc: Unused.
  - id: casting_sound_wav
    type: strz
    size: 8
    encoding: ASCII
  - id: flags
    size: 4
    type: flags
  - id: spell_type
    type: u2
    enum: spell_type
  - id: exclusion_flags
    type:
      switch-on: spell_type
      cases:
        "spell_type::priest": exclusion_flags_priest
        _: exclusion_flags
    size: 4
  - id: casting_animation
    type: u2
  - size: 1
  - id: primary_school
    type: u1
    doc: See `school.2da`/`mschool.2da`.
  - id: min_strength
    type: u1
  - id: secondary_type
    type: u1
    doc: See `msectype.2da`.
  - size: 12
  - id: spell_level
    type: u4
  - size: 2
  - id: spellbook_icon_bam
    type: strz
    size: 8
    encoding: ASCII
  - size: 2
  - size: 8
  - size: 4
  - id: unidentified_description_ref
    type: u4
  - id: identified_description_ref
    type: u4
    doc: Unused.
  - id: description_icon_bam
    doc: Unused.
    type: strz
    size: 8
    encoding: ASCII
  - size: 4
  - id: ofs_extended_header
    type: u4
  - id: num_extended_header
    type: u2
  - id: ofs_effects
    type: u4
  - id: first_effect_index
    type: u2
  - id: num_effects
    type: u2
instances:
  extended_headers:
    pos: ofs_extended_header
    size: 40
    type: extended_header
    repeat: expr
    repeat-expr: num_extended_header
  effects:
    pos: ofs_effects + first_effect_index * 48 # eff::header_v1 size
    size: 48
    type: eff::header_v1
    repeat: expr
    repeat-expr: num_effects
types:
  extended_header:
    seq:
      - id: spell_form
        type: u1
        enum: spell_form
      - id: type_flags
        type: type_flags
        size: 1
      - id: location
        type: u2
        enum: itm::location
      - id: icon_bam
        type: strz
        size: 8
        encoding: ASCII
      - id: target
        type: u1
        enum: itm::target_type
      - id: target_count
        type: u1
      - id: range
        type: u2
      - id: minimum_level
        type: u2
      - id: casting_speed
        type: u2
      - id: times_per_day
        type: u2
      - size: 8
      - id: num_effects
        type: u2
      - id: first_effect_index
        type: u2
      - size: 4
      - id: projectile_pro
        type: u2
        doc: See `projectl.ids`
    instances:
      effects:
        pos: _root.ofs_effects + first_effect_index * 48
        io: _root._io
        size: 0x30
        type: eff::header_v1
        repeat: expr
        repeat-expr: num_effects
    types:
      type_flags:
        seq:
          - id: usable_after_id
            type: b1
          - id: usable_before_id
            type: b1
          - id: pst_non_hostile_projectile
            type: b1
    enums:
      spell_form:
        0: none
        1: melee
        2: ranged
        3: magical
        4: launcher
  flags:
    seq:
      - type: b9
      - id: breaks_sanctuary_invisibility
        type: b1
      - id: hostile
        type: b1
      - id: no_los_required
        type: b1
      - id: allow_spotting
        type: b1
      - id: outdoors_only
        type: b1
      - id: ignore_dead_magic_and_wild_surge_effect
        type: b1
      - id: ignore_wild_surge_effect
        type: b1
      - id: not_in_combat
        type: b1
      - type: b7
      - id: can_target_invisible
        type: b1
      - id: castable_when_silenced
        type: b1
      - type: b6
  exclusion_flags:
    seq:
      - id: berserker
        type: b1
      - id: wizard_slayer
        type: b1
      - id: kensai
        type: b1
      - id: cavalier
        type: b1
      - id: inquisitor
        type: b1
      - id: undead_hunter
        type: b1
      - id: abjurer
        type: b1
      - id: conjurer
        type: b1
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
      - id: generalist
        type: b1
      - id: archer
        type: b1
      - id: stalker
        type: b1
      - id: beastmaster
        type: b1
      - id: assassin
        type: b1
      - id: bounty_hunter
        type: b1
      - id: swashbuckler
        type: b1
      - id: blade
        type: b1
      - id: jester
        type: b1
      - id: skald
        type: b1
      - id: cleric_of_talos
        type: b1
      - id: cleric_of_helm
        type: b1
      - id: cleric_of_lathander
        type: b1
      - id: totemic_druid
        type: b1
      - id: shapeshifter
        type: b1
      - id: avenger
        type: b1
      - id: barbarian
        type: b1
      - id: wild_mage
        type: b1
  exclusion_flags_priest:
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
      - type: b24
      - id: cleric_paladin
        type: b1
      - id: druid_ranger_shaman
        type: b1
enums:
  spell_type:
    0: special
    1: wizard
    2: priest
    3: psionic
    4: innate
    5: bard_song
