package v2tov3

import (
	"github.com/davyxu/golexer"
	"github.com/davyxu/tabtoy/v2tov3/model"
	"github.com/davyxu/tabtoy/v3/helper"
	v3model "github.com/davyxu/tabtoy/v3/model"
	"github.com/davyxu/tabtoy/v3/report"
	"github.com/tealeg/xlsx"
	"reflect"
	"strings"
)

func importDataHeader(globals *model.Globals, sourceSheet, targetSheet *xlsx.Sheet, tabPragma *golexer.KVPair, isMergeFile bool) (headerList []model.ObjectFieldType) {
	isVertical := false
	tableName := tabPragma.GetString("TableName")
	// 判断是否纵表
	if tabPragma.ContainValue("Vertical", "true") {
		isVertical = true
	}
	if isVertical {
		headerList = setImportDataHeaderVertical(sourceSheet, targetSheet, tableName)
	} else {
		headerList = setImportDataHeader(globals, sourceSheet, targetSheet, tableName, isMergeFile)
	}
	return
}

func setImportDataHeader(globals *model.Globals, sourceSheet, targetSheet *xlsx.Sheet, tableName string, isMergeFile bool) (headerList []model.ObjectFieldType) {
	var headerRow *xlsx.Row
	// 遍历所有列
	for col := 0; ; col++ {
		var oft model.ObjectFieldType
		oft.ObjectType = tableName
		oft.Kind = v3model.TypeUsage_HeaderStruct
		oft.FieldName = helper.GetSheetValueString(sourceSheet, 0, col)
		// 空列，终止
		if oft.FieldName == "" {
			break
		}
		// 列头中带有#的，特别是最后一行
		if strings.HasPrefix(oft.FieldName, "#") {
			continue
		}
		if headerRow == nil {
			headerRow = targetSheet.AddRow()
		}
		oft.FieldType = helper.GetSheetValueString(sourceSheet, 1, col)
		// 元信息
		meta := helper.GetSheetValueString(sourceSheet, 2, col)
		oft.Meta = golexer.NewKVPair()
		if err := oft.Meta.Parse(meta); err != nil {
			continue
		}

		if strings.HasPrefix(oft.FieldType, "[]") || strings.HasPrefix(oft.FieldType, "repeated") {
			//oft.FieldType = oft.FieldType[2:]
			oft.FieldType = strings.TrimLeft(oft.FieldType, "[]")
			oft.FieldType = strings.TrimLeft(oft.FieldType, "repeated ")
			oft.ArraySplitter = oft.Meta.GetString("ListSpliter")

			if oft.ArraySplitter == "" {
				log.Warnln("array list no ListSpliter:", oft.FieldName, oft.ObjectType)
			}
		}

		if isMergeFile {
			oft2 := globals.ObjectTypeByObjectTypeAndFieldName(oft.ObjectType, oft.FieldName)
			oft.Name = oft2.Name
		} else {
			oft.Name = helper.GetSheetValueString(sourceSheet, 3, col)
		}
		if oft.Name == "" {
			log.Warnf("v2的字段注释为空, %s | %s", oft.FieldName, tableName)
			oft.Name = oft.FieldName
		}

		var disabledForV3 string
		// 添加V3表头
		if globals.TypeIsNoneKind(oft.FieldType) {
			disabledForV3 = "#"
		}
		// 结构体等类型，标记为none，输出为#
		if !model.IsNativeType(oft.FieldType) {

			targetOft := globals.ObjectTypeByObjectType(oft.FieldType)
			// 类型已经被前置定义，且不是枚举（那就是结构体）时，标记为空，后面不会被使用
			if targetOft != nil && targetOft.Kind != v3model.TypeUsage_Enum {
				//oft.Kind = v3model.TypeUsage_None
				report.Log.Debugf("oft.FieldType: %s, targetOft.Kind: %s", oft.FieldType, targetOft.Kind)
			}

		}
		// 新表的表头加列
		headerRow.AddCell().SetValue(disabledForV3 + oft.Name)
		// 拆分字段填充的数组
		if !globals.SourceTypeExists(oft.ObjectType, oft.FieldName) {
			globals.AddSourceType(oft)
		}
		headerList = append(headerList, oft)
	}
	return
}

type typeDefine struct {
	FieldName     string `tb_name:"字段名" tg_name:"字段名"`
	FieldType     string `tb_name:"类型" tg_name:"字段类型"`
	Name          string `tb_name:"注释" tg_name:"标识名"`
	Value         string `tb_name:"值" tg_name:"值" json:",omitempty"`
	ArraySplitter string `tb_name:"特性" tg_name:"数组切割" json:",omitempty"`
}

func (p *typeDefine) matchFieldByTableName(field string) int {
	objType := reflect.TypeOf(p).Elem()
	for i := 0; i < objType.NumField(); i++ {
		fd := objType.Field(i)
		if fd.Tag.Get("tb_name") == field {
			return i
		}
	}
	return -1
}

func (p *typeDefine) getTargetNameByTableName(field string) string {
	objType := reflect.TypeOf(p).Elem()
	for i := 0; i < objType.NumField(); i++ {
		fd := objType.Field(i)
		if fd.Tag.Get("tb_name") == field {
			return fd.Tag.Get("tg_name")
		}
	}
	return ""
}

func setImportDataHeaderVertical(sourceSheet, targetSheet *xlsx.Sheet, tableName string) (headerList []model.ObjectFieldType) {
	var headerRow *xlsx.Row
	header := &typeDefine{}
	// 遍历所有列
	for col := 0; ; col++ {
		var oft model.ObjectFieldType
		oft.ObjectType = tableName
		oft.Kind = v3model.TypeUsage_HeaderStruct
		oft.FieldName = helper.GetSheetValueString(sourceSheet, 0, col)
		// 空列，终止
		if oft.FieldName == "" {
			break
		}
		if headerRow == nil {
			headerRow = targetSheet.AddRow()
		}
		oft.Name = header.getTargetNameByTableName(oft.FieldName)
		if oft.Name == "" {
			continue
		}
		// 新表的表头加列
		headerRow.AddCell().SetValue(oft.Name)
		headerList = append(headerList, oft)
	}
	return
}
