package converters

import (
	"fmt"
	"strings"
)

var supportedTypeToConverterType = map[string]string{
	// protobuf scalar types
	"double":   "Float64",
	"float":    "Float32",
	"int32":    "Int32",
	"int64":    "Int64",
	"uint32":   "UInt32",
	"uint64":   "UInt64",
	"sint32":   "Int32",
	"sint64":   "Int64",
	"fixed32":  "UInt32",
	"fixed64":  "UInt64",
	"sfixed32": "Int32",
	"sfixed64": "Int64",
	"bool":     "Bool",
	"string":   "String",
	"bytes":    "Bytes",

	// protobuf structs
	"google.protobuf.Timestamp":   "Timestamp",
	"google.protobuf.StringValue": "StringValue",
	"google.protobuf.FloatValue":  "FloatValue",
	"google.protobuf.BoolValue":   "BoolValue",
	"google.protobuf.Int32Value":  "Int32Value",
	"google.protobuf.Int64Value":  "Int64Value",
	"google.protobuf.UInt32Value": "UInt32Value",
	"google.protobuf.UInt64Value": "UInt64Value",
	"google.protobuf.DoubleValue": "DoubleValue",
	"google.protobuf.Struct":      "Struct",
	"google.protobuf.Value":       "Value",

	// additional types
	"time.Time":              "Time",
	"float32":                "Float32",
	"float64":                "Float64",
	"int":                    "Int",
	"uint":                   "UInt",
	"*string":                "StringPointer",
	"*time.Time":             "TimePointer",
	"*int32":                 "Int32Pointer",
	"*int64":                 "Int64Pointer",
	"*uint32":                "UInt32Pointer",
	"*uint64":                "UInt64Pointer",
	"*float32":               "Float32Pointer",
	"*float64":               "Float64Pointer",
	"*bool":                  "BoolPointer",
	"map[string]interface{}": "Map",
	"json":                   "Json",
	"interface{}":            "Interface",
}

// Converter is an object to represent a conversion between types.
type Converter struct {
	original string
	output   string
}

func (c *Converter) String() string {
	return c.output
}

func (c *Converter) Original() string {
	return c.original
}

// ConverterType converts a protobuf type (as string) into its respective internal
// supported type.
func ConverterType(protobufType string) (*Converter, error) {
	key := strings.TrimPrefix(protobufType, ".")

	t, ok := supportedTypeToConverterType[key]
	if !ok {
		return nil, fmt.Errorf("unsupported type '%s'", protobufType)
	}

	return &Converter{
		original: protobufType,
		output:   t,
	}, nil
}

var conversionMap = map[string]map[string]bool{
	"String": map[string]bool{
		"int":        true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint32":     true,
		"uint64":     true,
		"float32":    true,
		"float64":    true,
		"bool":       true,
		"time.Time":  true,
		"*time.Time": true,
		"*string":    true,
		"json":       true,
	},
	"Timestamp": map[string]bool{
		"time.Time":  true,
		"*time.Time": true,
	},
	"StringValue": map[string]bool{
		"int":        true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint32":     true,
		"uint64":     true,
		"float32":    true,
		"float64":    true,
		"bool":       true,
		"time.Time":  true,
		"*time.Time": true,
		"string":     true,
		"*string":    true,
	},
	"Int32Value": map[string]bool{
		"int32":  true,
		"*int32": true,
	},
	"Int64Value": map[string]bool{
		"int64":  true,
		"*int64": true,
	},
	"UInt32Value": map[string]bool{
		"uint32":  true,
		"*uint32": true,
	},
	"UInt64Value": map[string]bool{
		"uint64":  true,
		"*uint64": true,
	},
	"FloatValue": map[string]bool{
		"float32":  true,
		"*float32": true,
	},
	"DoubleValue": map[string]bool{
		"float64":  true,
		"*float64": true,
	},
	"BoolValue": map[string]bool{
		"bool":  true,
		"*bool": true,
	},
	"Struct": map[string]bool{
		"map[string]interface{}": true,
	},
	"Bool": map[string]bool{
		"*bool": true,
	},
	"Int32": map[string]bool{
		"*int32": true,
	},
	"Int64": map[string]bool{
		"*int64": true,
	},
	"Float32": map[string]bool{
		"*float32": true,
	},
	"Float64": map[string]bool{
		"*float64": true,
	},
	"UInt32": map[string]bool{
		"*uint32": true,
	},
	"UInt64": map[string]bool{
		"*uint64": true,
	},
	"Value": map[string]bool{
		"interface{}": true,
	},
}

// IsSupportedConversion checks if this package can execute this kind of
// conversion, from in to out. Both in and out must be a valid converter
// type.
//
// Conversion table
//
// Source          | Destination
// ----------------|--------------------------------------------------
// String          | int, int32, int64, uint, uint32, uint64, float32,
//                 | float64, bool, time.Time, *time.Time, *string
// ----------------|--------------------------------------------------
// Timestamp       | time.Time, *time.Time
// ----------------|--------------------------------------------------
// StringValue     | *string, int, int32, int64, uint, uint32
//                 | uint64, float32, float64, bool, time.Time,
//                 | *time.Time
// ----------------|--------------------------------------------------
// Int32Value      | *int32, int32
// ----------------|--------------------------------------------------
// Int64Value      | *int64, int64
// ----------------|--------------------------------------------------
// UInt32Value     | *uint32, uint32
// ----------------|--------------------------------------------------
// UInt64Value     | *uint64, uint64
// ----------------|--------------------------------------------------
// FloatValue      | *float32, float32
// ----------------|--------------------------------------------------
// DoubleValue     | *float64, float64
// ----------------|--------------------------------------------------
// BoolValue       | *bool, bool
// ----------------|--------------------------------------------------
// Struct          | map[string]interface{}
// ----------------|--------------------------------------------------
// bool            | *bool
// ----------------|--------------------------------------------------
// int32           | *int32
// ----------------|--------------------------------------------------
// int64           | *int64
// ----------------|--------------------------------------------------
// uint32          | *uint32
// ----------------|--------------------------------------------------
// uint64          | *uint64
// ----------------|--------------------------------------------------
// float		   | *float
// ----------------|--------------------------------------------------
// double		   | *double
// ----------------|--------------------------------------------------
//
func IsSupportedConversion(from, to *Converter) error {
	v, ok := conversionMap[from.String()]
	if !ok {
		return fmt.Errorf("'%s' is not supported as conversion source", from.String())
	}

	if _, ok := v[to.Original()]; !ok {
		return fmt.Errorf("'%s' type cannot be converted into '%s'", from.String(),
			to.Original())
	}

	return nil
}
