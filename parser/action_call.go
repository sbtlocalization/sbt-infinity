// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import "strings"

// ActionCall is one parsed source-form action statement from a dialog
// transition's action text, e.g. "DisplayStringHead(Myself,14381)".
type ActionCall struct {
	Name string
	Args []string
}

// ParseActionCalls splits a block of source-form action text into its
// individual statements (one or more "Name(args)" calls), returning the
// function name and raw, top-level argument tokens of each. Nested parens
// (object specs) and quoted strings (which may contain commas) are respected
// when splitting both statements and arguments; arguments are returned as raw
// strings without further decoding.
func ParseActionCalls(text string) []ActionCall {
	var calls []ActionCall

	for _, stmt := range splitStatements(text) {
		if call, ok := parseActionCall(stmt); ok {
			calls = append(calls, call)
		}
	}

	return calls
}

// splitStatements breaks the action text into individual "Name(...)"
// statements. Statements are conventionally one per line, but a depth- and
// quote-aware scan also handles several glued on a single line.
func splitStatements(text string) []string {
	var statements []string
	depth := 0
	inQuote := false
	start := 0

	for i := 0; i < len(text); i++ {
		ch := text[i]
		switch {
		case ch == '"':
			inQuote = !inQuote
		case inQuote:
			// skip everything inside a quoted string
		case ch == '(':
			depth++
		case ch == ')':
			if depth > 0 {
				depth--
			}
			if depth == 0 {
				statements = append(statements, text[start:i+1])
				start = i + 1
			}
		}
	}

	return statements
}

func parseActionCall(stmt string) (ActionCall, bool) {
	stmt = strings.TrimSpace(stmt)

	open := strings.IndexByte(stmt, '(')
	if open == -1 || !strings.HasSuffix(stmt, ")") {
		return ActionCall{}, false
	}

	name := strings.TrimSpace(stmt[:open])
	if name == "" {
		return ActionCall{}, false
	}

	argsStr := stmt[open+1 : len(stmt)-1]

	return ActionCall{Name: name, Args: splitArgs(argsStr)}, true
}

// splitArgs splits an argument list at the top level only, respecting nested
// parens and quoted strings so that object specs and quoted commas stay intact.
func splitArgs(argsStr string) []string {
	if strings.TrimSpace(argsStr) == "" {
		return nil
	}

	var args []string
	depth := 0
	inQuote := false
	start := 0

	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]
		switch {
		case ch == '"':
			inQuote = !inQuote
		case inQuote:
			// skip everything inside a quoted string
		case ch == '(':
			depth++
		case ch == ')':
			if depth > 0 {
				depth--
			}
		case ch == ',' && depth == 0:
			args = append(args, strings.TrimSpace(argsStr[start:i]))
			start = i + 1
		}
	}

	args = append(args, strings.TrimSpace(argsStr[start:]))

	return args
}
