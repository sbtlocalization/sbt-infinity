// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	tlkHeaderSize = 18
	tlkEntrySize  = 26
)

type TlkWriteEntry struct {
	Text           string
	HasText        bool
	HasSound       bool
	HasToken       bool
	AudioName      string // max 8 chars, will be null-padded/truncated
	VolumeVariance uint32
	PitchVariance  uint32
}

type TlkWriteOptions struct {
	Lang uint16
}

func NewEmptyTlkEntry() TlkWriteEntry {
	return TlkWriteEntry{
		Text:           "",
		HasText:        false,
		HasSound:       false,
		HasToken:       false,
		AudioName:      "",
		VolumeVariance: 0,
		PitchVariance:  0,
	}
}

func WriteTlkFile(path string, entries []TlkWriteEntry, opts TlkWriteOptions) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create TLK file: %w", err)
	}
	defer file.Close()

	return WriteTlk(file, entries, opts)
}

func WriteTlk(w io.Writer, entries []TlkWriteEntry, opts TlkWriteOptions) error {
	numEntries := uint32(len(entries))
	ofsData := uint32(tlkHeaderSize + tlkEntrySize*int(numEntries))

	// Build string data with deduplication
	stringOffsets := make(map[string]uint32)
	stringData := bytes.Buffer{}

	type entryOffsets struct {
		ofsString uint32
		lenString uint32
	}
	offsets := make([]entryOffsets, numEntries)

	for i, entry := range entries {
		text := entry.Text
		textLen := uint32(len(text))

		if textLen == 0 {
			// Empty string: offset 0, length 0
			offsets[i] = entryOffsets{ofsString: 0, lenString: 0}
			continue
		}

		if existingOffset, exists := stringOffsets[text]; exists {
			// Reuse existing string offset
			offsets[i] = entryOffsets{ofsString: existingOffset, lenString: textLen}
		} else {
			// Add new string
			offset := uint32(stringData.Len())
			stringOffsets[text] = offset
			stringData.WriteString(text)
			offsets[i] = entryOffsets{ofsString: offset, lenString: textLen}
		}
	}

	if err := writeHeader(w, opts.Lang, numEntries, ofsData); err != nil {
		return err
	}

	for i, entry := range entries {
		if err := writeEntry(w, entry, offsets[i].ofsString, offsets[i].lenString); err != nil {
			return fmt.Errorf("failed to write entry %d: %w", i, err)
		}
	}

	if _, err := w.Write(stringData.Bytes()); err != nil {
		return fmt.Errorf("failed to write string data: %w", err)
	}

	return nil
}

func writeHeader(w io.Writer, lang uint16, numEntries, ofsData uint32) error {
	// Magic: "TLK "
	if _, err := w.Write([]byte("TLK ")); err != nil {
		return fmt.Errorf("failed to write magic: %w", err)
	}

	// Version: "V1  "
	if _, err := w.Write([]byte("V1  ")); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	// Lang (uint16 LE)
	if err := binary.Write(w, binary.LittleEndian, lang); err != nil {
		return fmt.Errorf("failed to write lang: %w", err)
	}

	// NumEntries (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, numEntries); err != nil {
		return fmt.Errorf("failed to write num_entries: %w", err)
	}

	// OfsData (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, ofsData); err != nil {
		return fmt.Errorf("failed to write ofs_data: %w", err)
	}

	return nil
}

func writeEntry(w io.Writer, entry TlkWriteEntry, ofsString, lenString uint32) error {
	// Flags (uint16 LE): bit 0 = TextExists, bit 1 = SoundExists, bit 2 = TokenExists
	var flags uint16
	if entry.HasText {
		flags |= 1 << 0
	}
	if entry.HasSound {
		flags |= 1 << 1
	}
	if entry.HasToken {
		flags |= 1 << 2
	}
	if err := binary.Write(w, binary.LittleEndian, flags); err != nil {
		return fmt.Errorf("failed to write flags: %w", err)
	}

	// AudioName (8 bytes, null-padded)
	audioName := padAudioName(entry.AudioName)
	if _, err := w.Write(audioName[:]); err != nil {
		return fmt.Errorf("failed to write audio_name: %w", err)
	}

	// VolumeVariance (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, entry.VolumeVariance); err != nil {
		return fmt.Errorf("failed to write volume_variance: %w", err)
	}

	// PitchVariance (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, entry.PitchVariance); err != nil {
		return fmt.Errorf("failed to write pitch_variance: %w", err)
	}

	// OfsString (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, ofsString); err != nil {
		return fmt.Errorf("failed to write ofs_string: %w", err)
	}

	// LenString (uint32 LE)
	if err := binary.Write(w, binary.LittleEndian, lenString); err != nil {
		return fmt.Errorf("failed to write len_string: %w", err)
	}

	return nil
}

// padAudioName pads or truncates the audio name to exactly 8 bytes.
func padAudioName(name string) [8]byte {
	var result [8]byte
	nameBytes := []byte(name)
	if len(nameBytes) > 8 {
		nameBytes = nameBytes[:8]
	}
	copy(result[:], nameBytes)
	return result
}
