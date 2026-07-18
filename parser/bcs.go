// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ScriptCall struct {
	StrRef uint32
	Call   string
}

type bcsObject struct {
	// enumerated specification fields (EA.IDS, GENERAL.IDS, …); all zero for
	// a plain object reference
	ea, general, race, class, specific, gender, alignment int32
	identifiers                                           int32 // OBJECT.IDS value (e.g. LastTrigger, Myself)
	name                                                  string
}

type bcsAction struct {
	id   int32
	o1   bcsObject // reserved for an ActionOverride target
	o2   bcsObject // the normal single-object slot
	o3   bcsObject
	ints [5]int32 // int0, x, y, int1, int2
}

// The point coordinates occupy physical fields 1 and 2 of bcsAction.ints, so
// the logical integer slots land on 0, 3 and 4.
var intSlotField = map[int]int{0: 0, 1: 3, 2: 4}

// A malformed action or block is skipped rather than aborting the parse; an
// error is returned only when the input cannot be read at all.
func DecodeCompiledScript(r io.Reader, actionIds *Ids, family DisplayFamilyActions, objectIds *Ids) ([]ScriptCall, error) {
	tokens, err := tokenizeBcs(r)
	if err != nil {
		return nil, err
	}

	p := &bcsParser{
		tokens:    tokens,
		actionIds: actionIds,
		family:    family,
		objectIds: objectIds,
	}
	return p.parseScript(), nil
}

type bcsParser struct {
	tokens []string
	pos    int

	actionIds *Ids
	family    DisplayFamilyActions
	objectIds *Ids

	calls []ScriptCall
}

// Block markers glued to a preceding number are split apart ("100AC" →
// "100", "AC"); quoted strings stay single tokens so empty strings and
// embedded spaces survive.
func tokenizeBcs(r io.Reader) ([]string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return splitBcsTokens(sb.String()), nil
}

var bcsMarkers = []string{"SC", "CR", "CO", "TR", "RS", "RE", "AC", "OB"}

func splitBcsTokens(s string) []string {
	var tokens []string
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == '"':
			end := i + 1
			for end < len(s) && s[end] != '"' {
				end++
			}
			if end < len(s) {
				end++ // include the closing quote
			}
			tokens = append(tokens, s[i:end])
			i = end
		case isBcsSpace(ch):
			i++
		default:
			end := i
			for end < len(s) && !isBcsSpace(s[end]) && s[end] != '"' {
				end++
			}
			tokens = append(tokens, splitGluedMarker(s[i:end])...)
			i = end
		}
	}
	return tokens
}

func isBcsSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

// At most one trailing marker is split off — the format never glues more
// than one.
func splitGluedMarker(run string) []string {
	for _, m := range bcsMarkers {
		if run == m {
			return []string{run}
		}
		if len(run) > len(m) && strings.HasSuffix(run, m) {
			prefix := run[:len(run)-len(m)]
			if isAllDigitsOrSign(prefix) {
				return []string{prefix, m}
			}
		}
	}
	return []string{run}
}

func isAllDigitsOrSign(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			continue
		}
		if i == 0 && (ch == '-' || ch == '+') {
			continue
		}
		return false
	}
	return true
}

func (p *bcsParser) peek() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	return p.tokens[p.pos], true
}

func (p *bcsParser) next() (string, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

// Only AC action blocks contribute script context. An integer glued before
// the marker (e.g. "100AC") is the enclosing response's weight, not the
// action id — the id is the first field of the block's own body. Consecutive
// actions in one response appear as adjacent "AC <body1> AC AC <body2> AC"
// pairs, so scanning for a bare AC marker and skipping everything else
// handles them naturally.
func (p *bcsParser) parseScript() []ScriptCall {
	for {
		tok, ok := p.next()
		if !ok {
			break
		}
		if tok == "AC" {
			p.parseActionBlock()
		}
	}
	return p.calls
}

// A malformed block is consumed up to its closing AC (or end of stream) and
// skipped so the rest of the script still parses.
func (p *bcsParser) parseActionBlock() {
	action, ok := p.readAction()
	if !ok {
		return
	}

	identifier, ok := p.actionIds.Entries[action.id]
	if !ok {
		return
	}
	// ACTION.IDS stores the full signature string (e.g.
	// "DisplayString(O:Object*,I:StrRef*)"); the display family is keyed by
	// the bare action name.
	sig, err := ParseActionSignature(identifier)
	if err != nil {
		return
	}
	familySig, ok := p.family[sig.Name]
	if !ok {
		// Not a display-family action: contributes no script context.
		return
	}

	field, ok := intSlotField[familySig.StrRefIntSlot]
	if !ok {
		return
	}
	strref := uint32(action.ints[field])

	p.calls = append(p.calls, ScriptCall{
		StrRef: strref,
		Call:   p.renderCall(sig, action, strref),
	})
}

func (p *bcsParser) readAction() (bcsAction, bool) {
	var a bcsAction

	idTok, ok := p.next()
	if !ok {
		return a, false
	}
	id, err := strconv.ParseInt(idTok, 0, 64)
	if err != nil {
		p.skipToActionEnd()
		return a, false
	}
	a.id = int32(id)

	objs := [...]*bcsObject{&a.o1, &a.o2, &a.o3}
	for _, obj := range objs {
		if !p.expect("OB") {
			p.skipToActionEnd()
			return a, false
		}
		o, ok := p.readObjectBody()
		if !ok {
			p.skipToActionEnd()
			return a, false
		}
		*obj = o
	}

	// Five integer fields: int0, x, y, int1, int2.
	for i := range a.ints {
		v, ok := p.readInt()
		if !ok {
			p.skipToActionEnd()
			return a, false
		}
		a.ints[i] = v
	}

	// Two string parameters (read and discarded).
	for i := 0; i < 2; i++ {
		if _, ok := p.readString(); !ok {
			p.skipToActionEnd()
			return a, false
		}
	}

	if !p.expect("AC") {
		p.skipToActionEnd()
		return a, false
	}

	return a, true
}

// The OB fields after the eight enumerated/identifier integers vary by game,
// so the reader is tolerant: it consumes integers and an optional [x.y]
// point until it reaches the name string, then the closing OB marker.
func (p *bcsParser) readObjectBody() (bcsObject, bool) {
	var o bcsObject

	fields := [...]*int32{
		&o.ea, &o.general, &o.race, &o.class,
		&o.specific, &o.gender, &o.alignment, &o.identifiers,
	}
	for _, f := range fields {
		v, ok := p.readInt()
		if !ok {
			return o, false
		}
		*f = v
	}

	// Trailing unknown integers and an optional point precede the name string.
	for {
		tok, ok := p.peek()
		if !ok {
			return o, false
		}
		if tok == "OB" {
			// No name string present; the block closes immediately.
			p.next()
			return o, true
		}
		if isQuoted(tok) {
			o.name = unquote(tok)
			p.next()
			break
		}
		// An integer, a point [x.y], or any other scalar field: skip it.
		p.next()
	}

	if !p.expect("OB") {
		return o, false
	}
	return o, true
}

func (p *bcsParser) readInt() (int32, bool) {
	tok, ok := p.next()
	if !ok {
		return 0, false
	}
	v, err := strconv.ParseInt(tok, 0, 64)
	if err != nil {
		return 0, false
	}
	return int32(v), true
}

func (p *bcsParser) readString() (string, bool) {
	tok, ok := p.next()
	if !ok {
		return "", false
	}
	if !isQuoted(tok) {
		return "", false
	}
	return unquote(tok), true
}

func (p *bcsParser) expect(marker string) bool {
	tok, ok := p.next()
	return ok && tok == marker
}

func (p *bcsParser) skipToActionEnd() {
	for {
		tok, ok := p.next()
		if !ok || tok == "AC" {
			return
		}
	}
}

func isQuoted(tok string) bool {
	return len(tok) >= 2 && tok[0] == '"' && tok[len(tok)-1] == '"'
}

func unquote(tok string) string {
	if isQuoted(tok) {
		return tok[1 : len(tok)-1]
	}
	return tok
}

// Object decoding is pragmatic — a normal single-object action renders its
// o2 object (o1 is reserved for an ActionOverride target). The object slots
// are physically present in every compiled action, so an objectless
// signature must suppress the slot explicitly or its zero value would leak
// out as a phantom "Nothing" argument.
func (p *bcsParser) renderCall(sig *ActionSignature, a bcsAction, strref uint32) string {
	if !sig.HasObject() {
		return fmt.Sprintf("%s(%d)", sig.Name, strref)
	}
	return fmt.Sprintf("%s(%s,%d)", sig.Name, p.renderObject(a.o2), strref)
}

// A name string wins when present; otherwise an OBJECT.IDS identifier is
// resolved by value. Exotic EA/RACE/GENERAL/… specifications fall back to a
// compact placeholder rather than a full [EA.GENERAL.RACE…] expansion.
func (p *bcsParser) renderObject(o bcsObject) string {
	if o.name != "" {
		return o.name
	}
	if o.isPlainIdentifier() {
		if ident, ok := p.objectIds.Entries[o.identifiers]; ok {
			return ident
		}
	}
	return "[object]"
}

func (o bcsObject) isPlainIdentifier() bool {
	return o.ea == 0 && o.general == 0 && o.race == 0 && o.class == 0 &&
		o.specific == 0 && o.gender == 0 && o.alignment == 0
}
