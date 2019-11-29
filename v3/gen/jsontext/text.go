package jsontext

// 报错行号+3
const templateText = `{
	"@Tool": "github.com/davyxu/tabtoy",
	"@Version": "{{.Version}}",	{{range $di, $tab := .Datas.AllTables}}
	"{{$tab.HeaderType}}":[ {{range $unusedrow,$row := $tab.DataRowIndex}}{{$headers := $.Types.AllFieldByName $tab.OriginalHeaderType}}
		{ {{range $col, $header := $headers}}{{$isEmpty := IsEmptyTabValue $ $tab $headers $row $col}}{{if not $isEmpty}}"{{$header.FieldName}}": {{WrapTabValue $ $tab $headers $row $col}}{{GenJsonTailComma $col $headers}} {{end}}{{end}}}{{GenJsonTailComma $row $tab.Rows}}{{end}} 
	]{{GenJsonTailComma $di $.Datas.AllTables}}{{end}}
}`
