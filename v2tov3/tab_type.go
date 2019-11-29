package v2tov3

import (
	"errors"
	"fmt"
	"strings"

	"github.com/davyxu/golexer"
	"github.com/davyxu/tabtoy/v2tov3/model"
	"github.com/davyxu/tabtoy/v3/helper"
	v3model "github.com/davyxu/tabtoy/v3/model"
	"github.com/tealeg/xlsx"
)

func ExportTypes(globals *model.Globals) error {

	for _, oft := range globals.SourceTypes {

		var disableKind string
		if oft.Kind == v3model.TypeUsage_None {
			disableKind = "#"
		}

		helper.WriteRowValues(globals.TargetTypesSheet,
			disableKind+oft.Kind.String(),
			oft.ObjectType,
			oft.Name,
			oft.FieldName,
			oft.FieldType,
			oft.ArraySplitter,
			oft.Value)
	}

	return nil
}

func importTypes(globals *model.Globals, sheet *xlsx.Sheet, tabPragma *golexer.KVPair, fileName string, isMergeFile bool) error {
	pragma := helper.GetSheetValueString(sheet, 0, 0)
	if err := tabPragma.Parse(pragma); err != nil {
		return err
	}

	if !isMergeFile {
		tableName := tabPragma.GetString("TableName")
		if globals.SourceFileMetaExists(tableName) {
			return errors.New(fmt.Sprintf("重复定义的表名 %s", tableName))
		}
		globals.AddSourceFileMeta(tableName, tabPragma)
	}
	// 匹配表头
	headerList := make(map[string]int)
	for col := 0; ; col++ {
		value := helper.GetSheetValueString(sheet, 2, col)
		// 空列，终止
		if value == "" {
			break
		}
		headerList[value] = col
	}
	// 遍历所有行
	for row := 3; ; row++ {
		var oft model.ObjectFieldType
		// 遍历所有列
		for name, col := range headerList {
			value := helper.GetSheetValueString(sheet, row, col)
			switch name {
			case "对象类型":
				oft.ObjectType = value
			case "字段名":
				oft.FieldName = value
			case "字段类型":
				oft.FieldType = value
				// V3无需添加数组前缀
				//if strings.HasPrefix(oft.FieldType, "[]") {
				//	oft.FieldType = oft.FieldType[2:]
				//}
				oft.FieldType = strings.TrimLeft(oft.FieldType, "[]")
				oft.FieldType = strings.TrimLeft(oft.FieldType, "repeated ")
			case "枚举值":
				oft.Value = value
				if oft.Value == "" {
					//oft.Kind = v3model.TypeUsage_None
					oft.Kind = v3model.TypeUsage_HeaderStruct
				} else {
					oft.Kind = v3model.TypeUsage_Enum
				}
			case "别名":
				oft.Name = value
			case "数组切割":
				// 元信息
				meta := golexer.NewKVPair()
				if err := meta.Parse(value); err != nil {
					continue
				}
				oft.Meta = meta
			}
		}

		// 空行，终止
		if oft.ObjectType == "" {
			break
		}

		if globals.SourceTypeExists(oft.ObjectType, oft.FieldName) {
			return errors.New(fmt.Sprintf("重复定义的类型 %s %s @ %s", oft.ObjectType, oft.FieldName, fileName))
		} else {
			globals.AddSourceType(oft)
		}
	}

	return nil
}
