// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"fmt"
	"strings"
)

// ActionParameter is one typed, named argument in an ACTION.IDS/TRIGGER.IDS
// identifier string, e.g. "O:Object*" or "I:StrRef*".
type ActionParameter struct {
	Type string
	Name string
}

// ActionSignature is the parsed form of an ACTION.IDS/TRIGGER.IDS identifier
// string such as "DisplayString(O:Object*,I:StrRef*)".
type ActionSignature struct {
	Name       string
	Parameters []ActionParameter

	// StrRefPositional is the index of the StrRef parameter among all
	// parameters, or -1 if the signature carries none.
	StrRefPositional int

	// StrRefIntSlot is the index of the StrRef parameter among the
	// integer-typed ("I") parameters only, or -1 if the signature carries
	// none. Parameters of other types do not advance this count, which
	// keeps games that reorder parameters (e.g. Planescape: Torment)
	// resolving to the correct slot in a compiled action block.
	StrRefIntSlot int
}

// HasStrRef reports whether the signature carries a StrRef parameter, i.e.
// whether it is a display-family action.
func (s *ActionSignature) HasStrRef() bool {
	return s.StrRefPositional >= 0
}

func (s *ActionSignature) HasObject() bool {
	for _, p := range s.Parameters {
		if p.Type == "O" {
			return true
		}
	}
	return false
}

// ParseActionSignature parses an ACTION.IDS/TRIGGER.IDS identifier string
// into its function name and ordered, typed parameter list.
func ParseActionSignature(identifier string) (*ActionSignature, error) {
	open := strings.IndexByte(identifier, '(')
	if open == -1 || !strings.HasSuffix(identifier, ")") {
		return nil, fmt.Errorf("invalid signature: %s", identifier)
	}

	sig := &ActionSignature{
		Name:             identifier[:open],
		StrRefPositional: -1,
		StrRefIntSlot:    -1,
	}

	paramsStr := strings.TrimSpace(identifier[open+1 : len(identifier)-1])
	if paramsStr == "" {
		return sig, nil
	}

	intSlot := 0
	for i, raw := range strings.Split(paramsStr, ",") {
		param, err := parseActionParameter(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", identifier, err)
		}
		sig.Parameters = append(sig.Parameters, param)

		isInt := param.Type == "I"
		if isInt && isStrRefParam(param.Name) {
			sig.StrRefPositional = i
			sig.StrRefIntSlot = intSlot
		}
		if isInt {
			intSlot++
		}
	}

	return sig, nil
}

// A StrRef is always an integer parameter, but ACTION.IDS does not name it
// consistently.
func isStrRefParam(name string) bool {
	return strings.EqualFold(name, "strref") ||
		strings.EqualFold(name, "stringref") ||
		strings.EqualFold(name, "entry")
}

func parseActionParameter(raw string) (ActionParameter, error) {
	raw = strings.TrimSpace(raw)

	typ, name, found := strings.Cut(raw, ":")
	if !found {
		return ActionParameter{}, fmt.Errorf("invalid parameter: %s", raw)
	}

	// The name ends at the first "*"; anything after it (e.g. the constraining
	// IDS file in "I:Value*Time") is not part of the name.
	name, _, _ = strings.Cut(name, "*")

	return ActionParameter{
		Type: typ,
		Name: name,
	}, nil
}

// DisplayFamilyActions maps action names to their parsed signature, for
// display-family actions (actions whose signature carries a StrRef parameter).
type DisplayFamilyActions map[string]*ActionSignature

// BuildDisplayFamilyActions resolves the display-family action signatures
// from a parsed ACTION.IDS. The set is derived from the game's own ACTION.IDS
// entries rather than hard-coded, so games that add or reorder actions
// (e.g. Planescape: Torment) resolve correctly.
func BuildDisplayFamilyActions(ids *Ids) (DisplayFamilyActions, error) {
	family := make(DisplayFamilyActions)

	for _, identifier := range ids.Entries {
		sig, err := ParseActionSignature(identifier)
		if err != nil {
			return nil, err
		}
		if sig.HasStrRef() {
			family[sig.Name] = sig
		}
	}

	return family, nil
}
