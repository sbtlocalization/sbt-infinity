// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import "testing"

func TestIsOgg(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid OggS header", []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00"), true},
		{"RIFF header", []byte("RIFF\x00\x00\x00\x00WAVE"), false},
		{"WAVC header", []byte("WAVCV1.0"), false},
		{"too short", []byte("Ogg"), false},
		{"empty", []byte{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOgg(tt.data); got != tt.want {
				t.Errorf("IsOgg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeOgg_NotOgg(t *testing.T) {
	_, _, _, _, err := DecodeOgg([]byte("RIFF\x00\x00\x00\x00WAVE"))
	if err == nil {
		t.Error("DecodeOgg() expected error for non-Ogg data, got nil")
	}
}

func TestDecodeOgg_InvalidStream(t *testing.T) {
	// Valid OggS magic but garbage payload
	data := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	_, _, _, _, err := DecodeOgg(data)
	if err == nil {
		t.Error("DecodeOgg() expected error for invalid Ogg stream, got nil")
	}
}
