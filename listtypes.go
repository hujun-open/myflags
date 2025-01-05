package myflags

import (
	"fmt"
	"reflect"
	"strconv"
)

// these are needed to support slice/array of the types
func init() {
	Register[string](new(strType))
	Register[float32](&floatType{len: 32})
	Register[float64](&floatType{len: 64})
	Register[bool](new(boolType))
}

type strType string

func (s *strType) ToStr(in any, tag reflect.StructTag) string {
	return in.(string)
}

func (s *strType) FromStr(input string, tag reflect.StructTag) (any, error) {
	return input, nil
}

type boolType bool

func (b *boolType) ToStr(in any, tag reflect.StructTag) string {
	return fmt.Sprint(in)
}
func (b *boolType) FromStr(input string, tag reflect.StructTag) (any, error) {
	return strconv.ParseBool(input)
}

type floatType struct {
	len int
}

func (f *floatType) ToStr(in any, tag reflect.StructTag) string {
	return fmt.Sprint(in)
}
func (f *floatType) FromStr(s string, tag reflect.StructTag) (any, error) {
	f64, err := strconv.ParseFloat(s, f.len)
	if err != nil {
		return 0, nil
	}
	switch f.len {
	case 32:
		return float32(f64), nil

	}
	return f64, nil
}
