package printer

import (
	"text/template"

	"github.com/davyxu/tabtoy/v2/i18n"
	"github.com/davyxu/tabtoy/v2/model"
)

// TODO pbmeta解析换rune的lexer [tabtoy] {{.Comment}}
const protoTemplate = `// Generated by github.com/davyxu/tabtoy
// Version: {{.ToolVersion}}
// DO NOT EDIT!!
{{if ge .ProtoVersion 3}}
syntax = "proto3";
{{end}}
package {{.Package}};
{{range .Imports}}
import "{{.}}";
{{end}}
{{range .Enums}}
// Defined in table: {{.DefinedTable}}
enum {{.Name}}
{	
{{range .ProtoFields}}
	{{.Alias}}
	{{.Name}} = {{.Number}}; {{.Comment}}
{{end}}
}
{{end}}
{{range .Messages}}
// Defined in table: {{.DefinedTable}}
message {{.Name}}
{	
{{range .ProtoFields}}	
	{{.Alias}}
	{{.Label}}{{.TypeString}} {{.Name}} = {{.Number}}; {{.Comment}}
{{end}}
}
{{end}}
`

type protoFieldDescriptor struct {
	*model.FieldDescriptor

	d *protoDescriptor

	Number int
}

func (self protoFieldDescriptor) Label() string {
	if self.IsRepeated {
		return "repeated "
	}

	if self.d.file.ProtoVersion == 2 {
		return "optional "
	}

	return ""
}

func (self protoFieldDescriptor) Alias() string {

	if self.FieldDescriptor.Meta.GetString("Alias") == "" {
		return ""
	}

	return "// " + self.FieldDescriptor.Meta.GetString("Alias")
}

func (self protoFieldDescriptor) Comment() string {

	if self.FieldDescriptor.Comment == "" {
		return ""
	}

	return "// " + self.FieldDescriptor.Comment

}

type protoDescriptor struct {
	*model.Descriptor

	ProtoFields []protoFieldDescriptor

	file *protoFileModel
}

func (self *protoDescriptor) DefinedTable() string {
	return self.File.Name
}

type protoFileModel struct {
	Package      string
	ProtoVersion int
	ToolVersion  string
	Messages     []protoDescriptor
	Enums        []protoDescriptor
	Imports      []string
	IgnoreFiles  []string
}

func (self *protoFileModel) isIgnoreFile(name string) bool {
	for _, v := range self.IgnoreFiles {
		if v == name {
			return true
		}
	}
	return false
}

type protoPrinter struct {
}

func (self *protoPrinter) Run(g *Globals) *Stream {

	tpl, err := template.New("proto").Parse(protoTemplate)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	var m protoFileModel

	m.Package = g.FileDescriptor.Pragma.GetString("Package")
	m.ProtoVersion = g.ProtoVersion
	m.ToolVersion = g.Version
	m.Imports = g.ProtoImportFiles
	m.IgnoreFiles = g.ProtoOutputIgnoreFiles

	// 遍历所有类型
	for _, d := range g.FileDescriptor.Descriptors {

		// 这个文件被限制输出
		if m.isIgnoreFile(d.File.Name) {
			log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), d.Name)
			continue
		}

		// 这个文件被限制输出
		if !d.File.MatchTag(".proto") {
			log.Infof("%s: %s", i18n.String(i18n.Printer_IgnoredByOutputTag), d.Name)
			continue
		}

		var protoD protoDescriptor
		protoD.Descriptor = d
		protoD.file = &m

		// 遍历字段
		delta := 0
		for index, fd := range d.Fields {

			if fd.Meta.GetBool("Memo") {
				delta += 1
				continue
			}

			// 对CombineStruct的XXDefine对应的字段
			if d.Usage == model.DescriptorUsage_CombineStruct {
				// 这个字段被限制输出
				if fd.Complex != nil && !fd.Complex.File.MatchTag(".proto") {
					continue
				}
			}

			var field protoFieldDescriptor
			field.FieldDescriptor = fd
			field.d = &protoD

			switch d.Kind {
			case model.DescriptorKind_Struct:
				field.Number = index + 1 - delta
			case model.DescriptorKind_Enum:
				field.Number = int(fd.EnumValue)
			}

			protoD.ProtoFields = append(protoD.ProtoFields, field)

		}

		switch d.Kind {
		case model.DescriptorKind_Struct:
			m.Messages = append(m.Messages, protoD)
		case model.DescriptorKind_Enum:
			m.Enums = append(m.Enums, protoD)
		}

	}

	bf := NewStream()

	err = tpl.Execute(bf.Buffer(), &m)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	return bf
}

func init() {

	RegisterPrinter("proto", &protoPrinter{})

}
