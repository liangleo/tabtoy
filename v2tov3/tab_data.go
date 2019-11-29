package v2tov3

import (
	"strings"

	"github.com/davyxu/golexer"
	"github.com/davyxu/tabtoy/util"
	"github.com/davyxu/tabtoy/v2tov3/model"
	"github.com/davyxu/tabtoy/v3/helper"
	"github.com/tealeg/xlsx"
)

func importDatas(sourceSheet, targetSheet *xlsx.Sheet, headerList []model.ObjectFieldType, tabPragma *golexer.KVPair, fileName string) error {
	isVertical := false
	// 判断是否纵表
	if tabPragma.ContainValue("Vertical", "true") {
		isVertical = true
	}
	if isVertical {
		return setImportDatasVertical(sourceSheet, targetSheet, headerList, fileName)
	}
	return setImportDatas(sourceSheet, targetSheet, headerList, fileName)
}

func setImportDatas(sourceSheet, targetSheet *xlsx.Sheet, headerList []model.ObjectFieldType, fileName string) error {
	var row, col int
	for row = 4; ; row++ {
		if helper.IsFullRowEmpty(sourceSheet, row) {
			break
		}
		rowData := targetSheet.AddRow()
		var header model.ObjectFieldType
		for col, header = range headerList {
			sourceCell := sourceSheet.Cell(row, col)
			sourceCell.Value = strings.TrimSpace(sourceCell.Value)
			targetCell := rowData.AddCell()
			if header.IsArray() {
				targetCell.SetValue(sourceCell.Value)
				continue
			}
			if err := setTargetCell(header.FieldType, sourceCell, targetCell, row, col, fileName); err != nil {
				return err
			}
		}
	}
	return nil
}

func setImportDatasVertical(sourceSheet, targetSheet *xlsx.Sheet, headerList []model.ObjectFieldType, fileName string) error {
	for row := 1; ; row++ {
		if helper.IsFullRowEmpty(sourceSheet, row) {
			break
		}
		rowData := targetSheet.AddRow()

		rowHeader := &model.ObjectFieldType{}
		valueCellList := make(map[int]*xlsx.Cell)
		for col, header := range headerList {
			sourceCell := sourceSheet.Cell(row, col)
			sourceCell.Value = strings.TrimSpace(sourceCell.Value)
			targetCell := rowData.AddCell()
			switch header.FieldName {
			case "字段名":
				rowHeader.FieldName = sourceCell.Value
			case "类型":
				sourceCell.Value = strings.TrimLeft(sourceCell.Value, "[]")
				sourceCell.Value = strings.TrimLeft(sourceCell.Value, "repeated ")
				rowHeader.FieldType = sourceCell.Value
			case "特性":
				if sourceCell.Value == "" {
					continue
				}
				kvpair := golexer.NewKVPair()
				if err := kvpair.Parse(sourceCell.Value); err != nil {
					continue
				}
				sourceCell.Value = kvpair.GetString("ListSpliter")
				rowHeader.ArraySplitter = sourceCell.Value
			case "注释":
				rowHeader.Name = sourceCell.Value
			case "值":
				rowHeader.Value = sourceCell.Value
				valueCellList[col] = targetCell
			default:
				continue
			}
			targetCell.SetValue(sourceCell.Value)
		}
		// convert value by type
		for col, cell := range valueCellList {
			if rowHeader.IsArray() {
				continue
			}
			if err := setTargetCell(rowHeader.FieldType, cell, cell, row, col, fileName); err != nil {
				return err
			}
		}
	}
	return nil
}

func setTargetCell(headerFieldType string, sourceCell, targetCell *xlsx.Cell, row, col int, fileName string) (err error) {
	switch headerFieldType {
	case "int32", "uint32":
		if sourceCell.Value == "" {
			targetCell.SetValue("")
			break
		}

		var v int
		v, err = sourceCell.Int()
		if err != nil {
			goto OnError
		}

		if v == 0 {
			targetCell.SetValue("")
		} else {
			targetCell.SetInt(v)
		}

	case "int64", "uint64":
		if sourceCell.Value == "" {
			targetCell.SetValue("")
			break
		}

		var v int64
		v, err = sourceCell.Int64()
		if err != nil {
			goto OnError
		}

		if v == 0 {
			targetCell.SetValue("")
		} else {
			targetCell.SetInt64(v)
		}

	case "float":
		if sourceCell.Value == "" {
			targetCell.SetFloat(0)
			break
		}

		var v float64
		v, err = sourceCell.Float()
		if err != nil {
			goto OnError
		}

		if v == 0 {
			targetCell.SetValue("")
		} else {
			targetCell.SetFloat(v)
		}

	case "bool":
		var v bool

		if err, _ = util.StringToPrimitive(sourceCell.Value, &v); err != nil {
			goto OnError
		}

		if v {
			targetCell.SetValue("是")
		} else {
			targetCell.SetValue("")
		}

	default:
		targetCell.SetValue(sourceCell.Value)
	}

	return

OnError:
	log.Errorf("单元格转换错误 %s|%s, %s", fileName, util.R1C1ToA1(row+1, col+1), err.Error())

	return
}
