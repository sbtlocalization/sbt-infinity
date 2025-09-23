package parser

var typeToExtension = map[Key_ResEntry_ResType]string{
	Key_ResEntry_ResType__Bmp:   "BMP",
	Key_ResEntry_ResType__Mve:   "MVE",
	Key_ResEntry_ResType__Wav:   "WAV",
	Key_ResEntry_ResType__Wfx:   "WFX",
	Key_ResEntry_ResType__Plt:   "PLT",
	Key_ResEntry_ResType__Tga:   "TGA",
	Key_ResEntry_ResType__Bam:   "BAM",
	Key_ResEntry_ResType__Wed:   "WED",
	Key_ResEntry_ResType__Chu:   "CHU",
	Key_ResEntry_ResType__Tis:   "TIS",
	Key_ResEntry_ResType__Mos:   "MOS",
	Key_ResEntry_ResType__Itm:   "ITM",
	Key_ResEntry_ResType__Spl:   "SPL",
	Key_ResEntry_ResType__Bcs:   "BCS",
	Key_ResEntry_ResType__Ids:   "IDS",
	Key_ResEntry_ResType__Cre:   "CRE",
	Key_ResEntry_ResType__Are:   "ARE",
	Key_ResEntry_ResType__Dlg:   "DLG",
	Key_ResEntry_ResType__TwoDa: "2DA",
	Key_ResEntry_ResType__Gam:   "GAM",
	Key_ResEntry_ResType__Sto:   "STO",
	Key_ResEntry_ResType__Wmp:   "WMP",
	Key_ResEntry_ResType__Eff:   "EFF",
	Key_ResEntry_ResType__Bs:    "BS",
	Key_ResEntry_ResType__Chr:   "CHR",
	Key_ResEntry_ResType__Vvc:   "VVC",
	Key_ResEntry_ResType__Vef:   "VEF",
	Key_ResEntry_ResType__Pro:   "PRO",
	Key_ResEntry_ResType__Bio:   "BIO",
	Key_ResEntry_ResType__Wbm:   "WBM",
	Key_ResEntry_ResType__Fnt:   "FNT",
	Key_ResEntry_ResType__Gui:   "GUI",
	Key_ResEntry_ResType__Sql:   "SQL",
	Key_ResEntry_ResType__Pvrz:  "PVRZ",
	Key_ResEntry_ResType__Glsl:  "GLSL",
	Key_ResEntry_ResType__Tot:   "TOT",
	Key_ResEntry_ResType__Toh:   "TOH",
	Key_ResEntry_ResType__Menu:  "MENU",
	Key_ResEntry_ResType__Lua:   "LUA",
	Key_ResEntry_ResType__Ttf:   "TTF",
	Key_ResEntry_ResType__Png:   "PNG",
	Key_ResEntry_ResType__Bah:   "BAH",
	Key_ResEntry_ResType__Ini:   "INI",
	Key_ResEntry_ResType__Src:   "SRC",
	Key_ResEntry_ResType__Maze:  "MAZE",
	Key_ResEntry_ResType__Mus:   "MUS",
	Key_ResEntry_ResType__Acm:   "ACM",
}

func (t Key_ResEntry_ResType) String() string {
	if ext, ok := typeToExtension[t]; ok {
		return ext
	}
	return "unknown"
}
