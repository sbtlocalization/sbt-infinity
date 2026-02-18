// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"log"
	"strings"

	"github.com/gobwas/glob"
)

type CompiledFilter struct {
	filter glob.Glob
	caseSensitive bool
	removePrefix bool //Remove `data/` prefix from input string if pattern has no slashes
	removeExtension bool //Remove `.bif` suffix from input string if pattern has no dots
}

func CompileFilter (pattern string, caseSensitive bool, removePrefix bool, removeExtension bool) *CompiledFilter {
	if len(pattern) == 0 {
		// Empty filter is not an error
		return nil
	}
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
	}
	if removePrefix && strings.ContainsAny(pattern, "/\\") {
		removePrefix = false
	}
	if removeExtension && strings.ContainsAny(pattern, ".") {
		removeExtension = false
	}
	filter, err := glob.Compile(pattern)
	if err != nil {
		log.Println("Error compiling filter:", err)
		return nil
	}
	return &CompiledFilter {
		filter: filter,
		caseSensitive: caseSensitive,
		removePrefix: removePrefix,
		removeExtension: removeExtension,
	}
}

func (f *CompiledFilter) Match(input string) bool {
	if f.removePrefix {
		input, _ = strings.CutPrefix(input, "data/")
	}
	bifExtension := ".bif"
	if f.removeExtension && len(input) > len(bifExtension) {
		extension := strings.ToLower(input[len(input) - len(bifExtension):])
		if strings.HasSuffix(extension, bifExtension) {
			input = input[:len(input) - len(bifExtension)]
		}
	}
	if !f.caseSensitive {
		input = strings.ToLower(input)
	}
	return f.filter.Match(input)
}
