package model

import (
	"github.com/davyxu/golexer"
	"github.com/davyxu/tabtoy/v3/helper"
	"github.com/davyxu/tabtoy/v3/model"
	"github.com/tealeg/xlsx"
	"path/filepath"
)

type Globals struct {
	TableGetter helper.FileGetter

	SourceTypes []ObjectFieldType

	SourceFileList []string

	TargetTypesSheet helper.TableSheet

	TargetIndexSheet helper.TableSheet

	TargetTables *helper.MemFile

	SourceFileMetas map[string]*golexer.KVPair

	OutputDir        string
	OutputFileSuffix string
}

func (self *Globals) AddTableByFile(tableFileName, tableName string, inputFile *xlsx.File) {

	file := helper.NewXlsxFile()

	file.(interface {
		FromXFile(file *xlsx.File)
	}).FromXFile(inputFile)

	tableFileName = filepath.Base(tableFileName)

	self.TargetTables.AddFile(tableFileName, file).TableName = tableName
}

func (self *Globals) AddTable(tableFileName string) helper.TableSheet {

	return self.TargetTables.CreateXLSXFile(tableFileName)
}

func (self *Globals) SourceTypeExists(objectTypeName, fieldName string) bool {
	for _, ft := range self.SourceTypes {

		if ft.ObjectType == objectTypeName && ft.FieldName == fieldName {
			return true
		}
	}

	return false
}

func (self *Globals) ObjectTypeByObjectType(objectType string) *ObjectFieldType {
	for _, ft := range self.SourceTypes {

		if ft.ObjectType == objectType {
			return &ft
		}
	}

	return nil
}

func (self *Globals) ObjectTypeByObjectTypeAndFieldName(objectType string, fieldName string) *ObjectFieldType {
	for _, ft := range self.SourceTypes {

		if ft.ObjectType == objectType && ft.FieldName == fieldName {
			return &ft
		}
	}

	return nil
}

func (self *Globals) AddSourceType(oft ObjectFieldType) {

	self.SourceTypes = append(self.SourceTypes, oft)
}

func (self *Globals) PrintTypes() {

	for _, ft := range self.SourceTypes {

		log.Debugf("%+v", ft)
	}
}

func (self *Globals) TypeIsNoneKind(objectTypeName string) bool {
	for _, oft := range self.SourceTypes {
		if oft.ObjectType == objectTypeName && oft.Kind == model.TypeUsage_None {
			return true
		}
	}

	return false
}

func (self *Globals) AddSourceFileMeta(tableName string, meta *golexer.KVPair) {
	self.SourceFileMetas[tableName] = meta
}

func (self *Globals) SourceFileMetaByName(tableName string) (*golexer.KVPair, bool) {
	meta, exist := self.SourceFileMetas[tableName]
	return meta, exist
}

func (self *Globals) SourceFileMetaExists(tableName string) bool {
	_, exist := self.SourceFileMetas[tableName]
	return exist
}

func NewGlobals() *Globals {

	return &Globals{
		TargetTables: helper.NewMemFile(),
	}
}
