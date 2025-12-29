// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type TwoDA struct {
	DefaultValue string
	Columns      []string
	RowKeys      []string
	rows         map[string][]string
}

func ParseTwoDA(r io.Reader) (*TwoDA, error) {
	t := &TwoDA{
		rows: make(map[string][]string),
	}

	scanner := bufio.NewScanner(r)
	lineNum := 0
	headersParsed := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if lineNum == 1 {
			if !isTwoDASignature(line) {
				return nil, fmt.Errorf("line %d: invalid 2DA signature: %s", lineNum, line)
			}
			continue
		}

		if t.DefaultValue == "" && !headersParsed {
			t.DefaultValue = line
			continue
		}

		if !headersParsed {
			t.Columns = strings.Fields(line)
			headersParsed = true
			continue
		}

		// Data rows
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		rowKey := fields[0]
		var values []string
		if len(fields) > 1 {
			values = fields[1:]
		}

		t.RowKeys = append(t.RowKeys, rowKey)
		t.rows[rowKey] = values
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

func isTwoDASignature(line string) bool {
	return strings.HasPrefix(line, "2DA")
}

func (t *TwoDA) Len() int {
	return len(t.RowKeys)
}

func (t *TwoDA) Row(rowKey string) ([]string, bool) {
	row, ok := t.rows[rowKey]
	return row, ok
}

func (t *TwoDA) ColumnIndex(name string) int {
	for i, col := range t.Columns {
		if col == name {
			return i
		}
	}
	return -1
}

func (t *TwoDA) Get(rowKey, column string) (string, bool) {
	colIdx := t.ColumnIndex(column)
	if colIdx == -1 {
		return "", false
	}
	return t.GetByIndex(rowKey, colIdx)
}

func (t *TwoDA) GetByIndex(rowKey string, colIndex int) (string, bool) {
	row, ok := t.rows[rowKey]
	if !ok {
		return "", false
	}
	if colIndex < 0 || colIndex >= len(row) {
		return "", false
	}
	return row[colIndex], true
}

func (t *TwoDA) GetOrDefault(rowKey, column string) string {
	val, ok := t.Get(rowKey, column)
	if !ok {
		return t.DefaultValue
	}
	return val
}

func (t *TwoDA) GetByIndexOrDefault(rowKey string, colIndex int) string {
	val, ok := t.GetByIndex(rowKey, colIndex)
	if !ok {
		return t.DefaultValue
	}
	return val
}
