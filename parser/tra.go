// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type TraEntry struct {
	ID         uint32
	MaleText   string
	FemaleText string // empty if no female variant
	SoundFile  string // optional
}

type TraFile struct {
	Entries []TraEntry
}

func ParseTra(r io.Reader) (*TraFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	return parseTra(string(data))
}

func ParseTraFile(path string) (*TraFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseTra(file)
}

func parseTra(content string) (*TraFile, error) {
	tf := &TraFile{}
	pos := 0

	for pos < len(content) {
		pos = skipWhitespaceAndComments(content, pos)
		if pos >= len(content) {
			break
		}

		// Expect @ or !
		if content[pos] != '@' && content[pos] != '!' {
			return nil, fmt.Errorf("expected '@' or '!' at position %d, got '%c'", pos, content[pos])
		}

		pos++

		// Parse ID
		idStart := pos
		for pos < len(content) && unicode.IsDigit(rune(content[pos])) {
			pos++
		}
		if idStart == pos {
			return nil, fmt.Errorf("expected numeric ID at position %d", idStart)
		}

		idVal, err := strconv.ParseUint(content[idStart:pos], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid ID at position %d: %w", idStart, err)
		}

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Expect =
		if pos >= len(content) || content[pos] != '=' {
			return nil, fmt.Errorf("expected '=' after ID at position %d", pos)
		}
		pos++

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Parse first text expression (male text)
		maleText, newPos, err := parseTextExpression(content, pos)
		if err != nil {
			return nil, fmt.Errorf("failed to parse male text at position %d: %w", pos, err)
		}
		pos = newPos

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		var femaleText string
		var soundFile string

		// Check for optional female text or sound file
		if pos < len(content) {
			// Check if next char starts a text expression (female text)
			if isDelimiterStart(content, pos) {
				femaleText, newPos, err = parseTextExpression(content, pos)
				if err != nil {
					return nil, fmt.Errorf("failed to parse female text at position %d: %w", pos, err)
				}
				pos = newPos

				// Skip whitespace
				for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
					pos++
				}
			}

			// Check for sound file
			if pos < len(content) && content[pos] == '[' {
				soundFile, newPos, err = parseSoundFile(content, pos)
				if err != nil {
					return nil, fmt.Errorf("failed to parse sound file at position %d: %w", pos, err)
				}
				pos = newPos
			}
		}

		tf.Entries = append(tf.Entries, TraEntry{
			ID:         uint32(idVal),
			MaleText:   maleText,
			FemaleText: femaleText,
			SoundFile:  soundFile,
		})
	}

	return tf, nil
}

// Skips whitespace and C++ style comments.
// Returns the new position after skipping.
func skipWhitespaceAndComments(content string, pos int) int {
	for pos < len(content) {
		// Skip whitespace
		if unicode.IsSpace(rune(content[pos])) {
			pos++
			continue
		}

		// Check for single-line comment
		if pos+1 < len(content) && content[pos] == '/' && content[pos+1] == '/' {
			pos += 2
			for pos < len(content) && content[pos] != '\n' {
				pos++
			}
			continue
		}

		// Check for multi-line comment
		if pos+1 < len(content) && content[pos] == '/' && content[pos+1] == '*' {
			pos += 2
			for pos+1 < len(content) && !(content[pos] == '*' && content[pos+1] == '/') {
				pos++
			}
			if pos+1 < len(content) {
				pos += 2 // Skip */
			}
			continue
		}

		break
	}
	return pos
}

// Checks if position starts a text delimiter.
func isDelimiterStart(content string, pos int) bool {
	if pos >= len(content) {
		return false
	}
	ch := content[pos]
	return ch == '~' || ch == '%' || ch == '"'
}

// Parses a text expression which may include concatenation.
// Returns the parsed text, new position, and any error.
func parseTextExpression(content string, pos int) (string, int, error) {
	var result strings.Builder

	for {
		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		if pos >= len(content) {
			break
		}

		// Parse a delimited string
		text, newPos, err := parseDelimitedString(content, pos)
		if err != nil {
			return "", pos, err
		}
		result.WriteString(text)
		pos = newPos

		// Skip whitespace
		for pos < len(content) && unicode.IsSpace(rune(content[pos])) {
			pos++
		}

		// Check for concatenation operator
		if pos < len(content) && content[pos] == '^' {
			pos++
			continue
		}

		break
	}

	return result.String(), pos, nil
}

// Parses a string with delimiters (~, %, ", or ~~~~~).
func parseDelimitedString(content string, pos int) (string, int, error) {
	if pos >= len(content) {
		return "", pos, fmt.Errorf("unexpected end of content")
	}

	// Check for ~~~~~ delimiter (5 tildes)
	if pos+4 < len(content) && content[pos:pos+5] == "~~~~~" {
		return parseWithDelimiter(content, pos+5, "~~~~~")
	}

	ch := content[pos]
	switch ch {
	case '~':
		return parseWithDelimiter(content, pos+1, "~")
	case '%':
		return parseWithDelimiter(content, pos+1, "%")
	case '"':
		return parseWithDelimiter(content, pos+1, "\"")
	default:
		return "", pos, fmt.Errorf("expected string delimiter at position %d, got '%c'", pos, ch)
	}
}

// Extracts text until the closing delimiter.
func parseWithDelimiter(content string, pos int, delimiter string) (string, int, error) {
	start := pos
	delimLen := len(delimiter)

	for pos <= len(content)-delimLen {
		if content[pos:pos+delimLen] == delimiter {
			return content[start:pos], pos + delimLen, nil
		}
		pos++
	}

	return "", start, fmt.Errorf("unclosed string starting at position %d", start-delimLen)
}

// Parses a sound file reference like [SOUNDNAME].
func parseSoundFile(content string, pos int) (string, int, error) {
	if pos >= len(content) || content[pos] != '[' {
		return "", pos, fmt.Errorf("expected '[' at position %d", pos)
	}
	pos++

	start := pos
	for pos < len(content) && content[pos] != ']' {
		pos++
	}

	if pos >= len(content) {
		return "", start, fmt.Errorf("unclosed sound file reference starting at position %d", start-1)
	}

	soundFile := content[start:pos]
	pos++ // Skip ]

	return soundFile, pos, nil
}

// GetByID returns the entry with the given ID, or nil if not found.
func (tf *TraFile) GetByID(id uint32) *TraEntry {
	for i := range tf.Entries {
		if tf.Entries[i].ID == id {
			return &tf.Entries[i]
		}
	}
	return nil
}

// ToMap converts entries to a map keyed by ID.
func (tf *TraFile) ToMap() map[uint32]*TraEntry {
	m := make(map[uint32]*TraEntry, len(tf.Entries))
	for i := range tf.Entries {
		m[tf.Entries[i].ID] = &tf.Entries[i]
	}
	return m
}
