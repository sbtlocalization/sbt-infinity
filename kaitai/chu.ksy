# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: chu
  title: CHU v1
  file-extension: chu
  ks-version: "0.11"
  endian: le
  bit-endian: le
doc: |
  This file format describes the layout of the GUI screens (the graphics for the screens are held
  in MOS and BAM files).
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/chu_v1.htm
seq:
  - id: magic
    contents: "CHUI"
    size: 4
  - id: version
    contents: "V1  "
    size: 4
  - id: num_windows
    type: u4
  - id: ofs_controls
    type: u4
  - id: ofs_windows
    type: u4
instances:
  windows:
    pos: ofs_windows
    type: window
    repeat: expr
    repeat-expr: num_windows
types:
  window:
    seq:
      - id: win_id
        type: u2
      - size: 2
      - id: x
        type: u2
      - id: y
        type: u2
      - id: width
        type: u2
      - id: height
        type: u2
      - id: flags
        type: flags
        size: 2
      - id: num_controls
        type: u2
      - id: background_mos
        type: strz
        size: 8
        encoding: ASCII
      - id: first_control_index
        type: u2
      - id: options
        type: options
        size: 2
    instances:
      controls:
        pos: _parent.ofs_controls + first_control_index.as<u4> * 8 # size of control
        type: control
        repeat: expr
        repeat-expr: num_controls
    types:
      flags:
        seq:
          - id: has_background
            type: b1
      options:
        seq:
          - id: do_not_dim_background
            type: b1
  control:
    seq:
      - id: ofs_data
        type: u4
      - id: len_data
        type: u4
    instances:
      data:
        pos: ofs_data
        size: len_data
        type: control_struct
    types:
      control_struct:
        seq:
          - id: control_id
            type: u2
          - size: 2
          - id: x
            type: u2
          - id: y
            type: u2
          - id: width
            type: u2
          - id: height
            type: u2
          - id: type
            type: u1
            enum: struct_type
          - size: 1
          - id: properties
            type:
              switch-on: type
              cases:
                struct_type::button: button
                struct_type::slider: slider
                struct_type::text_edit: text_edit
                struct_type::text_area: text_area
                struct_type::label: label
                struct_type::scrollbar: scrollbar
        types:
          button:
            seq:
              - id: image_bam
                type: strz
                size: 8
                encoding: ASCII
              - id: animation_cycle
                type: u1
              - id: flags
                type: flags
                size: 1
              - id: normal_frame_index
                type: u1
              - id: anchor_x1
                type: u1
              - id: pressed_frame_index
                type: u1
              - id: anchor_x2
                type: u1
              - id: selected_frame_index
                type: u1
              - id: anchor_y1
                type: u1
              - id: disabled_frame_index
                type: u1
              - id: anchor_y2
                type: u1
            types:
              flags:
                doc: |
                  With no bits set the text is centred (horizontally and vertically)
                seq:
                  - id: align_left
                    type: b1
                  - id: align_right
                    type: b1
                  - id: align_top
                    type: b1
                  - id: align_bottom
                    type: b1
                  - id: anchor_to_zero
                    type: b1
                  - id: reduce_text_size
                    type: b1
          slider:
            seq:
              - id: background_image_mos
                type: strz
                size: 8
                encoding: ASCII
              - id: knob_image_bam
                type: strz
                size: 8
                encoding: ASCII
              - id: animation_cycle
                type: u2
              - id: normal_slider_frame_index
                type: u2
              - id: grabbed_slider_frame_index
                type: u2
              - id: knob_x
                type: u2
              - id: knob_y
                type: u2
              - id: step_width
                type: u2
              - id: step_count
                type: u2
              - id: region_top
                type: u2
              - id: region_bottom
                type: u2
              - id: region_left
                type: u2
              - id: region_right
                type: u2
          text_edit:
            seq:
              - id: background_image_1_mos
                type: strz
                size: 8
                encoding: ASCII
              - id: background_image_2_mos
                type: strz
                size: 8
                encoding: ASCII
              - id: background_image_3_mos
                type: strz
                size: 8
                encoding: ASCII
              - id: cursor_bam
                type: strz
                size: 8
                encoding: ASCII
              - id: cursor_animation_cycle
                type: u2
              - id: cursor_frame_index
                type: u2
              - id: x
                type: u2
              - id: y
                type: u2
              - id: attached_scrollbar_id
                type: u2
              - size: 2
              - id: font
                type: strz
                size: 8
                encoding: ASCII
              - size: 2
              - id: initial_text
                type: strz
                size: 32
                encoding: UTF-8
              - id: max_length
                type: u2
              - id: text_case
                type: u4
                enum: text_case
            enums:
              text_case:
                0: sentence_case
                1: upper_case
                2: lower_case
          text_area:
            seq:
              - id: initials_font
                type: strz
                size: 8
                encoding: ASCII
              - id: main_text_font
                type: strz
                size: 8
                encoding: ASCII
              - id: text_color
                type: u4
              - id: initials_color
                type: u4
              - id: outline_color
                type: u4
              - id: attached_scrollbar_id
                type: u2
              - size: 2
          label:
            seq:
              - id: initial_text_ref
                type: u4
              - id: font
                type: strz
                size: 8
                encoding: ASCII
              - id: text_color
                type: u4
              - id: outline_color
                type: u4
              - id: flags
                type: flags
                size: 2
            types:
              flags:
                seq:
                  - id: use_rgb_colors
                    type: b1
                  - id: true_color
                    type: b1
                  - id: align_hcenter
                    type: b1
                  - id: align_left
                    type: b1
                  - id: alight_right
                    type: b1
                  - id: align_top
                    type: b1
                  - id: align_vcenter
                    type: b1
                  - id: align_bottom
                    type: b1
          scrollbar:
            seq:
              - id: image_bam
                type: strz
                size: 8
                encoding: ASCII
              - id: animation_cycle
                type: u2
              - id: normal_up_arrow_frame_index
                type: u2
              - id: pressed_up_arrow_frame_index
                type: u2
              - id: normal_down_arrow_frame_index
                type: u2
              - id: pressed_down_arrow_frame_index
                type: u2
              - id: through_frame_index
                type: u2
              - id: slider_frame_index
                type: u2
              - id: attached_text_area_id
                type: u2
              - size: 2
        enums:
          struct_type:
            0: button
            2: slider
            3: text_edit
            5: text_area
            6: label
            7: scrollbar
