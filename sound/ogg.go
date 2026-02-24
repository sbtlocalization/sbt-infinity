// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/jfreymuth/oggvorbis"
)

// IsOgg returns true if data starts with the OggS capture pattern.
func IsOgg(data []byte) bool {
	return len(data) >= 4 && string(data[0:4]) == "OggS"
}

// DecodeOgg decodes Ogg Vorbis data to raw 16-bit LE PCM samples.
func DecodeOgg(data []byte) (pcm []byte, channels, sampleRate, bitsPerSample int, err error) {
	if !IsOgg(data) {
		return nil, 0, 0, 0, fmt.Errorf("not an Ogg Vorbis file")
	}

	samples, format, decErr := decodeOggSafe(data)
	if decErr != nil {
		return nil, 0, 0, 0, fmt.Errorf("decoding Ogg Vorbis: %w", decErr)
	}

	channels = format.Channels
	sampleRate = format.SampleRate
	bitsPerSample = 16 // Ogg Vorbis decoded to 16-bit PCM

	// Convert interleaved float32 samples [-1.0, 1.0] to interleaved int16 LE bytes.
	pcm = make([]byte, len(samples)*2)
	for i, s := range samples {
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(int16(math.Round(float64(s)*32767.0))))
	}

	return pcm, channels, sampleRate, bitsPerSample, nil
}

// decodeOggSafe wraps oggvorbis.ReadAll with panic recovery,
// since the library panics on malformed streams.
func decodeOggSafe(data []byte) (samples []float32, format *oggvorbis.Format, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("corrupt Ogg stream: %v", r)
		}
	}()
	samples, format, err = oggvorbis.ReadAll(bytes.NewReader(data))
	return
}
