package model

import (
	"encoding/json"
	"fmt"

	"github.com/ahmetb/go-linq"
)

type TypeData struct {
	Define *TypeDefine
	Tab    *DataTable // 类型引用的表
	Row    int        // 类型引用的原始数据(DataTable)中的行
}

type TypeTable struct {
	fields            []*TypeData
	objectType2Fields map[string][]*TypeDefine // objectType -> TypeDefines
	enumFields        map[string][]*TypeDefine // objectType -> TypeDefines
}

func (self *TypeTable) ToJSON(all bool) []byte {

	data, _ := json.MarshalIndent(self.AllFields(all), "", "\t")

	return data
}

func (self *TypeTable) Print(all bool) {

	fmt.Println(string(self.ToJSON(all)))
}

// refData，类型表对应源表的位置信息
func (self *TypeTable) AddField(tf *TypeDefine, data *DataTable, row int) {

	if self.FieldByName(tf.ObjectType, tf.FieldName) != nil {
		panic("Duplicate table field: " + tf.FieldName)
	}

	typeData := &TypeData{
		Tab:    data,
		Define: tf,
		Row:    row,
	}
	self.fields = append(self.fields, typeData)

	if self.objectType2Fields == nil {
		self.objectType2Fields = make(map[string][]*TypeDefine)
	}
	if self.objectType2Fields[tf.ObjectType] == nil {
		self.objectType2Fields[tf.ObjectType] = []*TypeDefine{}
	}
	self.objectType2Fields[tf.ObjectType] = append(self.objectType2Fields[tf.ObjectType], tf)

	if tf.Kind == TypeUsage_Enum {
		if self.enumFields == nil {
			self.enumFields = make(map[string][]*TypeDefine)
		}
		if self.enumFields[tf.ObjectType] == nil {
			self.enumFields[tf.ObjectType] = []*TypeDefine{}
		}
		self.enumFields[tf.ObjectType] = append(self.enumFields[tf.ObjectType], tf)
	}
}

func (self *TypeTable) Raw() []*TypeData {
	return self.fields
}

func (self *TypeTable) AllFields(all bool) (ret []*TypeDefine) {

	linq.From(self.fields).WhereT(func(td *TypeData) bool {

		if !all && td.Define.IsBuiltin {
			return false
		}

		return true
	}).SelectT(func(td *TypeData) interface{} {

		return td.Define
	}).ToSlice(&ret)

	return
}

// 类型是枚举
func (self *TypeTable) IsEnumKind(objectType string) bool {
	_, exist := self.enumFields[objectType]
	return exist
	//return linq.From(self.rawEnumNames(true)).WhereT(func(name string) bool {
	//	return name == objectType
	//}).Count() == 1
}

// 匹配枚举值
func (self *TypeTable) ResolveEnumValue(objectType, value string) string {

	//enumFields := self.getEnumFields(objectType)
	enumFields, exist := self.enumFields[objectType]

	if !exist || len(enumFields) == 0 {
		return ""
	}

	for _, td := range enumFields {

		if td.Name == value || td.FieldName == value {
			return td.Value
		}

	}

	// 默认取第一个
	return enumFields[0].Value
}

func (self *TypeTable) getEnumFields(objectType string) (ret []*TypeData) {

	for _, td := range self.fields {

		if td.Define.ObjectType == objectType {
			ret = append(ret, td)
		}

	}

	return
}

func (self *TypeTable) EnumNames() (ret []string) {

	return self.rawEnumNames(BuiltinSymbolsVisible)
}

func (self *TypeTable) StructNames() (ret []string) {

	return self.rawStructNames(BuiltinSymbolsVisible)
}

// 获取所有的结构体名
func (self *TypeTable) rawStructNames(all bool) (ret []string) {

	linq.From(self.fields).WhereT(func(td *TypeData) bool {

		tf := td.Define

		if !all && tf.IsBuiltin {
			return false
		}

		return tf.Kind == TypeUsage_HeaderStruct
	}).SelectT(func(td *TypeData) string {

		return td.Define.ObjectType
	}).Distinct().ToSlice(&ret)

	return
}

// 获取所有的枚举名
func (self *TypeTable) rawEnumNames(all bool) (ret []string) {

	linq.From(self.fields).WhereT(func(td *TypeData) bool {

		tf := td.Define

		if !all && tf.IsBuiltin {
			return false
		}

		return tf.Kind == TypeUsage_Enum
	}).SelectT(func(td *TypeData) string {

		return td.Define.ObjectType
	}).Distinct().ToSlice(&ret)

	//for objectType, _ := range self.enumFields {
	//	ret = append(ret, objectType)
	//}
	//
	//sort.Strings(ret)

	return
}

func (self *TypeTable) inProtoOutputIgnoreFiles(globals *Globals, name string) bool {
	if len(globals.ProtoOutputIgnoreFiles) == 0 {
		return false
	}
	for _, v := range globals.ProtoOutputIgnoreFiles {
		if name == v {
			return true
		}
	}
	return false
}

func (self *TypeTable) namesPlus(globals *Globals, kind TypeUsage) (ret []string) {
	set := make(map[string]struct{})
	for _, v := range self.fields {
		tf := v.Define
		_, exist := set[tf.ObjectType]
		if exist {
			continue
		}

		if v.Tab != nil && self.inProtoOutputIgnoreFiles(globals, v.Tab.FileName) {
			continue
		}

		if !BuiltinSymbolsVisible && tf.IsBuiltin {
			continue
		}

		if tf.Kind != kind {
			continue
		}

		ret = append(ret, tf.ObjectType)
		set[tf.ObjectType] = struct{}{}
	}
	return
}

func (self *TypeTable) EnumNamesPlus(globals *Globals) []string {
	return self.namesPlus(globals, TypeUsage_Enum)
}

func (self *TypeTable) StructNamesPlus(globals *Globals) []string {
	return self.namesPlus(globals, TypeUsage_HeaderStruct)
}

// 对象的所有字段
func (self *TypeTable) AllFieldByName(objectType string) (ret []*TypeDefine) {

	//linq.From(self.fields).WhereT(func(td *TypeData) bool {
	//
	//	return td.Define.ObjectType == objectType
	//}).SelectT(func(td *TypeData) *TypeDefine {
	//
	//	return td.Define
	//}).ToSlice(&ret)

	ret, _ = self.objectType2Fields[objectType]

	return
}

// 数据表中表头对应类型表
func (self *TypeTable) FieldByName(objectType, name string) (ret *TypeDefine) {

	//linq.From(self.fields).WhereT(func(td *TypeData) bool {
	//
	//	tf := td.Define
	//
	//	return tf.ObjectType == objectType &&
	//		(tf.Name == name || tf.FieldName == name)
	//}).ForEachT(func(td *TypeData) {
	//
	//	ret = td.Define
	//
	//})

	fields, exist := self.objectType2Fields[objectType]
	if !exist {
		return
	}

	for _, v := range fields {
		if v.Name == name || v.FieldName == name {
			return v
		}
	}

	return
}

func (self *TypeTable) ObjectExists(objectType string) bool {

	//return linq.From(self.fields).WhereT(func(td *TypeData) bool {
	//
	//	return td.Define.ObjectType == objectType
	//}).Count() > 0

	_, exist := self.objectType2Fields[objectType]
	return exist
}

func NewSymbolTable() *TypeTable {
	return new(TypeTable)
}

var BuiltinSymbolsVisible bool
