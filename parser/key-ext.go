package parser

var typeToExtension = map[Key_ResType]string{
	Key_ResType__Bmp:   "BMP",
	Key_ResType__Mve:   "MVE",
	Key_ResType__Wav:   "WAV",
	Key_ResType__Wfx:   "WFX",
	Key_ResType__Plt:   "PLT",
	Key_ResType__Tga:   "TGA",
	Key_ResType__Bam:   "BAM",
	Key_ResType__Wed:   "WED",
	Key_ResType__Chu:   "CHU",
	Key_ResType__Tis:   "TIS",
	Key_ResType__Mos:   "MOS",
	Key_ResType__Itm:   "ITM",
	Key_ResType__Spl:   "SPL",
	Key_ResType__Bcs:   "BCS",
	Key_ResType__Ids:   "IDS",
	Key_ResType__Cre:   "CRE",
	Key_ResType__Are:   "ARE",
	Key_ResType__Dlg:   "DLG",
	Key_ResType__TwoDa: "2DA",
	Key_ResType__Gam:   "GAM",
	Key_ResType__Sto:   "STO",
	Key_ResType__Wmp:   "WMP",
	Key_ResType__Eff:   "EFF",
	Key_ResType__Bs:    "BS",
	Key_ResType__Chr:   "CHR",
	Key_ResType__Vvc:   "VVC",
	Key_ResType__Vef:   "VEF",
	Key_ResType__Pro:   "PRO",
	Key_ResType__Bio:   "BIO",
	Key_ResType__Wbm:   "WBM",
	Key_ResType__Fnt:   "FNT",
	Key_ResType__Gui:   "GUI",
	Key_ResType__Sql:   "SQL",
	Key_ResType__Pvrz:  "PVRZ",
	Key_ResType__Glsl:  "GLSL",
	Key_ResType__Tot:   "TOT",
	Key_ResType__Toh:   "TOH",
	Key_ResType__Menu:  "MENU",
	Key_ResType__Lua:   "LUA",
	Key_ResType__Ttf:   "TTF",
	Key_ResType__Png:   "PNG",
	Key_ResType__Bah:   "BAH",
	Key_ResType__Ini:   "INI",
	Key_ResType__Src:   "SRC",
	Key_ResType__Maze:  "MAZE",
	Key_ResType__Mus:   "MUS",
	Key_ResType__Acm:   "ACM",
}

func (t Key_ResType) String() string {
	if ext, ok := typeToExtension[t]; ok {
		return ext
	}
	return "unknown"
}
