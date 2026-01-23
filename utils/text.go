// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package utils

import "strings"

// Splits "male <sep> female" text into male/female variants.
// Returns (male, female, hasSplit).
func SplitMaleFemaleText(text, separator string) (string, string, bool) {
	if separator == "" {
		return text, "", false
	}

	idx := strings.Index(text, separator)
	if idx == -1 {
		return text, "", false
	}

	male := strings.TrimSpace(text[:idx])
	female := strings.TrimSpace(text[idx+len(separator):])
	return male, female, true
}
