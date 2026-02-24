// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dcanvas

import (
	"encoding/json"
	"fmt"
	"io"
)

// Encode writes a Canvas as indented JSON to w.
func Encode(c *Canvas, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("can't encode dcanvas: %w", err)
	}
	return nil
}

// Decode reads a Canvas from JSON in r.
func Decode(r io.Reader) (*Canvas, error) {
	var c Canvas
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return nil, fmt.Errorf("can't decode dcanvas: %w", err)
	}
	return &c, nil
}
