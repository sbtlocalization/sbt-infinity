// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"encoding/binary"
	"fmt"
	"io"
)

// IsRIFF returns true if data starts with a RIFF WAV header.
func IsRIFF(data []byte) bool {
	return len(data) >= 12 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WAVE"
}

// IsWAVC returns true if data starts with a WAVC header.
func IsWAVC(data []byte) bool {
	return len(data) >= 8 &&
		string(data[0:4]) == "WAVC" &&
		string(data[4:8]) == "V1.0"
}

// IsACM returns true if data starts with the ACM signature.
func IsACM(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	sig := binary.LittleEndian.Uint32(data[0:4])
	return sig == 0x01032897
}

// DecodeRiff parses a standard RIFF WAV and returns raw PCM samples
// along with audio parameters.
func DecodeRiff(data []byte) (pcm []byte, channels, sampleRate, bitsPerSample int, err error) {
	if !IsRIFF(data) {
		return nil, 0, 0, 0, fmt.Errorf("not a RIFF WAV file")
	}
	if len(data) < 44 {
		return nil, 0, 0, 0, fmt.Errorf("RIFF WAV too short: %d bytes", len(data))
	}

	channels = int(binary.LittleEndian.Uint16(data[22:24]))
	sampleRate = int(binary.LittleEndian.Uint32(data[24:28]))
	bitsPerSample = int(binary.LittleEndian.Uint16(data[34:36]))

	// Find the "data" chunk by scanning from byte 12 onward.
	offset := 12
	for offset+8 <= len(data) {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if chunkID == "data" {
			start := offset + 8
			end := start + chunkSize
			if end > len(data) {
				end = len(data)
			}
			return data[start:end], channels, sampleRate, bitsPerSample, nil
		}
		offset += 8 + chunkSize
		// RIFF chunks are word-aligned
		if chunkSize%2 != 0 {
			offset++
		}
	}

	return nil, 0, 0, 0, fmt.Errorf("no data chunk found in RIFF WAV")
}

// WriteWav writes raw PCM samples as a standard RIFF WAV file.
func WriteWav(w io.Writer, samples []byte, channels, sampleRate, bitsPerSample int) error {
	if channels < 1 || channels > 2 {
		return fmt.Errorf("unsupported number of channels: %d", channels)
	}
	if sampleRate < 4096 || sampleRate > 192000 {
		return fmt.Errorf("unsupported sample rate: %d", sampleRate)
	}

	blockAlign := channels * bitsPerSample / 8
	totalSize := len(samples)
	byteRate := sampleRate * blockAlign
	chunkSize := 36 + totalSize

	header := make([]byte, 44)
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(chunkSize))
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16) // SubChunk1 size
	binary.LittleEndian.PutUint16(header[20:22], 1)  // PCM format
	binary.LittleEndian.PutUint16(header[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(totalSize))

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("writing WAV header: %w", err)
	}
	if _, err := w.Write(samples); err != nil {
		return fmt.Errorf("writing WAV data: %w", err)
	}
	return nil
}
