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
  - id: magic
    contents: "TLK "
  - id: version
    contents: "V1  "
  - id: lang
    type: u2
  - id: num_entries
    type: u4
  - id: ofs_data
    type: u4
  - id: entries
    type: string_entry
    repeat: expr
    repeat-expr: num_entries
types:
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
      - id: ofs_string
        type: u4
      - id: len_string
        type: u4
    instances:
      text:
        pos: _root.ofs_data + ofs_string
        size: len_string
        type: str
        terminator: 0
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
