package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/davyxu/tabtoy/v3/compiler"
	"github.com/davyxu/tabtoy/v3/gen"
	"github.com/davyxu/tabtoy/v3/gen/binpak"
	"github.com/davyxu/tabtoy/v3/gen/cssrc"
	"github.com/davyxu/tabtoy/v3/gen/gosrc"
	"github.com/davyxu/tabtoy/v3/gen/jsontext"
	"github.com/davyxu/tabtoy/v3/gen/luasrc"
	"github.com/davyxu/tabtoy/v3/gen/pbttext"
	"github.com/davyxu/tabtoy/v3/gen/protosrc"
	"github.com/davyxu/tabtoy/v3/helper"
	"github.com/davyxu/tabtoy/v3/model"
	"github.com/davyxu/tabtoy/v3/report"
)

type V3GenEntry struct {
	name    string
	f       gen.GenFunc
	flagstr *string
}

// v3新增
var (
	paramIndexFile = flag.String("index", "", "input multi-files configs")

	paramUseGBKCSV = flag.Bool("use_gbkcsv", true, "use gbk format in csv file")
	paramMatchTag  = flag.String("matchtag", "", "match data table file tags in v3 Index file")

	paramProtoImportFiles       = flag.String("protoimport", "", "import .proto files paths (*.proto)")
	paramProtoOutputIgnoreFiles = flag.String("protooutputignorefile", "", "ignore output .proto files (*.proto)")

	v3GenList = []V3GenEntry{
		{"gosrc", gosrc.Generate, paramGoOut},
		{"jsontext", jsontext.Generate, paramJsonOut},
		{"luasrc", luasrc.Generate, paramLuaOut},
		{"cssrc", cssrc.Generate, paramCSharpOut},
		{"binpak", binpak.Generate, paramBinaryOut},
		{"protosrc", protosrc.Generate, paramProtoOut},
		{"pbttext", pbttext.Generate, paramPbtOut},
	}
)

func getCurrentMicroSecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func GenFile(globals *model.Globals) error {
	start := getCurrentMicroSecond()

	for _, entry := range v3GenList {

		if *entry.flagstr == "" {
			continue
		}

		filename := *entry.flagstr
		s := getCurrentMicroSecond()
		if data, err := entry.f(globals); err != nil {
			return err
		} else {
			err = helper.WriteFile(filename, data)
			if err != nil {
				return err
			}
			e := getCurrentMicroSecond()
			report.Log.Infof("  [%s] %s %.2f(s)", entry.name, filename, float64(e-s)/1000)
		}
	}

	end := getCurrentMicroSecond()
	report.Log.Debugf("Total cost %.2f seconds", float64(end-start)/1000)

	return nil
}

func V3Entry() {
	globals := model.NewGlobals()
	globals.Version = Version

	globals.IndexFile = *paramIndexFile
	globals.PackageName = *paramPackageName
	globals.CombineStructName = *paramCombineStructName
	globals.GenBinary = *paramBinaryOut != ""
	globals.MatchTag = *paramMatchTag
	globals.ProtoVersion = *paramProtoVersion

	if *paramProtoImportFiles != "" {
		globals.ProtoImportFiles = strings.Split(*paramProtoImportFiles, ";")
	}

	if *paramProtoOutputIgnoreFiles != "" {
		globals.ProtoOutputIgnoreFiles = strings.Split(*paramProtoOutputIgnoreFiles, ";")
	}

	if globals.MatchTag != "" {
		report.Log.Infof("MatchTag: %s", globals.MatchTag)
	}

	idxloader := helper.NewFileLoader(true)
	idxloader.UseGBKCSV = *paramUseGBKCSV
	globals.IndexGetter = idxloader
	globals.UseGBKCSV = *paramUseGBKCSV

	var err error

	err = compiler.Compile(globals)

	if err != nil {
		goto Exit
	}

	report.Log.Debugln("Generate files...")

	err = GenFile(globals)
	if err != nil {
		goto Exit
	}

	return
Exit:
	report.Log.Errorln(err)
	os.Exit(1)
}
