// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
)

type TlkFile struct {
	*Tlk
	File afero.File
}

func (t *TlkFile) GetText(strref uint32) string {
	invalid_result := fmt.Sprintf("<invalid text reference #%d>", strref)

	if t == nil || t.Tlk == nil {
		log.Fatal("TLK file is not loaded")
		return invalid_result
	}

	if strref == 0xFFFFFFFF || strref > t.NumEntries {
		return invalid_result
	}

	text, err := t.Entries[strref].Text()
	if err != nil {
		log.Printf("Error retrieving TLK text for entry #%d: %v", strref, err)
		return invalid_result
	}

	return text
}

func (t *TlkFile) FileName() string {
	return t.File.Name()
}

func (t *TlkFile) Close() error {
	return t.File.Close()
}
