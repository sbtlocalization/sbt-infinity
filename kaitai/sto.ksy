# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: @definitelythehuman
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: sto
  file-extension: sto
  endian: le
  bit-endian: le
doc: |
  These files contain a description of the types of items and services available for sale in a given store, inn, tavern, or temple.

doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/sto_v1.htm
seq:
  - id: magic
    contents: "STOR"
  - id: version
    contents: "V1.0"
  - id: type
    type: u4
    enum: store_type
  - id: name_ref
    type: u4
  - id: flags_bits
    type: flags
    size: 4
  - id: sell_markup
    type: u4
  - id: buy_markup
    type: u4
  - id: depreciation_rate
    type: u4
  - id: steal_failure
    type: u2
  - id: capacity
    type: u2
  - id: unknown_1
    size: 8
  - id: offset_purchased_items
    type: u4
  - id: count_purchased_items
    type: u4
  - id: offset_saled_items
    type: u4
  - id: count_saled_items
    type: u4
  - id: lore
    type: u4
  - id: id_price
    type: u4
  - id: rumors_tavern
    type: resref
  - id: offset_drinks
    type: u4
  - id: count_drinks
    type: u4
  - id: rumors_temple
    type: resref
  - id: room_bits
    type: room_flags
    size: 4



enums:
  store_type:
    0x000: store
    0x001: tavern
    0x002: inn
    0x003: temple
    0x005: container

types:
  resref:
    seq:
    - id: res_name
      type: strz
      encoding: ASCII
      size: 8
  
  flags: #What the heck? it's LE but reads like BE???
    seq:
    - id: user_allowed_sell_critical
      type: b1
    - id: toggle_item_recharge
      type: b1
    - id: reputation_not_affect_price
      type: b1
    - id: buy_fenced_goods
      type: b1
    - id: unknown_3
      type: b1
    - id: quality_2
      type: b1
    - id: qualuty_1
      type: b1
    - id: unknown_2
      type: b1
    - id: unknown_1
      type: b1
    - id: user_allowed_purchase_drinks
      type: b1
    - id: user_allowed_purchase_cures
      type: b1
    - id: user_allowed_donate_money
      type: b1
    - id: user_allowed_steal
      type: b1
    - id: user_allowed_identify_items
      type: b1
    - id: user_allowed_sell
      type: b1
    - id: user_allowed_buy
      type: b1
  
  room_flags:
    seq:
    - id: royal
      type: b1
    - id: noble
      type: b1
    - id: merchant
      type: b1
    - id: peasant
      type: b1