// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"fmt"
	"log"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/spf13/afero"
)

type TlkFile struct {
	*Tlk
	File afero.File
}

func ReadTlkFile(fs afero.Fs, fileName string) (*TlkFile, error) {
	file, err := fs.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open TLK file %s: %w", fileName, err)
	}

	tlk := NewTlk()
	stream := kaitai.NewStream(file)
	err = tlk.Read(stream, nil, tlk)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLK file %s: %w", fileName, err)
	}

	tlkFile := &TlkFile{
		Tlk:  tlk,
		File: file,
	}
	return tlkFile, nil
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

func (t *TlkFile) GetSound(strref uint32) string {
	if t == nil || t.Tlk == nil || strref == 0xFFFFFFFF || strref > t.NumEntries {
		return ""
	}

	entry := t.Entries[strref]
	if !entry.Flags.SoundExists {
		return ""
	}

	return entry.AudioName
}

func (t *TlkFile) FileName() string {
	return t.File.Name()
}

func (t *TlkFile) Close() error {
	return t.File.Close()
}
