// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"testing"

	p "github.com/sbtlocalization/sbt-infinity/parser"
)

func TestFormatUsageSite(t *testing.T) {
	tests := []struct {
		slot, site, want string
	}{
		{"trap", "AR1234", "trap @ AR1234"},
		{"override", "CHICKEN.CRE", "override @ CHICKEN.CRE"},
		{"area", "AR0100", "area @ AR0100"},
		{"door", "AR0202", "door @ AR0202"},
	}

	for _, tt := range tests {
		if got := formatUsageSite(tt.slot, tt.site); got != tt.want {
			t.Errorf("formatUsageSite(%q, %q) = %q, want %q", tt.slot, tt.site, got, tt.want)
		}
	}
}

func TestScriptUsageIndex_Add(t *testing.T) {
	idx := make(ScriptUsageIndex)

	idx.Add("ctfb", "trap", "AR1234")
	if sites := idx["CTFB"]; len(sites) != 1 || sites[0] != "trap @ AR1234" {
		t.Errorf("expected resref to be upper-cased, got %v", idx)
	}

	// Empty slots and "no script" placeholders must be ignored.
	idx.Add("", "area", "AR0100")
	idx.Add("None", "override", "CHICKEN.CRE")
	idx.Add("*", "class", "CHICKEN.CRE")
	if len(idx) != 1 {
		t.Errorf("expected placeholders to be skipped, got %v", idx)
	}

	// A nil index silently ignores additions.
	var nilIdx ScriptUsageIndex
	nilIdx.Add("CTFB", "trap", "AR1234")
}

func TestAnnotateScriptKey(t *testing.T) {
	idx := make(ScriptUsageIndex)
	idx.Add("CTFB", "trap", "AR1234")

	if got := idx.AnnotateScriptKey("CTFB.BCS"); got != "CTFB.BCS [trap @ AR1234]" {
		t.Errorf("AnnotateScriptKey = %q, want %q", got, "CTFB.BCS [trap @ AR1234]")
	}

	// No recorded usage → bare file name.
	if got := idx.AnnotateScriptKey("NOWHERE.BCS"); got != "NOWHERE.BCS" {
		t.Errorf("AnnotateScriptKey = %q, want bare NOWHERE.BCS", got)
	}

	// Nil index → bare file name.
	var nilIdx ScriptUsageIndex
	if got := nilIdx.AnnotateScriptKey("CTFB.BCS"); got != "CTFB.BCS" {
		t.Errorf("AnnotateScriptKey on nil index = %q, want bare CTFB.BCS", got)
	}
}

func TestAnnotateScriptKey_DeduplicatesSites(t *testing.T) {
	idx := make(ScriptUsageIndex)
	idx.Add("CHICKEN", "override", "CHICKEN.CRE")
	idx.Add("CHICKEN", "override", "CHICKEN.CRE")
	idx.Add("CHICKEN", "default", "CHICKEN.CRE")

	want := "CHICKEN.BCS [default @ CHICKEN.CRE, override @ CHICKEN.CRE]"
	if got := idx.AnnotateScriptKey("CHICKEN.BCS"); got != want {
		t.Errorf("AnnotateScriptKey = %q, want %q", got, want)
	}
}

func TestAnnotateScriptKey_CapsLongSiteLists(t *testing.T) {
	idx := make(ScriptUsageIndex)
	for i := range 8 {
		idx.Add("SHOUT", "override", fmt.Sprintf("CRE%04d.CRE", i))
	}

	want := "SHOUT.BCS [" +
		"override @ CRE0000.CRE, override @ CRE0001.CRE, override @ CRE0002.CRE, " +
		"override @ CRE0003.CRE, override @ CRE0004.CRE, +3 more]"
	if got := idx.AnnotateScriptKey("SHOUT.BCS"); got != want {
		t.Errorf("AnnotateScriptKey = %q, want %q", got, want)
	}
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
