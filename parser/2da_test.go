// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"os"
	"strings"
	"testing"
)

func TestParseTwoDA_Basic(t *testing.T) {
	input := `2DA V1.0
1234
        NAME    VALUE   WEIGHT
A       alpha
B
C       beta    2345    123
`

	twoda, err := ParseTwoDA(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "1234" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "1234")
	}

	if len(twoda.Columns) != 3 {
		t.Errorf("len(Columns) = %d, want 3", len(twoda.Columns))
	}

	if twoda.Len() != 3 {
		t.Errorf("Len() = %d, want 3", twoda.Len())
	}

	// Test Get with existing value
	val, ok := twoda.Get("A", "NAME")
	if !ok || val != "alpha" {
		t.Errorf("Get(A, NAME) = (%q, %v), want (alpha, true)", val, ok)
	}

	// Test Get with missing value
	val, ok = twoda.Get("A", "VALUE")
	if ok {
		t.Errorf("Get(A, VALUE) = (%q, %v), want ('', false)", val, ok)
	}

	// Test GetOrDefault with missing value
	val = twoda.GetOrDefault("A", "VALUE")
	if val != "1234" {
		t.Errorf("GetOrDefault(A, VALUE) = %q, want %q", val, "1234")
	}

	// Test Get with all values present
	val, ok = twoda.Get("C", "WEIGHT")
	if !ok || val != "123" {
		t.Errorf("Get(C, WEIGHT) = (%q, %v), want (123, true)", val, ok)
	}

	// Test Row
	row, ok := twoda.Row("C")
	if !ok {
		t.Error("Row(C) not found")
	}
	if len(row) != 3 || row[0] != "beta" || row[1] != "2345" || row[2] != "123" {
		t.Errorf("Row(C) = %v, want [beta 2345 123]", row)
	}

	// Test empty row
	row, ok = twoda.Row("B")
	if !ok {
		t.Error("Row(B) not found")
	}
	if len(row) != 0 {
		t.Errorf("Row(B) = %v, want []", row)
	}
}

func TestParseTwoDA_InvalidSignature(t *testing.T) {
	input := `INVALID
default
COL1
ROW1    val1
`

	_, err := ParseTwoDA(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for invalid signature, got nil")
	}
}

func TestParseTwoDA_ColumnIndex(t *testing.T) {
	input := `2DA V1.0
0
        COL1    COL2    COL3
ROW1    a       b       c
`

	twoda, err := ParseTwoDA(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if idx := twoda.ColumnIndex("COL1"); idx != 0 {
		t.Errorf("ColumnIndex(COL1) = %d, want 0", idx)
	}
	if idx := twoda.ColumnIndex("COL3"); idx != 2 {
		t.Errorf("ColumnIndex(COL3) = %d, want 2", idx)
	}
	if idx := twoda.ColumnIndex("NOTFOUND"); idx != -1 {
		t.Errorf("ColumnIndex(NOTFOUND) = %d, want -1", idx)
	}
}

func TestParseTwoDA_EFFTEXT(t *testing.T) {
	file, err := os.Open("../iwd/2DA/EFFTEXT.2DA")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer file.Close()

	twoda, err := ParseTwoDA(file)
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "-1" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "-1")
	}
	if len(twoda.Columns) != 2 {
		t.Errorf("len(Columns) = %d, want 2", len(twoda.Columns))
	}
	if twoda.Columns[0] != "EFFECT_NAME" || twoda.Columns[1] != "STRREF" {
		t.Errorf("Columns = %v, want [EFFECT_NAME STRREF]", twoda.Columns)
	}

	// Row "25" should have POISON, 35501
	val, ok := twoda.Get("25", "EFFECT_NAME")
	if !ok || val != "POISON" {
		t.Errorf("Get(25, EFFECT_NAME) = (%q, %v), want (POISON, true)", val, ok)
	}
	val, ok = twoda.Get("25", "STRREF")
	if !ok || val != "35501" {
		t.Errorf("Get(25, STRREF) = (%q, %v), want (35501, true)", val, ok)
	}
}

func TestParseTwoDA_MSCHOOL(t *testing.T) {
	file, err := os.Open("../iwd/2DA/MSCHOOL.2DA")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer file.Close()

	twoda, err := ParseTwoDA(file)
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "4294967296" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "4294967296")
	}
	if twoda.Columns[0] != "RES_REF" {
		t.Errorf("Columns[0] = %q, want RES_REF", twoda.Columns[0])
	}

	// Check some rows
	val, ok := twoda.Get("ABJURER", "RES_REF")
	if !ok || val != "37451" {
		t.Errorf("Get(ABJURER, RES_REF) = (%q, %v), want (37451, true)", val, ok)
	}

	// None row should use default (has 4294967296)
	val, ok = twoda.Get("None", "RES_REF")
	if !ok || val != "4294967296" {
		t.Errorf("Get(None, RES_REF) = (%q, %v), want (4294967296, true)", val, ok)
	}
}

func TestParseTwoDA_TRACKING(t *testing.T) {
	file, err := os.Open("../iwd/2DA/TRACKING.2DA")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer file.Close()

	twoda, err := ParseTwoDA(file)
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "O_19534" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "O_19534")
	}

	// Check AR1000 row
	val, ok := twoda.Get("AR1000", "STRREF")
	if !ok || val != "O_24826" {
		t.Errorf("Get(AR1000, STRREF) = (%q, %v), want (O_24826, true)", val, ok)
	}

	// Check row count
	if twoda.Len() < 100 {
		t.Errorf("Len() = %d, expected at least 100 rows", twoda.Len())
	}
}

func TestParseTwoDA_ENGINEST(t *testing.T) {
	file, err := os.Open("../iwd/2DA/ENGINEST.2DA")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer file.Close()

	twoda, err := ParseTwoDA(file)
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "0" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "0")
	}

	// Check first row
	val, ok := twoda.Get("STRREF_ATTENTION_DIALOG", "StrRef")
	if !ok || val != "10467" {
		t.Errorf("Get(STRREF_ATTENTION_DIALOG, StrRef) = (%q, %v), want (10467, true)", val, ok)
	}

	// Large file - should have many rows
	if twoda.Len() < 1000 {
		t.Errorf("Len() = %d, expected at least 1000 rows", twoda.Len())
	}
}

func TestParseTwoDA_MSECTYPE(t *testing.T) {
	file, err := os.Open("../iwd/2DA/MSECTYPE.2DA")
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer file.Close()

	twoda, err := ParseTwoDA(file)
	if err != nil {
		t.Fatalf("ParseTwoDA failed: %v", err)
	}

	if twoda.DefaultValue != "4294967296" {
		t.Errorf("DefaultValue = %q, want %q", twoda.DefaultValue, "4294967296")
	}

	val, ok := twoda.Get("SpellProtections", "RES_REF")
	if !ok || val != "37461" {
		t.Errorf("Get(SpellProtections, RES_REF) = (%q, %v), want (37461, true)", val, ok)
	}
}
