package main

import (
	"flag"
	"os"

	"github.com/davyxu/golexer"
	"github.com/davyxu/tabtoy/v2tov3"
	"github.com/davyxu/tabtoy/v2tov3/model"
	"github.com/davyxu/tabtoy/v3/helper"
)

var (
	paramUpgradeOut    = flag.String("up_out", "", "upgrade v2 table to v3 format output dir")
	paramUpgradeSuffix = flag.String("up_suffix", "", "upgrade v2 table to v3 format Index and Type filename suffix")
)

func V2ToV3Entry() {

	globals := model.NewGlobals()

	globals.TableGetter = helper.NewFileLoader(true)

	globals.SourceFileList = flag.Args()
	globals.SourceFileMetas = make(map[string]*golexer.KVPair)
	globals.OutputDir = *paramUpgradeOut
	globals.OutputFileSuffix = *paramUpgradeSuffix

	if err := v2tov3.Upgrade(globals); err != nil {
		log.Errorln(err)
		os.Exit(1)
		return
	}

}
