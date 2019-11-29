package v2tov3

import (
	"sort"

	"github.com/davyxu/tabtoy/util"
	"github.com/davyxu/tabtoy/v2tov3/model"
	"github.com/davyxu/tabtoy/v3/helper"
)

func ExportIndexTable(globals *model.Globals) error {
	globals.TargetIndexSheet = globals.AddTable(getIndexFileName(globals))

	helper.WriteIndexTableHeader(globals.TargetIndexSheet)

	var tabList []*helper.MemFileData

	globals.TargetTables.VisitAllTable(func(data *helper.MemFileData) bool {

		if data.FileName == getIndexFileName(globals) {
			return true
		}

		tabList = append(tabList, data)

		return true
	})

	// 内容排序
	sort.SliceStable(tabList, func(i, j int) bool {

		a := tabList[i]
		b := tabList[j]

		if aMode, bMode := getMode(globals, a), getMode(globals, b); aMode != bMode {

			if aMode == "类型表" {
				return true
			}

			if aMode == "数据表" {
				return false
			}

		}

		if a.TableName != b.TableName {
			return a.TableName < b.TableName
		}

		return a.FileName < b.FileName
	})

	for _, data := range tabList {
		mode := getMode(globals, data)
		if mode == "数据表" {
			meta, exist := globals.SourceFileMetaByName(data.TableName)
			if exist && meta.ContainValue("Vertical", "true") {
				mode = "键值表"
			}
		}
		helper.WriteRowValues(globals.TargetIndexSheet, mode, data.TableName, util.ChangeExtension(data.FileName, ".xlsx"))
	}

	return nil
}

func getMode(globals *model.Globals, data *helper.MemFileData) (mode string) {
	if data.FileName == getTypeFileName(globals) {
		mode = "类型表"
	} else {
		mode = "数据表"
	}

	return
}
