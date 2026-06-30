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

// ScriptCall is one decoded display-family action found in a compiled script
// (a standalone .BCS resource): the StrRef it references together with the
// rendered source-form call, e.g. {14381, "DisplayString(LastTrigger,14381)"}.
type ScriptCall struct {
	StrRef uint32
	Call   string
}

// bcsObject is one parsed OB object specification from a compiled script. Only
// the fields needed to render a pragmatic decoded call are retained.
type bcsObject struct {
	// ea..alignment are the enumerated specification fields (EA.IDS,
	// GENERAL.IDS, …). For a plain object reference they are all zero.
	ea, general, race, class, specific, gender, alignment int32
	// identifiers is the OBJECT.IDS value (e.g. LastTrigger, Myself).
	identifiers int32
	// name is the object name string, shown verbatim when present.
	name string
}

// bcsAction is one parsed AC action block: its numeric ACTION.IDS id, the
// three object slots (o1 reserved for an ActionOverride target, o2 the normal
// single-object slot, o3 a secondary object) and the integer slots in their
// physical layout int0, point(x,y), int1, int2.
type bcsAction struct {
	id   int32
	o1   bcsObject
	o2   bcsObject
	o3   bcsObject
	ints [5]int32 // int0, x, y, int1, int2
}

// intSlotField maps a logical integer slot (ActionSignature.StrRefIntSlot) to
// the physical field index within bcsAction.ints. The point coordinates
// occupy physical fields 1 and 2, so the logical integer slots land on 0, 3
// and 4.
var intSlotField = map[int]int{0: 0, 1: 3, 2: 4}

// DecodeCompiledScript parses the textual source of a compiled script (.BCS)
// and returns one ScriptCall for every display-family action it contains.
//
// actionIds is the parsed ACTION.IDS, used to resolve a block's numeric action
// id to its name; family selects the display-family actions and supplies the
// integer slot that carries the StrRef; objectIds is the parsed OBJECT.IDS,
// used to resolve object identifiers such as LastTrigger or Myself.
//
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

// bcsParser walks a flat token stream of a compiled script. Markers and the
// numeric/string fields between them are all individual tokens.
type bcsParser struct {
	tokens []string
	pos    int

	actionIds *Ids
	family    DisplayFamilyActions
	objectIds *Ids

	calls []ScriptCall
}

// tokenizeBcs splits the BCS source into a flat token stream. Two-letter block
// markers are emitted as their own tokens even when a numeric id is glued to
// the following marker (e.g. "100AC" → "100", "AC"); quoted strings are kept
// as single tokens (with the surrounding quotes) so empty strings and embedded
// spaces survive.
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
			// Consume a quoted string as a single token.
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
			// Consume a run of non-space, non-quote characters, then split off
			// any trailing two-letter block marker glued to a number
			// (e.g. "100AC" → "100", "AC").
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

// splitGluedMarker breaks a run like "100AC" into "100", "AC" while leaving a
// bare marker or a bare number untouched. It splits off at most one trailing
// marker (the format never glues more than one).
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

// parseScript walks the whole token stream. The block structure is shallow:
// the only thing that contributes script context is an AC action block. An
// action block opens with its numeric id glued to the AC marker (e.g.
// "100AC"), so the parser scans for an integer token immediately followed by
// an AC marker and delegates that block, skipping everything else
// (SC/CR/CO/TR/RS/RE markers, trigger fields, response weights).
func (p *bcsParser) parseScript() []ScriptCall {
	for {
		tok, ok := p.next()
		if !ok {
			break
		}
		// The action id precedes the AC marker. When an integer is followed
		// by AC, it opens an action block.
		if isAllDigitsOrSign(tok) {
			if marker, ok := p.peek(); ok && marker == "AC" {
				p.next() // consume the opening AC marker
				p.parseActionBlock(tok)
			}
		}
	}
	return p.calls
}

// parseActionBlock decodes the body of an action block whose numeric id and
// opening AC marker have already been consumed. A malformed block is consumed
// (up to its closing AC, or end of stream) and skipped so the rest of the
// script still parses.
func (p *bcsParser) parseActionBlock(idTok string) {
	action, ok := p.readAction(idTok)
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
		Call:   p.renderCall(sig.Name, action, strref),
	})
}

// readAction parses the body of an action block whose numeric id (idTok) and
// opening AC marker have already been consumed: three OB objects, then the five
// int/point fields and two strings, up to the closing AC marker. It returns
// ok=false if the block is structurally broken, having consumed tokens up to
// the next AC.
func (p *bcsParser) readAction(idTok string) (bcsAction, bool) {
	var a bcsAction

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

// readObjectBody parses the fields of an OB block (the opening OB marker has
// already been consumed) up to and including the closing OB marker: eight
// enumerated/identifier integers, then up to four trailing integers, an
// optional point and a name string. The trailing fields vary by game, so the
// reader is tolerant: it consumes integers and an optional [x.y] point until
// it reaches the name string, then the closing OB marker.
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

// skipToActionEnd consumes tokens until the closing AC marker (or end of
// stream) so a malformed action block does not derail the rest of the parse.
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

// renderCall builds the source-form call string for a display-family action:
// the action name with its decoded object and the StrRef. Object decoding is
// pragmatic — a normal single-object action renders its o2 object (o1 is
// reserved for an ActionOverride target).
func (p *bcsParser) renderCall(name string, a bcsAction, strref uint32) string {
	return fmt.Sprintf("%s(%s,%d)", name, p.renderObject(a.o2), strref)
}

// renderObject decodes an object specification into a compact source-form
// token. A name string wins when present; otherwise an OBJECT.IDS identifier
// is resolved by value. Exotic EA/RACE/GENERAL/… specifications fall back to a
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

// isPlainIdentifier reports whether the object is a simple OBJECT.IDS reference
// with no enumerated EA/RACE/GENERAL/… specification.
func (o bcsObject) isPlainIdentifier() bool {
	return o.ea == 0 && o.general == 0 && o.race == 0 && o.class == 0 &&
		o.specific == 0 && o.gender == 0 && o.alignment == 0
}
