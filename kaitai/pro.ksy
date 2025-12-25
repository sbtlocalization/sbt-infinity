# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: pro
  title: PRO v1
  file-extension: pro
  endian: le
  bit-endian: le

doc: |
  This file format describes projectiles, and the files are referenced spells and projectile weapons.
  Projectile files can control:

  - Projectile graphics
  - Projectile speed
  - Projectile area of effect
  - Projectile sound

  These files have constant length 256 bytes (no BAM), 512 bytes (single target), or 768 bytes (area of effect).

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/pro_v1.htm
seq:
  - id: magic
    contents: "PRO "
  - id: version
    contents: "V1.0"
  - id: projectile_type
    type: u2
    enum: projectile_type
  - id: speed
    type: u2
  - id: flags
    type: flags
    size: 4
  - id: fire_sound_wav
    type: strz
    size: 8
    encoding: ASCII
  - id: impact_sound_wav
    type: strz
    size: 8
    encoding: ASCII
  - id: source_animation
    doc: VEF > VVC > BAM
    type: strz
    size: 8
    encoding: ASCII
  - id: particle_color
    type: u2
    doc: See `SPRKCLR.2DA`.
  - id: projectile_width
    type: u2
  - id: extended_flags
    size: 4
    type: extended_flags
  - id: message_ref
    type: u4
  - id: pulse_color
    type: rgba_color
  - id: color_speed
    type: u2
  - id: screen_shake_amount
    type: u2
  - size: 2
  - id: ids_target_1
    type: u2
  - size: 2
  - id: ids_target_2
    type: u2
  - id: default_spell_spl
    type: strz
    size: 8
    encoding: ASCII
  - id: success_spell_spl
    type: strz
    size: 8
    encoding: ASCII
  - id: pst_angle_increase_minimum
    type: u2
  - id: pst_angle_increase_maximum
    type: u2
  - id: pst_curve_minimum
    type: u2
  - id: pst_curve_maximum
    type: u2
  - id: pst_thac0_bonus
    type: u2
  - id: pst_thac0_bonus_non_actor
    type: u2
  - id: pst_radius_minimum
    type: u2
  - id: pst_radius_maximum
    type: u2
  - size: 156
  - id: projectile_info
    type: projectile_info
    if: |
      projectile_type == projectile_type::single_target or
      projectile_type == projectile_type::area_of_effect
  - id: area_effect_info
    type: area_effect_info
    if: |
      projectile_type == projectile_type::area_of_effect

types:
  rgba_color:
    seq:
      - id: red
        type: u1
      - id: green
        type: u1
      - id: blue
        type: u1
      - id: alpha
        type: u1
  flags:
    seq:
      - id: show_sparkle
        type: b1
      - id: use_height
        type: b1
      - id: loop_fire_sound
        type: b1
      - id: loop_impact_sound
        type: b1
      - id: ignore_center
        type: b1
      - id: draw_as_background
        type: b1
      - id: allow_saving
        type: b1
      - id: loop_spread_animation
        type: b1
  extended_flags:
    seq:
      - id: bounce_from_walls
        type: b1
      - id: pass_target
        type: b1
      - id: draw_center_vvc_once
        type: b1
      - id: hit_immemdiately
        type: b1
      - id: face_target
        type: b1
      - id: curved_path
        type: b1
      - id: start_random_frame
        type: b1
      - id: pillar
        type: b1
      - id: semi_transparent_trail_puff_vef
        type: b1
      - id: tinted_trail_puff_vef
        type: b1
      - id: multiple_projectiles
        type: b1
      - id: default_spell_on_missed
        type: b1
      - id: falling_path
        type: b1
      - id: comet
        type: b1
      - id: lined_up_aoe
        type: b1
      - id: rectangular_aoe
        type: b1
      - id: draw_behind_target
        type: b1
      - id: casing_flow_effect
        type: b1
      - id: travel_door
        type: b1
      - id: stop_fade_after_hit
        type: b1
      - id: display_message
        type: b1
      - id: random_path
        type: b1
      - id: start_random_sequence
        type: b1
      - id: color_pulse_on_hit
        type: b1
      - id: touch_projectile
        type: b1
      - id: negate_ids1
        type: b1
      - id: negate_ids2
        type: b1
      - id: use_either_ids
        type: b1
      - id: delayed_payload
        type: b1
      - id: limited_path_count
        type: b1
      - id: iwd_style_check
        type: b1
      - id: caster_affected
        type: b1
  projectile_info:
    seq:
      - id: flags
        type: flags
        size: 4
      - id: projectile_animation_bam
        type: strz
        size: 8
        encoding: ASCII
      - id: shadow_animation_bam
        type: strz
        size: 8
        encoding: ASCII
      - id: projectile_animation_number
        type: u1
      - id: shadow_animation_number
        type: u1
      - id: light_spot_intensity
        type: u2
      - id: light_spot_width
        type: u2
      - id: light_spot_height
        type: u2
      - id: palette_bmp
        type: strz
        size: 8
        encoding: ASCII
      - id: projectile_colors
        type: u1
        repeat: expr
        repeat-expr: 7
      - id: smoke_puff_delay
        type: u1
      - id: smoke_colors
        type: u1
        repeat: expr
        repeat-expr: 7
      - id: face_target_granularity_mirroring
        type: u1
        enum: granularity_mirroring
      - id: smoke_animation
        type: u2
        doc: See `ANIMATE.IDS`.
      - id: trailing_animations_bam
        type: strz
        size: 8
        encoding: ASCII
        repeat: expr
        repeat-expr: 3
      - id: trailing_animation_delays
        type: u2
        repeat: expr
        repeat-expr: 3
      - id: trail_flags
        type: trail_flags
        size: 4
      - size: 168

    types:
      flags:
        seq:
          - id: enable_bam_coloring
            type: b1
          - id: create_smoke
            type: b1
          - id: colored_smoke
            type: b1
          - id: not_light_source
            type: b1
          - id: modify_for_height
            type: b1
          - id: casts_shadows
            type: b1
          - id: enable_light_spot
            type: b1
          - id: translucent
            type: b1
          - id: mid_level_brighten
            type: b1
          - id: blended
            type: b1
      trail_flags:
        seq:
          - id: puff_at_target
            type: b1
          - id: puff_at_source
            type: b1
    enums:
      granularity_mirroring:
        0: do_not_mirror
        1: do_not_face_target
        5: mirrored_east_reduced
        9: mirrored_east_full
        16: not_mirrored_full
  area_effect_info:
    seq:
      - id: flags
        size: 2
        type: flags
      - id: ray_count
        type: u2
      - id: trap_size
        type: u2
      - id: explosion_size
        type: u2
      - id: explosion_sound_wav
        type: strz
        size: 8
        encoding: ASCII
      - id: explosion_frequency
        type: u2
      - id: fragment_animation
        type: u2
        doc: See `ANIMATE.IDS`.
      - id: secondary_projectile
        type: u2
        doc: See `PROJECTL.IDS`.
      - id: num_repetitions
        type: u1
      - id: explosion_effect
        type: u1
        doc: See `FIREBALL.IDS`.
      - id: explosion_color
        type: u1
      - size: 1
      - id: explosion_projectile
        type: u2
        doc: See `PROJECTL.IDS`.
      - id: explosion_animation
        type: strz
        size: 8
        encoding: ASCII
      - id: cone_width
        type: u2
      - id: rotate_rays_clockwise
        type: u2
      - id: spread_animation
        type: strz
        size: 8
        encoding: ASCII
      - id: ring_animation
        type: strz
        size: 8
        encoding: ASCII
      - id: area_sound_wav
        type: strz
        size: 8
        encoding: ASCII
      - id: extended_flags
        size: 4
        type: extended_flags
      - id: num_dice_for_multiple_targets
        type: u2
      - id: dice_size_for_multiple_targets
        type: u2
      - id: animation_granularity
        type: u2
      - id: animation_granularity_divider
        type: u2
      - size: 180

    types:
      flags:
        seq:
          - id: trap_visible
            type: b1
          - id: triggered_by_inanimate_objects
            type: b1
          - id: triggered_by_condition
            type: b1
          - id: delayed_trigger
            type: b1
          - id: use_secondary_projectile
            type: b1
          - id: use_fragment_graphics
            type: b1
          - id: affect_only_enemies
            type: b1
          - id: affect_only_allies
            type: b1
          - id: mage_level_duration
            type: b1
          - id: cleric_level_duration
            type: b1
          - id: draw_animation
            type: b1
          - id: cone_shaped
            type: b1
          - id: ignore_los
            type: b1
          - id: delayed_explosion
            type: b1
          - id: skip_first_condition
            type: b1
          - id: single_target
            type: b1
      extended_flags:
        seq:
          - id: paletted_ring
            type: b1
          - id: random_speed
            type: b1
          - id: start_scattered
            type: b1
          - id: paletted_center
            type: b1
          - id: repeat_scattering
            type: b1
          - id: paletted_animation
            type: b1
          - type: b3
          - id: oriented_fireball_puffs
            type: b1
          - id: use_hit_dice_lookup
            type: b1
          - type: b2
          - id: blend_area_ring_animation
            type: b1
          - id: glow_area_ring_animation
            type: b1
          - id: hit_point_limit
            type: b1

enums:
  projectile_type:
    0: no_projectile
    1: no_bam
    2: single_target
    3: area_of_effect
