package model

type FieldType struct {
	InputFieldName string `tb_name:"输入字段"`
	GoFieldName    string `tb_name:"Go字段"`
	CSFieldName    string `tb_name:"C#字段"`
	ProtoFieldName string `tb_name:"Proto字段"`
	DefaultValue   string `tb_name:"默认值"`
}

// 将表中输入的字段类型转换为各种语言类型

var (
	FieldTypes = []*FieldType{
		{"int16", "int16", "Int16", "int32", "0"},
		{"int32", "int32", "Int32", "int32", "0"},
		{"int64", "int64", "Int64", "int64", "0"},
		{"int", "int32", "Int32", "int32", "0"},
		{"uint16", "uint16", "UInt16", "uint32", "0"},
		{"uint32", "uint32", "UInt32", "uint32", "0"},
		{"uint64", "uint64", "UInt64", "uint64", "0"},
		{"float", "float32", "float", "float", "0"},
		{"double", "float64", "double", "double", "0"},
		{"float32", "float32", "float", "float64", "0"},
		{"float64", "float64", "double", "float64", "0"},
		{"bool", "bool", "bool", "bool", "FALSE"},
		{"string", "string", "string", "string", ""},
	}

	fieldTypeDefaultValues map[string]string
	goFileds               map[string]string
	csFields               map[string]string
	protoFields            map[string]string
)

func InitFieldTypes() {
	fieldTypeDefaultValues = make(map[string]string)
	goFileds = make(map[string]string)
	csFields = make(map[string]string)
	protoFields = make(map[string]string)
	for _, v := range FieldTypes {
		fieldTypeDefaultValues[v.InputFieldName] = v.DefaultValue
		goFileds[v.InputFieldName] = v.GoFieldName
		csFields[v.InputFieldName] = v.CSFieldName
		protoFields[v.InputFieldName] = v.ProtoFieldName
	}
}

// 取类型的默认值
func FetchDefaultValue(fieldType string) (ret string) {
	//linq.From(FieldTypes).WhereT(func(ft *FieldType) bool {
	//
	//	return ft.InputFieldName == fieldType
	//}).ForEachT(func(ft *FieldType) {
	//
	//	ret = ft.DefaultValue
	//})

	ret, _ = fieldTypeDefaultValues[fieldType]

	return
}

// 将类型转为对应语言的原始类型
func LanguagePrimitive(fieldType string, lanType string) string {

	var convertedType string
	//linq.From(FieldTypes).WhereT(func(ft *FieldType) bool {
	//
	//	return ft.InputFieldName == fieldType
	//}).SelectT(func(ft *FieldType) string {
	//
	//	switch lanType {
	//	case "cs":
	//		return ft.CSFieldName
	//	case "go":
	//		return ft.GoFieldName
	//	case "proto":
	//		return ft.ProtoFieldName
	//	default:
	//		panic("unknown lan type: " + lanType)
	//	}
	//}).ForEachT(func(typeName string) {
	//
	//	convertedType = typeName
	//})

	switch lanType {
	case "cs":
		convertedType, _ = csFields[fieldType]
	case "go":
		convertedType, _ = goFileds[fieldType]
	case "proto":
		convertedType, _ = protoFields[fieldType]
	default:
		panic("unknown lan type: " + lanType)
	}

	if convertedType == "" {
		convertedType = fieldType
	}

	return convertedType
}

// 原始类型是否存在，例如: int32, int64
func PrimitiveExists(fieldType string) bool {

	//return linq.From(FieldTypes).WhereT(func(ft *FieldType) bool {
	//
	//	return ft.InputFieldName == fieldType
	//}).Count() > 0

	_, exist := fieldTypeDefaultValues[fieldType]
	return exist
}
