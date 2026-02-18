// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

const flacBlockSize = 4096

func channelsConst(n int) frame.Channels {
	if n == 1 {
		return frame.ChannelsMono
	}
	return frame.ChannelsLR
}

// WriteFlac writes raw PCM samples as a FLAC file.
func WriteFlac(w io.Writer, samples []byte, channels, sampleRate, bitsPerSample int) error {
	if channels < 1 || channels > 2 {
		return fmt.Errorf("unsupported number of channels: %d", channels)
	}
	if sampleRate < 4096 || sampleRate > 192000 {
		return fmt.Errorf("unsupported sample rate: %d", sampleRate)
	}

	bytesPerSample := bitsPerSample / 8
	totalSamples := len(samples) / (bytesPerSample * channels)

	info := &meta.StreamInfo{
		BlockSizeMin:  flacBlockSize,
		BlockSizeMax:  flacBlockSize,
		SampleRate:    uint32(sampleRate),
		NChannels:     uint8(channels),
		BitsPerSample: uint8(bitsPerSample),
		NSamples:      uint64(totalSamples),
	}

	enc, err := flac.NewEncoder(w, info)
	if err != nil {
		return fmt.Errorf("creating FLAC encoder: %w", err)
	}

	samplesWritten := 0
	for samplesWritten < totalSamples {
		nSamples := flacBlockSize
		if samplesWritten+nSamples > totalSamples {
			nSamples = totalSamples - samplesWritten
		}

		f := &frame.Frame{
			Header: frame.Header{
				HasFixedBlockSize: true,
				BlockSize:         uint16(nSamples),
				SampleRate:        uint32(sampleRate),
				Channels:          channelsConst(channels),
				BitsPerSample:     uint8(bitsPerSample),
			},
			Subframes: make([]*frame.Subframe, channels),
		}

		for ch := 0; ch < channels; ch++ {
			subSamples := make([]int32, nSamples)
			for s := 0; s < nSamples; s++ {
				idx := ((samplesWritten + s) * channels + ch) * bytesPerSample
				if idx+1 < len(samples) {
					subSamples[s] = int32(int16(binary.LittleEndian.Uint16(samples[idx:])))
				}
			}

			f.Subframes[ch] = &frame.Subframe{
				SubHeader: frame.SubHeader{
					Pred: frame.PredVerbatim,
				},
				NSamples: nSamples,
				Samples:  subSamples,
			}
		}

		if err := enc.WriteFrame(f); err != nil {
			return fmt.Errorf("writing FLAC frame: %w", err)
		}

		samplesWritten += nSamples
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("closing FLAC encoder: %w", err)
	}

	return nil
}
