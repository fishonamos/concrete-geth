// Copyright 2023 The concrete-geth Authors
//
// The concrete-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The concrete library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the concrete library. If not, see <http://www.gnu.org/licenses/>.

package datamod

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ValueType = iota
	BytesType
	TableType
)

type FieldType struct {
	Name       string
	Type       int
	Size       int
	GoType     string
	SolType    string
	EncodeFunc string
	DecodeFunc string
}

func nameToFieldType(name string) (FieldType, error) {
	switch name {
	case "address":
		return FieldType{
			Name:       "address",
			Size:       20,
			GoType:     "common.Address",
			SolType:    "address",
			EncodeFunc: "EncodeAddress",
			DecodeFunc: "DecodeAddress",
		}, nil
	case "bool":
		return FieldType{
			Name:       "bool",
			Size:       1,
			GoType:     "bool",
			SolType:    "bool",
			EncodeFunc: "EncodeBool",
			DecodeFunc: "DecodeBool",
		}, nil
	case "uint":
		break
	case "int":
		break
	case "bytes":
		return FieldType{
			Name:       "bytes",
			Size:       32,
			GoType:     "[]byte",
			SolType:    "memory bytes",
			EncodeFunc: "EncodeBytes",
			DecodeFunc: "DecodeBytes",
			Type:       BytesType,
		}, nil
	case "string":
		return FieldType{
			Name:       "string",
			Size:       32,
			GoType:     "string",
			SolType:    "memory string",
			EncodeFunc: "EncodeString",
			DecodeFunc: "DecodeString",
			Type:       BytesType,
		}, nil
	default:
	}

	var sizeStr string
	var size int
	var err error

	if strings.HasPrefix(name, "bytes") {
		sizeStr = strings.TrimPrefix(name, "bytes")
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			return FieldType{}, err
		}
		if size < 1 || size > 32 {
			return FieldType{}, fmt.Errorf("invalid bytes size %d", size)
		}
		fieldType := FieldType{
			Name:       name,
			Size:       size,
			GoType:     "[]byte",
			SolType:    fmt.Sprintf("bytes%d", size),
			EncodeFunc: "EncodeFixedBytes",
			DecodeFunc: "DecodeFixedBytes",
		}
		if size == 32 {
			fieldType.GoType = "common.Hash"
			fieldType.EncodeFunc = "EncodeHash"
			fieldType.DecodeFunc = "DecodeHash"
		}
		return fieldType, nil
	}

	matchesUint := strings.HasPrefix(name, "uint")
	matchesInt := strings.HasPrefix(name, "int")

	if matchesUint || matchesInt {
		var noSizeTypeStr string
		if matchesUint {
			noSizeTypeStr = "uint"
		} else {
			noSizeTypeStr = "int"
		}

		sizeStr = strings.TrimPrefix(name, noSizeTypeStr)
		if sizeStr == "" {
			size = 256
		} else {
			size, err = strconv.Atoi(sizeStr)
			if err != nil {
				return FieldType{}, err
			}
		}
		if size < 8 || (size > 64 && size != 256) || size%8 != 0 {
			return FieldType{}, fmt.Errorf("invalid integer size %d", size)
		}

		if noSizeTypeStr == "int" && size == 256 {
			return FieldType{}, fmt.Errorf("int256 is not valid, use uint256 with Sign() and Neg() to represent negative big numbers")
		}

		fieldType := FieldType{
			Name: name,
			Size: size / 8,
		}
		var (
			goType      string
			codecSuffix string
		)
		if size <= 64 {
			goType = noSizeTypeStr + fmt.Sprint(size)
			codecSuffix = fmt.Sprintf("%s%d", upperFirstLetter(noSizeTypeStr), size)
		} else {
			goType = "*uint256.Int"
			codecSuffix = fmt.Sprintf("%s256", upperFirstLetter(noSizeTypeStr))
		}
		fieldType.GoType = goType
		fieldType.SolType = fmt.Sprintf("%s%d", noSizeTypeStr, size)
		fieldType.EncodeFunc = "Encode" + codecSuffix
		fieldType.DecodeFunc = "Decode" + codecSuffix
		return fieldType, nil
	}

	if strings.HasPrefix(name, "table ") {
		tableName := strings.TrimPrefix(name, "table ")
		if !isValidName(tableName) {
			return FieldType{}, fmt.Errorf("invalid table name %s", tableName)
		}
		return FieldType{
			Name:    tableName,
			Size:    32,
			GoType:  formatTableName(tableName),
			SolType: fmt.Sprintf("memory %s", formatTableName(tableName)),
			Type:    TableType,
		}, nil
	}

	return FieldType{}, fmt.Errorf("unknown field type %s", name)
}
