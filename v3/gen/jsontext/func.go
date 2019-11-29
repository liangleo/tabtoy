package jsontext

import (
	"text/template"

	"github.com/davyxu/tabtoy/v3/gen"
	"github.com/davyxu/tabtoy/v3/model"
)

var UsefulFunc = template.FuncMap{}

func init() {
	UsefulFunc["WrapTabValue"] = func(globals *model.Globals, dataTable *model.DataTable, allHeaders []*model.TypeDefine, row, col int) string {

		// 找到完整的表头（按完整表头遍历）
		header := allHeaders[col]

		if header == nil {
			return ""
		}

		// 在单元格找到值
		valueCell := dataTable.GetCell(row, col)

		if valueCell != nil {

			return gen.WrapValue(globals, header, valueCell.Value, true)
		} else {
			// 这个表中没有这列数据
			return gen.WrapValue(globals, header, "", true)
		}
	}

	UsefulFunc["IsEmptyTabValue"] = func(globals *model.Globals, dataTable *model.DataTable, allHeaders []*model.TypeDefine, row, col int) bool {

		// 找到完整的表头（按完整表头遍历）
		header := allHeaders[col]

		if header == nil {
			return true
		}

		// 在单元格找到值
		valueCell := dataTable.GetCell(row, col)

		if valueCell != nil {
			return gen.IsEmptyTabValue(globals, header, valueCell.Value)
		}
		return true
	}

}
