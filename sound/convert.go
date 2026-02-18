// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"bytes"
	"fmt"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

// DecodeWavc decodes WAVC data to raw 16-bit PCM samples.
func DecodeWavc(data []byte) (pcm []byte, channels, sampleRate, bitsPerSample int, err error) {
	wavc := p.NewWavc()
	stream := kaitai.NewStream(bytes.NewReader(data))
	if err = wavc.Read(stream, nil, wavc); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("parsing WAVC header: %w", err)
	}

	acmData, err := wavc.AcmData()
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("reading WAVC ACM data: %w", err)
	}

	channels = int(wavc.NumChannels)
	sampleRate = int(wavc.SampleRate)
	bitsPerSample = int(wavc.BitsPerSample)

	pcm, err = decodeAcmData(acmData, channels, sampleRate)
	return pcm, channels, sampleRate, bitsPerSample, err
}

// DecodeAcm decodes raw ACM data to 16-bit PCM samples.
func DecodeAcm(data []byte) (pcm []byte, channels, sampleRate, bitsPerSample int, err error) {
	acm := p.NewAcm()
	stream := kaitai.NewStream(bytes.NewReader(data))
	if err = acm.Read(stream, nil, acm); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("parsing ACM header: %w", err)
	}

	channels = int(acm.NumChannels)
	sampleRate = int(acm.SampleRate)
	bitsPerSample = 16 // ACM always outputs 16-bit

	compressedData, err := acm.CompressedData()
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("reading ACM compressed data: %w", err)
	}

	pcm, err = decodeAcmData(compressedData, channels, sampleRate)
	return pcm, channels, sampleRate, bitsPerSample, err
}

func decodeAcmData(acmRawData []byte, channelsHint, sampleRateHint int) ([]byte, error) {
	// Parse ACM header from the raw data to get decoder params
	acm := p.NewAcm()
	stream := kaitai.NewStream(bytes.NewReader(acmRawData))
	if err := acm.Read(stream, nil, acm); err != nil {
		return nil, fmt.Errorf("parsing ACM data: %w", err)
	}

	levels, err := acm.Levels()
	if err != nil {
		return nil, fmt.Errorf("reading ACM levels: %w", err)
	}
	subBlocks, err := acm.SubBlocks()
	if err != nil {
		return nil, fmt.Errorf("reading ACM sub_blocks: %w", err)
	}

	numChannels := int(acm.NumChannels)
	sampleRate := int(acm.SampleRate)
	if channelsHint > 0 {
		numChannels = channelsHint
	}
	if sampleRateHint > 0 {
		sampleRate = sampleRateHint
	}

	numSamples := int(acm.NumSamples)

	// Compressed data starts after the 14-byte ACM header
	compressedData := acmRawData[14:]

	decoder, err := NewAcmDecoder(numSamples, numChannels, sampleRate, levels, subBlocks, compressedData)
	if err != nil {
		return nil, fmt.Errorf("creating ACM decoder: %w", err)
	}

	// Allocate output buffer: numSamples * 2 bytes (16-bit)
	pcm := make([]byte, numSamples*2)
	decoder.ReadSamples(pcm, numSamples)

	return pcm, nil
}
