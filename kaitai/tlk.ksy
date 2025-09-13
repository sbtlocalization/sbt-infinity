# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: tlk
  file-extension: tlk
  endian: le
  bit-endian: le
doc: |
  Most strings shown in Infinity Engine games are stored in a TLK file, usually dialog.tlk (for 
  male/default text) and/or dialogf.tlk (for female text). Strings are stored with associated 
  information (e.g. a reference to sound file), and are indexed by a (0-indexed) 32 bit identigier 
  called a "Strref" (String Reference). Storing text in this way allows for a game to be easily 
  swapped between languages.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/tlk_v1.htm
seq:
  - id: header
    type: header
  - id: entries
    type: string_entry
    repeat: expr
    repeat-expr: header.string_count
types:
  header:
    seq:
      - id: magic
        contents: "TLK "
      - id: version
        contents: "V1  "
      - id: lang
        type: u2
      - id: string_count
        type: u4
      - id: data_offset
        type: u4
  string_entry:
    seq:
      - id: flags
        size: 2
        type: flags
      - id: audio_name
        type: str
        size: 8
        encoding: ASCII
      - id: volume_variance
        type: u4
      - id: pitch_variance
        type: u4
      - id: string_offset
        type: u4
      - id: string_length
        type: u4
    instances:
      text:
        pos: _root.header.data_offset + string_offset
        size: string_length
        type: str
        encoding: UTF-8
    types:
      flags:
        seq:
          - id: no_message
            type: b1
          - id: text_exists
            type: b1
          - id: sound_exists
            type: b1
          - id: standard_message
            type: b1
          - id: token_exists
            type: b1
