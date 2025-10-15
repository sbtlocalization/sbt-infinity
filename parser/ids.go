// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
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

type Ids struct {
	FileIdentifier string
	NumEntries     int
	Entries        map[int32]string
}

func ParseIds(r io.Reader) (*Ids, error) {
	ids := &Ids{
		Entries: make(map[int32]string),
	}

	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if lineNum == 1 && isHeaderIdentifier(line) {
			ids.FileIdentifier = line
			continue
		}

		if lineNum == 1 || (lineNum == 2 && ids.FileIdentifier != "") {
			if num, err := strconv.Atoi(line); err == nil {
				ids.NumEntries = num
				continue
			}
		}

		value, identifier, err := parseEntry(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		ids.Entries[int32(value)] = identifier
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func isHeaderIdentifier(line string) bool {
	return strings.HasPrefix(line, "IDS")
}

func parseEntry(line string) (value int64, identifier string, err error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid entry format: %s", line)
	}

	valueStr := parts[0]
	identifier = strings.Join(parts[1:], " ")

	value, err = strconv.ParseInt(valueStr, 0, 32)
	if err != nil {
		return 0, "", fmt.Errorf("invalid value: %s", valueStr)
	}

	return value, identifier, nil
}
