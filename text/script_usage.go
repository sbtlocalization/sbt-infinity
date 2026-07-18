// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/samber/lo"
	p "github.com/sbtlocalization/sbt-infinity/parser"
)

// ScriptUsageIndex maps a script's resref to the sites where the script is
// attached in the game — area/trap/door/container/actor slots of an ARE, or
// one of the five script slots of a CRE. Populated by the area and creature
// scans, consumed when SCRIPTS keys are built.
type ScriptUsageIndex map[string][]string

// A key annotated with too many sites is unreadable in the spreadsheet, so
// the tail is collapsed into a "+N more" marker.
const maxUsageSitesInKey = 5

func (idx ScriptUsageIndex) Add(scriptResRef, slot, site string) {
	if idx == nil {
		return
	}

	res := strings.ToUpper(strings.TrimSpace(scriptResRef))
	if res == "" || res == "NONE" || res == "*" || res == "0" {
		return
	}

	idx[res] = append(idx[res], formatUsageSite(slot, site))
}

func (idx ScriptUsageIndex) AnnotateScriptKey(scriptFilename string) string {
	sites := lo.Uniq(idx[resourceName(scriptFilename)])
	if len(sites) == 0 {
		return scriptFilename
	}

	slices.Sort(sites)
	if extra := len(sites) - maxUsageSitesInKey; extra > 0 {
		sites = append(sites[:maxUsageSitesInKey], fmt.Sprintf("+%d more", extra))
	}

	return fmt.Sprintf("%s [%s]", scriptFilename, strings.Join(sites, ", "))
}

func formatUsageSite(slot, site string) string {
	return fmt.Sprintf("%s @ %s", slot, site)
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
