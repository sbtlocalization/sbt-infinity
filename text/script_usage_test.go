// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"testing"

	p "github.com/sbtlocalization/sbt-infinity/parser"
)

func TestScriptUsageIndex_Add(t *testing.T) {
	idx := make(ScriptUsageIndex)

	idx.Add("ctfb", "trap", "AR1234.ARE")
	want := scriptUsage{slot: "trap", file: "AR1234.ARE"}
	if sites := idx["CTFB"]; len(sites) != 1 || sites[0] != want {
		t.Errorf("recorded usage = %v, want %v", sites, want)
	}

	for _, placeholder := range []string{"", "None", "*", "0"} {
		idx.Add(placeholder, "area", "AR0100.ARE")
	}
	if len(idx) != 1 {
		t.Errorf("expected placeholders to be skipped, got %v", idx)
	}

	var nilIdx ScriptUsageIndex
	nilIdx.Add("CTFB", "trap", "AR1234.ARE")
}

func TestRegionSlotName(t *testing.T) {
	tests := []struct {
		regionType p.Are_Region_RegionType
		want       string
	}{
		{p.Are_Region_RegionType__ProximityTrigger, "trap"},
		{p.Are_Region_RegionType__InfoPoint, "info point"},
		{p.Are_Region_RegionType__TravelRegion, "travel"},
		{p.Are_Region_RegionType(42), "region"},
	}

	for _, tt := range tests {
		if got := regionSlotName(tt.regionType); got != tt.want {
			t.Errorf("regionSlotName(%v) = %q, want %q", tt.regionType, got, tt.want)
		}
	}
}
