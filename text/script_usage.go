// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"path"
	"strings"

	p "github.com/sbtlocalization/sbt-infinity/parser"
)

type scriptUsage struct {
	slot string
	file string
}

type ScriptUsageIndex map[string][]scriptUsage

func (idx ScriptUsageIndex) Add(scriptResRef, slot, site string) {
	if idx == nil {
		return
	}

	res := strings.ToUpper(strings.TrimSpace(scriptResRef))
	if res == "" || res == "NONE" || res == "*" || res == "0" {
		return
	}

	idx[res] = append(idx[res], scriptUsage{slot: slot, file: site})
}

func regionSlotName(regionType p.Are_Region_RegionType) string {
	switch regionType {
	case p.Are_Region_RegionType__ProximityTrigger:
		return "trap"
	case p.Are_Region_RegionType__InfoPoint:
		return "info point"
	case p.Are_Region_RegionType__TravelRegion:
		return "travel"
	default:
		return "region"
	}
}

func resourceName(filename string) string {
	return strings.ToUpper(strings.TrimSuffix(filename, path.Ext(filename)))
}
