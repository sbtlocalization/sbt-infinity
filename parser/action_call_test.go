// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package parser

import (
	"reflect"
	"testing"
)

func TestParseActionCalls(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []ActionCall
	}{
		{
			name: "single display action",
			text: "DisplayStringHead(Myself,14381)",
			want: []ActionCall{
				{Name: "DisplayStringHead", Args: []string{"Myself", "14381"}},
			},
		},
		{
			name: "quoted string with comma is not split",
			text: `SetGlobal("v","GLOBAL",1)`,
			want: []ActionCall{
				{Name: "SetGlobal", Args: []string{`"v"`, `"GLOBAL"`, "1"}},
			},
		},
		{
			name: "nested parens in object spec",
			text: "ActionOverride(Player1,DisplayStringHead(Myself,14381))",
			want: []ActionCall{
				{Name: "ActionOverride", Args: []string{"Player1", "DisplayStringHead(Myself,14381)"}},
			},
		},
		{
			name: "multiple statements one per line",
			text: "DisplayStringHead(Myself,14381)\nSetGlobal(\"v\",\"GLOBAL\",1)",
			want: []ActionCall{
				{Name: "DisplayStringHead", Args: []string{"Myself", "14381"}},
				{Name: "SetGlobal", Args: []string{`"v"`, `"GLOBAL"`, "1"}},
			},
		},
		{
			name: "no arguments",
			text: "EndDialog()",
			want: []ActionCall{
				{Name: "EndDialog", Args: nil},
			},
		},
		{
			name: "empty text",
			text: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseActionCalls(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseActionCalls(%q) = %#v, want %#v", tt.text, got, tt.want)
			}
		})
	}
}
