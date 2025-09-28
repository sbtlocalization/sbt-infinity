// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package fs

import (
	"strings"

	"github.com/sbtlocalization/sbt-infinity/parser"
)

type FileType int

const (
	FileType_Invalid FileType = 0
	FileType_BMP     FileType = 1
	FileType_MVE     FileType = 2
	FileType_WAV     FileType = 4
	FileType_WFX     FileType = 5
	FileType_PLT     FileType = 6
	FileType_TGA     FileType = 952
	FileType_BAM     FileType = 1000
	FileType_WED     FileType = 1001
	FileType_CHU     FileType = 1002
	FileType_TIS     FileType = 1003
	FileType_MOS     FileType = 1004
	FileType_ITM     FileType = 1005
	FileType_SPL     FileType = 1006
	FileType_BCS     FileType = 1007
	FileType_IDS     FileType = 1008
	FileType_CRE     FileType = 1009
	FileType_ARE     FileType = 1010
	FileType_DLG     FileType = 1011
	FileType_2DA     FileType = 1012
	FileType_GAM     FileType = 1013
	FileType_STO     FileType = 1014
	FileType_WMP     FileType = 1015
	FileType_EFF     FileType = 1016
	FileType_BS      FileType = 1017
	FileType_CHR     FileType = 1018
	FileType_VVC     FileType = 1019
	FileType_VEF     FileType = 1020
	FileType_PRO     FileType = 1021
	FileType_BIO     FileType = 1022
	FileType_WBM     FileType = 1023
	FileType_FNT     FileType = 1024
	FileType_GUI     FileType = 1026
	FileType_SQL     FileType = 1027
	FileType_PVRZ    FileType = 1028
	FileType_GLSL    FileType = 1029
	FileType_TOT     FileType = 1030
	FileType_TOH     FileType = 1031
	FileType_MENU    FileType = 1032
	FileType_LUA     FileType = 1033
	FileType_TTF     FileType = 1034
	FileType_PNG     FileType = 1035
	FileType_BAH     FileType = 1100
	FileType_INI     FileType = 2050
	FileType_SRC     FileType = 2051
	FileType_MAZE    FileType = 2052
	FileType_MUS     FileType = 4094
	FileType_ACM     FileType = 4095
)

func (t FileType) IsValid() bool {
	_, ok := typeToExtension[t]
	return ok
}

func (t FileType) String() string {
	if ext, ok := typeToExtension[t]; ok {
		return ext
	}
	return "unknown"
}

func (t FileType) ToParserType() parser.Key_ResType {
	return parser.Key_ResType(t)
}

func FileTypeFromExtension(ext string) FileType {
	ext = strings.ToUpper(ext)
	if t, ok := extensionToType[ext]; ok {
		return t
	}
	return FileType_Invalid
}

func FileTypeFromParserType(t parser.Key_ResType) FileType {
	return FileType(t)
}

var typeToExtension = map[FileType]string{
	FileType_BMP:  "BMP",
	FileType_MVE:  "MVE",
	FileType_WAV:  "WAV",
	FileType_WFX:  "WFX",
	FileType_PLT:  "PLT",
	FileType_TGA:  "TGA",
	FileType_BAM:  "BAM",
	FileType_WED:  "WED",
	FileType_CHU:  "CHU",
	FileType_TIS:  "TIS",
	FileType_MOS:  "MOS",
	FileType_ITM:  "ITM",
	FileType_SPL:  "SPL",
	FileType_BCS:  "BCS",
	FileType_IDS:  "IDS",
	FileType_CRE:  "CRE",
	FileType_ARE:  "ARE",
	FileType_DLG:  "DLG",
	FileType_2DA:  "2DA",
	FileType_GAM:  "GAM",
	FileType_STO:  "STO",
	FileType_WMP:  "WMP",
	FileType_EFF:  "EFF",
	FileType_BS:   "BS",
	FileType_CHR:  "CHR",
	FileType_VVC:  "VVC",
	FileType_VEF:  "VEF",
	FileType_PRO:  "PRO",
	FileType_BIO:  "BIO",
	FileType_WBM:  "WBM",
	FileType_FNT:  "FNT",
	FileType_GUI:  "GUI",
	FileType_SQL:  "SQL",
	FileType_PVRZ: "PVRZ",
	FileType_GLSL: "GLSL",
	FileType_TOT:  "TOT",
	FileType_TOH:  "TOH",
	FileType_MENU: "MENU",
	FileType_LUA:  "LUA",
	FileType_TTF:  "TTF",
	FileType_PNG:  "PNG",
	FileType_BAH:  "BAH",
	FileType_INI:  "INI",
	FileType_SRC:  "SRC",
	FileType_MAZE: "MAZE",
	FileType_MUS:  "MUS",
	FileType_ACM:  "ACM",
}

var extensionToType = map[string]FileType{
	"BMP":  FileType_BMP,
	"MVE":  FileType_MVE,
	"WAV":  FileType_WAV,
	"WFX":  FileType_WFX,
	"PLT":  FileType_PLT,
	"TGA":  FileType_TGA,
	"BAM":  FileType_BAM,
	"WED":  FileType_WED,
	"CHU":  FileType_CHU,
	"TIS":  FileType_TIS,
	"MOS":  FileType_MOS,
	"ITM":  FileType_ITM,
	"SPL":  FileType_SPL,
	"BCS":  FileType_BCS,
	"IDS":  FileType_IDS,
	"CRE":  FileType_CRE,
	"ARE":  FileType_ARE,
	"DLG":  FileType_DLG,
	"2DA":  FileType_2DA,
	"GAM":  FileType_GAM,
	"STO":  FileType_STO,
	"WMP":  FileType_WMP,
	"EFF":  FileType_EFF,
	"BS":   FileType_BS,
	"CHR":  FileType_CHR,
	"VVC":  FileType_VVC,
	"VEF":  FileType_VEF,
	"PRO":  FileType_PRO,
	"BIO":  FileType_BIO,
	"WBM":  FileType_WBM,
	"FNT":  FileType_FNT,
	"GUI":  FileType_GUI,
	"SQL":  FileType_SQL,
	"PVRZ": FileType_PVRZ,
	"GLSL": FileType_GLSL,
	"TOT":  FileType_TOT,
	"TOH":  FileType_TOH,
	"MENU": FileType_MENU,
	"LUA":  FileType_LUA,
	"TTF":  FileType_TTF,
	"PNG":  FileType_PNG,
	"BAH":  FileType_BAH,
	"INI":  FileType_INI,
	"SRC":  FileType_SRC,
	"MAZE": FileType_MAZE,
	"MUS":  FileType_MUS,
	"ACM":  FileType_ACM,
}
