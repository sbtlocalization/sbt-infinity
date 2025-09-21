# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
#
# SPDX-License-Identifier: GPL-3.0-only

meta:
  id: dlg
  title: DLG v1
  file-extension: dlg
  ks-version: "0.10"
  endian: le
  bit-endian: le
doc: |
  DLG files contain the structure of conversations, in what is effectievly a state machine. Dialogs
  contains string references into the TLK file that make up the actual words of the conversation.
  Dialogs bear similarities to scripts; each state may have a series of trigger conditions, and
  effect a series of actions. If the any of the triggers for a state evaluate to false, the state
  is skipped and the triggers in the next state are evaluated - this occurs when entering into a
  dialog state, and when presenting a list of responses.

  ```
  state 0:
      trigger: NumTimesTalkedTo(0)
      Text: "Hello, sailor!"

  state 1:
      trigger: NumTimesTalkedToGT(5)
      Text: "Go away, already!"

  state 2:
      Text: "Hail and well met, yada yada yada."
  ```
  Dialog always attempt to start at state 0. The first time this sample dialog is entered the
  trigger in state 0 is true, hence the character responds "Hello, sailor!". Subsequent times the
  dialog is entered the trigger in state 0 will be false, and state 1 is evaluated - this trigger
  also fails and so state 2 is evaluated. This state evaluates true, and get the associated message
  is displayed. If the dialog is initiaed five or more times, the trigger in state 1 will evaluate
  to true and the message associated with that state will be displayed.

  In addition to the triggers outlined above, states present a list of responses (aka transitions).
  Each response may have a series of behaviours associated with it; the response text, a journal
  entry or an action.
doc-ref: |
  https://gibberlings3.github.io/iesdp/file_formats/ie_formats/dlg_v1.htm

seq:
  - id: magic
    contents: "DLG "
  - id: version
    contents: "V1.0"
  - id: num_states
    type: u4
  - id: ofs_states
    type: u4
  - id: num_transitions
    type: u4
  - id: ofs_transitions
    type: u4
  - id: ofs_state_triggers
    type: u4
  - id: num_state_triggers
    type: u4
  - id: ofs_transition_triggers
    type: u4
  - id: num_transition_triggers
    type: u4
  - id: ofs_actions
    type: u4
  - id: num_actions
    type: u4
  - id: threat_flags
    size: 4
    type: header_flags
instances:
  states:
    pos: ofs_states
    type: state_entry
    repeat: expr
    repeat-expr: num_states
  transitions:
    pos: ofs_transitions
    type: transition_entry
    repeat: expr
    repeat-expr: num_transitions
  state_triggers:
    pos: ofs_state_triggers
    type: text_entry
    repeat: expr
    repeat-expr: num_state_triggers
  transition_triggers:
    pos: ofs_transition_triggers
    type: text_entry
    repeat: expr
    repeat-expr: num_transition_triggers
  actions:
    pos: ofs_actions
    type: text_entry
    repeat: expr
    repeat-expr: num_actions
types:
  header_flags:
    seq:
      - id: turn_hostile
        type: b1
      - id: escape_area
        type: b1
      - id: do_nothing
        type: b1
  state_entry:
    seq:
      - id: text_ref
        type: u4
      - id: first_transition_index
        type: u4
      - id: num_transitions
        type: u4
      - id: state_trigger_index
        type: u4
    instances:
      transitions:
        pos: _root.ofs_transitions + first_transition_index * 32  # sizeof(transition_entry)
        type: transition_entry
        repeat: expr
        repeat-expr: num_transitions
      trigger:
        pos: _root.ofs_state_triggers + state_trigger_index * 8  # sizeof(text_entry)
        type: text_entry
        if: state_trigger_index != 0xFFFFFFFF
  transition_entry:
    seq:
      - id: flags
        size: 4
        type: flags
      - id: text_ref
        type: u4
      - id: journal_text_ref
        type: u4
      - id: transition_trigger_index
        type: u4
      - id: transition_action_index
        type: u4
      - id: next_state_resource
        type: strz
        size: 8
        encoding: ASCII
      - id: next_state_index
        type: u4
    instances:
      trigger:
        pos: _root.ofs_transition_triggers + transition_trigger_index * 8  # sizeof(text_entry)
        type: text_entry
        if: flags.with_trigger
      action:
        pos: _root.ofs_actions + transition_action_index * 8  # sizeof(text_entry)
        type: text_entry
        if: flags.with_action
    types:
      flags:
        seq:
          - id: with_text
            type: b1
          - id: with_trigger
            type: b1
          - id: with_action
            type: b1
          - id: dialog_end
            type: b1
          - id: with_journal_entry
            type: b1
          - id: interrupt
            type: b1
          - id: add_unsolved_quest
            type: b1
          - id: add_journal_note
            type: b1
          - id: add_solved_quest
            type: b1
          - id: immediate_action
            type: b1
          - id: clear_actions
            type: b1
  text_entry:
    seq:
      - id: ofs_text
        type: u4
      - id: len_text
        type: u4
    instances:
      text:
        pos: ofs_text
        size: len_text
        type: str
        encoding: ASCII
        io: _root._io
