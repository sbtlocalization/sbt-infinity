# SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: wavc
  file-extension: wav
  endian: le
  bit-endian: le
doc: |
  WAVC is a container format used by Infinity Engine games to store compressed audio. It wraps an ACM-compressed audio
  stream with a header that contains audio metadata (channels, sample rate, bits per sample) and size information.
  Unlike ACM files, WAVC can be played from either the override folder or BIF files. WAVC files are 22050 Hz, 16 bits,
  and must be renamed to *.WAV before being used in the game. The Infinity Engine plays WAVC files at twice the speed
  of the original WAV file.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/wavc_v1.htm
seq:
  - id: magic
    contents: "WAVC"
  - id: version
    contents: "V1.0"
  - id: uncompressed_size
    type: u4
  - id: len_acm_data
    type: u4
  - id: acm_offset
    type: u4
  - id: num_channels
    type: u2
  - id: bits_per_sample
    type: u2
  - id: sample_rate
    type: u2
  - id: reserved
    type: u2
instances:
  acm_data:
    pos: acm_offset
    size: len_acm_data
