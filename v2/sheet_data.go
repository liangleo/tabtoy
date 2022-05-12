package v2

import (
	"strings"

	"github.com/davyxu/tabtoy/util"
	"github.com/davyxu/tabtoy/v2/i18n"
	"github.com/davyxu/tabtoy/v2/model"
)

/*
	Sheet数据表单的处理

*/

type outputDataFilterTag struct {
	tableName        string
	fieldName        string
	filterConfIDs    map[string]struct{}
	filterFieldNames map[string]struct{}
}

var outputDataFilterTags map[string]map[string]map[string]map[string]struct{}

func parseOutputDataFilterValue(node *model.Node) map[string]struct{} {
	values := make(map[string]struct{}, len(node.Child))
	// 普通值
	if node.Type != model.FieldType_Struct {
		if node.IsRepeated {
			// repeated 值序列
			for _, valueNode := range node.Child {
				values[valueNode.Value] = struct{}{}
			}
		} else {
			// 单值
			valueNode := node.Child[0]
			values[valueNode.Value] = struct{}{}
		}
	} else {
		// 遍历repeated的结构体
		for _, structNode := range node.Child {
			// 遍历一个结构体的字段
			for _, fieldNode := range structNode.Child {
				if fieldNode.SugguestIgnore {
					continue
				}
				// 值节点总是在第一个
				valueNode := fieldNode.Child[0]
				values[valueNode.Value] = struct{}{}
			}
		}
	}
	return values
}

func InitOutputDataFilterTags(fileName string) {
	file, _ := NewFile(fileName, "")
	file.ExportLocalType(nil)

	dataModel := model.NewDataModel()
	file.ExportData(dataModel, nil)

	tab := model.NewTable()
	tab.LocalFD = file.LocalFD
	mergeValues(dataModel, tab, file)
	// 遍历每一行
	outputDataFilterTags = make(map[string]map[string]map[string]map[string]struct{})
	for _, r := range tab.Recs {
		filterTag := &outputDataFilterTag{}
		for _, node := range r.Nodes {
			if node.SugguestIgnore && !node.IsRepeated {
				continue
			}
			if node.Child == nil || len(node.Child) == 0 {
				continue
			}
			if node.Type == model.FieldType_Struct {
				continue
			}

			switch node.Name {
			case "TableName":
				valueNode := node.Child[0]
				filterTag.tableName = valueNode.Value
			case "FieldName":
				valueNode := node.Child[0]
				filterTag.fieldName = valueNode.Value
			case "FilterConfID":
				filterTag.filterConfIDs = parseOutputDataFilterValue(node)
			case "FilterFieldName":
				filterTag.filterFieldNames = parseOutputDataFilterValue(node)
			default:
			}
		}
		if outputDataFilterTags[filterTag.tableName] == nil {
			outputDataFilterTags[filterTag.tableName] = make(map[string]map[string]map[string]struct{})
		}
		if outputDataFilterTags[filterTag.tableName][filterTag.fieldName] == nil {
			outputDataFilterTags[filterTag.tableName][filterTag.fieldName] = make(map[string]map[string]struct{})
		}
		for confID := range filterTag.filterConfIDs {
			outputDataFilterTags[filterTag.tableName][filterTag.fieldName][confID] = filterTag.filterFieldNames
		}
	}
}

func getOutputDataFilterFields(tableName, fieldName, fieldValue string) (map[string]struct{}, bool) {
	tableFields, exist := outputDataFilterTags[tableName]
	if !exist {
		return nil, false
	}
	confFields, exist := tableFields[fieldName]
	if !exist {
		return nil, false
	}
	fields, exist := confFields[fieldValue]
	if !exist {
		return nil, false
	}
	return fields, true
}

type DataSheet struct {
	*Sheet
}

func (self *DataSheet) Valid() bool {

	name := strings.TrimSpace(self.Sheet.Name)
	if name != "" && name[0] == '#' {
		return false
	}

	return self.GetCellData(0, 0) != ""
}

func (self *DataSheet) Export(file *File, dataModel *model.DataModel, dataHeader, parentHeader *DataHeader) bool {

	verticalHeader := file.LocalFD.Pragma.GetBool("Vertical")

	if verticalHeader {
		return self.exportColumnMajor(file, dataModel, dataHeader, parentHeader)
	} else {
		return self.exportRowMajor(file, dataModel, dataHeader, parentHeader)
	}

}

// 导出以行数据延展的表格(普通表格)
func (self *DataSheet) exportRowMajor(file *File, dataModel *model.DataModel, dataHeader, parentHeader *DataHeader) bool {
	// 是否继续读行
	var readingLine bool = true

	var meetEmptyLine bool

	var warningAfterEmptyLineDataOnce bool

	// 遍历每一行
	for self.Row = DataSheetHeader_DataBegin; readingLine; self.Row++ {

		// 整行都是空的
		if self.IsFullRowEmpty(self.Row, dataHeader.RawFieldCount()) {

			// 再次碰空行, 表示确实是空的
			if meetEmptyLine {
				break

			} else {
				meetEmptyLine = true
			}

			continue

		} else {

			//已经碰过空行, 这里又碰到数据, 说明有人为隔出的空行, 做warning提醒, 防止数据没导出
			if meetEmptyLine && !warningAfterEmptyLineDataOnce {
				r, _ := self.GetRC()

				log.Warnf("%s %s|%s(%s)", i18n.String(i18n.DataSheet_RowDataSplitedByEmptyLine), self.file.FileName, self.Name, util.R1C1ToA1(r, 1))

				warningAfterEmptyLineDataOnce = true
			}

			// 曾经有过空行, 即便现在不是空行也没用, 结束
			if meetEmptyLine {
				break
			}

		}

		line := model.NewLineData()

		// 遍历每一列
		for self.Column = 0; self.Column < dataHeader.RawFieldCount(); self.Column++ {

			fieldDef, ok := fieldDefGetter(self.Column, dataHeader, parentHeader)

			if !ok {
				log.Errorf("%s %s|%s(%s)", i18n.String(i18n.DataHeader_FieldNotDefinedInMainTableInMultiTableMode), self.file.FileName, self.Name, util.R1C1ToA1(self.Row+1, self.Column+1))
				return false
			}

			op := self.processLine(fieldDef, line, dataHeader)

			if op == lineOp_Continue {
				continue
			} else if op == lineOp_Break {
				break
			}

		}

		// 是子表
		if parentHeader != nil {

			// 遍历母表所有的列头字段
			for c := 0; c < parentHeader.RawFieldCount(); c++ {
				fieldDef := parentHeader.RawField(c)

				// 在子表中有对应字段的, 忽略, 只要没有的字段
				if _, ok := dataHeader.HeaderByName[fieldDef.Name]; ok {
					continue
				}

				op := self.processLine(fieldDef, line, dataHeader)

				if op == lineOp_Continue {
					continue
				} else if op == lineOp_Break {
					break
				}

			}
		}

		// 判断是否需要过滤此行数据
		for _, field := range line.Values {
			fields, exist := getOutputDataFilterFields(file.FileName, field.FieldDef.Name, field.RawValue)
			if !exist {
				continue
			}
			dataModel.FilterFields = fields
			line.NeedFilter = true
			break
		}

		dataModel.Add(line)

	}

	return true
}

const (
	lineOp_none = iota
	lineOp_Break
	lineOp_Continue
)

func (self *DataSheet) processLine(fieldDef *model.FieldDescriptor, line *model.LineData, dataHeader *DataHeader) int {
	// 数据大于列头时, 结束这个列
	if fieldDef == nil {
		return lineOp_Break
	}

	// #开头表示注释, 跳过
	if strings.Index(fieldDef.Name, "#") == 0 {
		return lineOp_Continue
	}

	var rawValue string

	// 浮点数按本来的格式输出
	if fieldDef.Type == model.FieldType_Float && !fieldDef.IsRepeated {
		rawValue = self.GetCellDataAsNumeric(self.Row, self.Column)
	} else {
		rawValue = self.GetCellData(self.Row, self.Column)
	}

	r, c := self.GetRC()

	line.Add(&model.FieldValue{
		FieldDef:           fieldDef,
		RawValue:           rawValue,
		SheetName:          self.Name,
		FileName:           self.file.FileName,
		R:                  r,
		C:                  c,
		FieldRepeatedCount: dataHeader.FieldRepeatedCount(fieldDef),
	})

	return lineOp_none
}

// 多表合并时, 要从从表的字段名在主表的表头里做索引
func fieldDefGetter(index int, dataHeader, parentHeader *DataHeader) (*model.FieldDescriptor, bool) {

	fieldDef := dataHeader.RawField(index)
	if fieldDef == nil {
		return nil, true
	}

	if parentHeader != nil {

		if strings.Index(fieldDef.Name, "#") == 0 {
			return fieldDef, true
		}

		ret, ok := parentHeader.HeaderByName[fieldDef.Name]
		if !ok {
			return nil, false
		}
		return ret, true
	}

	return fieldDef, true

}

func mustFillCheck(fd *model.FieldDescriptor, raw string) bool {
	// 值重复检查
	if fd.Meta.GetBool("MustFill") {

		if raw == "" {
			log.Errorf("%s, %s", i18n.String(i18n.DataSheet_MustFill), fd.String())
			return false
		}
	}

	return true
}

func newDataSheet(sheet *Sheet) *DataSheet {

	return &DataSheet{
		Sheet: sheet,
	}
}
