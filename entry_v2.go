package main

import (
	"flag"
	"os"
	"strings"

	"github.com/davyxu/tabtoy/build"
	v2 "github.com/davyxu/tabtoy/v2"
	"github.com/davyxu/tabtoy/v2/i18n"
	"github.com/davyxu/tabtoy/v2/printer"
)

// v2特有
var (
	paramProtoVersion = flag.Int("protover", 3, "output .proto file version, 2 or 3")

	paramLuaEnumIntValue = flag.Bool("luaenumintvalue", false, "use int type in lua enum value")
	paramLuaTabHeader    = flag.String("luatabheader", "", "output string to lua tab header")

	paramGenCSharpBinarySerializeCode = flag.Bool("cs_gensercode", true, "generate c# binary serialize code, default is true")

	paramProtoImportFiles       = flag.String("protoimport", "", "import .proto files paths (*.proto)")
	paramGoPackage              = flag.String("gopackage", "", "import go package")
	paramProtoOutputIgnoreFiles = flag.String("protooutputignorefile", "", "ignore output .proto files (*.proto)")
	paramFieldTags              = flag.String("fieldOutTag", "", "filter field OutTags (separator: ';')")
)

func V2Entry() {
	g := printer.NewGlobals()

	if !i18n.SetLanguage(*paramLanguage) {
		log.Infof("language not support: %s", *paramLanguage)
		os.Exit(1)
	}

	g.Version = build.Version

	for _, v := range flag.Args() {
		g.InputFileList = append(g.InputFileList, v)
	}

	g.ParaMode = *paramPara
	g.CacheDir = *paramCacheDir
	g.UseCache = *paramUseCache
	g.CombineStructName = *paramCombineStructName
	g.ProtoVersion = *paramProtoVersion
	g.LuaEnumIntValue = *paramLuaEnumIntValue
	g.LuaTabHeader = *paramLuaTabHeader
	g.GenCSSerailizeCode = *paramGenCSharpBinarySerializeCode
	g.PackageName = *paramPackageName

	if *paramProtoOut != "" {
		g.AddOutputType("proto", *paramProtoOut)
	}

	if *paramPbtOut != "" {
		g.AddOutputType("pbt", *paramPbtOut)
	}

	if *paramJsonOut != "" {
		g.AddOutputType("json", *paramJsonOut)
	}

	if *paramLuaOut != "" {
		g.AddOutputType("lua", *paramLuaOut)
	}

	if *paramCSharpOut != "" {
		g.AddOutputType("cs", *paramCSharpOut)
	}

	if *paramGoOut != "" {
		g.AddOutputType("go", *paramGoOut)
	}

	if *paramCppOut != "" {
		g.AddOutputType("cpp", *paramCppOut)
	}

	if *paramBinaryOut != "" {
		g.AddOutputType("bin", *paramBinaryOut)
	}

	if *paramTypeOut != "" {
		g.AddOutputType("type", *paramTypeOut)
	}

	if *paramModifyList != "" {
		g.AddOutputType("modlist", *paramModifyList)
	}

	if *paramProtoImportFiles != "" {
		g.ProtoImportFiles = strings.Split(*paramProtoImportFiles, ",")
	}

	if *paramGoPackage != "" {
		g.ProtoGoPackage = *paramGoPackage
	}

	if *paramProtoOutputIgnoreFiles != "" {
		g.ProtoOutputIgnoreFiles = strings.Split(*paramProtoOutputIgnoreFiles, ";")
	}

	if len(*paramFieldTags) != 0 {
		v2.FieldOutTags = strings.Split(*paramFieldTags, ";")
	}

	if !v2.Run(g) {
		os.Exit(1)
	}
}
