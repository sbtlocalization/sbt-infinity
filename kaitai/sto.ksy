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
  - id: steal_failure_chance
    type: u2
  - id: capacity
    type: u2
  - size: 8
  - id: ofs_purchased_items
    type: u4
  - id: num_purchased_items
    type: u4
  - id: ofs_items_for_sale
    type: u4
  - id: num_items_for_sale
    type: u4
  - id: lore
    type: u4
  - id: id_price
    type: u4
  - id: rumours_tavern
    type: strz
    encoding: ASCII
    size: 8
  - id: ofs_drinks
    type: u4
  - id: num_drinks
    type: u4
  - id: rumours_temple
    type: strz
    encoding: ASCII
    size: 8
  - id: room_bits
    type: room_flags
    size: 4
  - id: peasant_room_price
    type: u4
  - id: merchant_room_price
    type: u4
  - id: noble_room_price
    type: u4
  - id: royal_room_price
    type: u4
  - id: ofs_cures
    type: u4
  - id: num_cures
    type: u4
  - size: 36
instances:
  items_for_sale:
    type: item_for_sale_entry
    pos: ofs_items_for_sale
    repeat: expr
    repeat-expr: num_items_for_sale
  drinks:
    type: drink_entry
    pos: ofs_drinks
    repeat: expr
    repeat-expr: num_drinks
  cures:
    type: cure_entry
    pos: ofs_cures
    repeat: expr
    repeat-expr: num_cures
  purchased_items:
    type: u4
    enum: purchased_item_type
    pos: ofs_purchased_items
    repeat: expr
    repeat-expr: num_purchased_items


types:
  flags:
    seq:
    - id: user_allowed_buy
      type: b1
    - id: user_allowed_sell
      type: b1
    - id: user_allowed_identify_items
      type: b1
    - id: user_allowed_steal
      type: b1
    - id: user_allowed_donate_money
      type: b1
    - id: user_allowed_purchase_cures
      type: b1
    - id: user_allowed_purchase_drinks
      type: b1
    - type: b1
    - type: b1
    - id: quality
      type: b1
      repeat: expr
      repeat-expr: 2
    - type: b1
    - id: buy_fenced_goods
      type: b1
    - id: reputation_not_affect_price
      type: b1
    - id: toggle_item_recharge
      type: b1
    - id: user_allowed_sell_critical
      type: b1

  room_flags:
    seq:
    - id: peasant
      type: b1
    - id: merchant
      type: b1
    - id: noble
      type: b1
    - id: royal
      type: b1

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

  item_for_sale_entry:
    seq:
    - id: item_itm
      type: strz
      encoding: ASCII
      size: 8
    - id: expiration_time
      type: u2
    - id: quantity_charges
      type: u2
      repeat: expr
      repeat-expr: 3
    - id: flags
      type: item_flags
      size: 4
    - id: amount_stock
      type: u4
    - id: supply
      type: u4
      enum: supply_type

  drink_entry:
    seq:
    - id: rumours_resourse
      type: strz
      encoding: ASCII
      size: 8
    - id: drink_name_ref
      type: u4
    - id: price
      type: u4
    - id: alcoholic_strenght
      type: u4

  cure_entry:
    seq:
    - id: spell_spl
      type: strz
      encoding: ASCII
      size: 8
    - id: price
      type: u4


enums:
  store_type:
    0: store
    1: tavern
    2: inn
    3: temple
    5: container

  supply_type:
    0: limited
    1: infinite

  purchased_item_type:
    0x00: books_misc
    0x01: amulets_and_necklaces
    0x02: armor
    0x03: belts_and_girdles
    0x04: boots
    0x05: arrows
    0x06: bracers_and_gauntlets
    0x07: headgear_helms_hats_and_other_head_wear
    0x08: keys_not_in_icewind_dale
    0x09: potions
    0x0a: rings
    0x0b: scrolls
    0x0c: shields_not_in_iwd
    0x0d: food
    0x0e: bullets_for_a_sling
    0x0f: bows
    0x10: daggers
    0x11: maces_in_bg_this_includes_clubs
    0x12: slings
    0x13: small_swords
    0x14: large_swords
    0x15: hammers
    0x16: morning_stars
    0x17: flails
    0x18: darts
    0x19: axes
    0x1a: quarterstaff
    0x1b: crossbow
    0x1c: hand_to_hand_weapons
    0x1d: spears
    0x1e: halberds_2_handed_polearms
    0x1f: crossbow_bolts
    0x20: cloaks_and_robes
    0x21: gold_pieces_not_an_inventory
    0x22: gems
    0x23: wands
    0x24: containers_eye_broken_armor
    0x25: books_broken_shields_bracelets
    0x26: familiars_broken_swords_earrings
    0x27: tattoos_pst
    0x28: lenses_pst
    0x29: bucklers_teeth
    0x2a: candles
    0x2b: unknown_0
    0x2c: clubs_iwd
    0x2d: unknown_1
    0x2e: unknown_2
    0x2f: large_shields_iwd
    0x30: unknown_3
    0x31: medium_shields_iwd
    0x32: notes
    0x33: unknown_4
    0x34: unknown_5
    0x35: small_shields_iwd
    0x36: unknown_6
    0x37: telescopes_iwd
    0x38: drinks_iwd
    0x39: great_swords_iwd
    0x3a: container
    0x3b: fur_pelt
    0x3c: leather_armor
    0x3d: studded_leather_armor
    0x3e: chain_mail
    0x3f: splint_mail
    0x40: half_plate
    0x41: full_plate
    0x42: hide_armor
    0x43: robe
    0x44: unknown_7
    0x45: bastard_sword
    0x46: scarf
    0x47: food_iwd2
    0x48: hat
    0x49: gauntlet
