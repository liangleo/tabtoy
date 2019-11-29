package gen

import (
	"strings"
	"text/template"

	"github.com/ahmetb/go-linq"
	"github.com/davyxu/tabtoy/util"
	"github.com/davyxu/tabtoy/v3/model"
	"github.com/davyxu/tabtoy/v3/report"
)

var UsefulFunc = template.FuncMap{}

type TableIndices struct {
	Table     *model.DataTable
	FieldInfo *model.TypeDefine
}

func KeyValueTypeNames(globals *model.Globals) (ret []string) {
	linq.From(globals.IndexList).WhereT(func(pragma *model.IndexDefine) bool {
		return pragma.Kind == model.TableKind_KeyValue
	}).SelectT(func(pragma *model.IndexDefine) string {

		return pragma.TableType
	}).Distinct().ToSlice(&ret)

	return
}

func WrapValue(globals *model.Globals, valueType *model.TypeDefine, value string, wrapQuote bool) string {
	if valueType.IsArray() {
		var sb strings.Builder
		sb.WriteString("[")
		// 空的单元格，导出空数组，除非强制指定填充默认值
		if value != "" {
			count := 0
			for _, elementValue := range strings.Split(value, valueType.ArraySplitter) {
				if count > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(wrapSingleValue(globals, valueType, elementValue, wrapQuote))
				count++
			}
		}
		sb.WriteString("]")
		return sb.String()
	}
	return wrapSingleValue(globals, valueType, value, wrapQuote)
}

func wrapSingleValue(globals *model.Globals, valueType *model.TypeDefine, value string, wrapQuote bool) string {
	if model.PrimitiveExists(valueType.FieldType) {
		switch valueType.FieldType {
		case "string": // 字符串
			return util.StringEscape(value)
		case "float":
			return value
		case "bool":
			switch value {
			case "是", "yes", "YES", "1", "true", "TRUE", "True":
				return "true"
			case "否", "no", "NO", "0", "false", "FALSE", "False":
				return "false"
			}
			return "false"
		}
		if value == "" {
			return model.FetchDefaultValue(valueType.FieldType)
		}
		return value
	} else if globals.Types.IsEnumKind(valueType.FieldType) { // 枚举
		return globals.Types.ResolveEnumValue(valueType.FieldType, value)
	}
	return parseStructValue(globals, valueType, value, wrapQuote)
}

func parseStructValue(globals *model.Globals, valueType *model.TypeDefine, value string, wrapQuote bool) string {
	var sb strings.Builder
	count := 0
	rawValue := strings.TrimSpace(value)
	sb.WriteString("{ ")
	for _, v := range strings.Split(rawValue, " ") {
		field := strings.Split(v, ":")
		if len(field) != 2 {
			continue
		}
		ret := globals.Types.FieldByName(valueType.FieldType, field[0])
		if ret == nil {
			report.Log.Debugf("fieldType: %s, field: %v", valueType.FieldType, field)
			continue
		}
		if count > 0 {
			sb.WriteString(", ")
		}
		if wrapQuote {
			sb.WriteString("\"")
		}
		sb.WriteString(ret.FieldName)
		if wrapQuote {
			sb.WriteString("\"")
		}
		sb.WriteString(": ")
		sb.WriteString(WrapValue(globals, ret, field[1], wrapQuote))
		count++
	}
	sb.WriteString(" }")
	return sb.String()
}

func IsEmptyTabValue(globals *model.Globals, valueType *model.TypeDefine, value string) bool {
	if value == "" {
		return true
	}
	if model.PrimitiveExists(valueType.FieldType) {
		switch valueType.FieldType {
		case "float":
			if value == "0" {
				return true
			}
		case "bool":
			switch value {
			case "是", "yes", "YES", "1", "true", "TRUE", "True":
				return false
			}
		}
	} else if globals.Types.IsEnumKind(valueType.FieldType) { // 枚举
		value = globals.Types.ResolveEnumValue(valueType.FieldType, value)
		if value == "0" {
			return true
		}
	}
	return false
}

func init() {
	UsefulFunc["HasKeyValueTypes"] = func(globals *model.Globals) bool {
		return len(KeyValueTypeNames(globals)) > 0
	}

	UsefulFunc["GetKeyValueTypeNames"] = KeyValueTypeNames

	UsefulFunc["GetIndices"] = func(globals *model.Globals) (ret []TableIndices) {

		for _, tab := range globals.Datas.AllTables() {

			// 遍历输入数据的每一列
			for _, header := range tab.Headers {

				// 输入的列头
				if header.TypeInfo == nil {
					continue
				}

				if header.TypeInfo.MakeIndex {

					ret = append(ret, TableIndices{
						Table:     tab,
						FieldInfo: header.TypeInfo,
					})
				}
			}
		}

		return

	}

}
