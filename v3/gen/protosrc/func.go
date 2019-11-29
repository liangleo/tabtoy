package protosrc

import (
	"text/template"

	"github.com/davyxu/tabtoy/v3/model"
)

var UsefulFunc = template.FuncMap{}

// 将定义用的类型，转换为不同语言对应的复合类型

func init() {
	UsefulFunc["GoType"] = func(tf *model.TypeDefine) string {

		convertedType := model.LanguagePrimitive(tf.FieldType, "proto")

		if tf.IsArray() {
			return "repeated " + convertedType
		}

		return convertedType
	}

	UsefulFunc["ProtoStructIndex"] = func(index int) int {
		return index + 1
	}

}
