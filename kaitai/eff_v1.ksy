meta:
  id: eff_v1
  title: EFF v1.0
  file-extension: eff_v1
  endian: le
  bit-endian: le
  imports:
    - eff

doc: |
  This file format describes an effect (opcode) and its parameters.
  The format is only ever found embedded in other files (e.g. ITM or SPL).
  The engine appears to roll a probability for each valid target type, rather than one probability per attack.

doc-ref: |
 https://gibberlings3.github.io/iesdp/file_formats/ie_formats/eff_v1.htm

types:
  header:
    seq:
      - id: opcode_number
        type: u2
      - id: target_type
        type: u1
        enum: eff::body_v2::target_type
      - id: power
        type: u1
        # [TODO] : Link IDS files to parameter_1 and parameter_2
      - id: parameter_1 #IDS value
        type: u4
      - id: parameter_2 #IDS target
        type: u4
      - id: timing_mode
        type: u1
        enum: eff::body_v2::timing_mode
      - id: dispel_resistance
        type: dispel_resistance
        size: 1
      - id: duration
        type: u4
      - id: probability1
        type: u1
      - id: probability2
        type: u1
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
        type: eff::body_v2::saving_throw_type
      - id: saving_throw_bonus
        type: u4
      - id: tobex_stacking_id
        type: u4

  dispel_resistance:
    seq:
      - id: dispel
        type: b1
      - id: bypass_resistance
        type: b1
