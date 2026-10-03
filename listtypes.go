package myflags

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Registered so flags, nouns, and slices of these types share one converter.
// The types package registers the integer kinds again, which replaces these
// and adds the base tag.
func init() {
	Register[string](new(strType))
	Register[float32](&floatType{len: 32})
	Register[float64](&floatType{len: 64})
	Register[bool](new(boolType))
	Register[int](&intType{bits: 0})
	Register[int8](&intType{bits: 8})
	Register[int16](&intType{bits: 16})
	Register[int32](&intType{bits: 32})
	Register[int64](&intType{bits: 64})
	Register[uint](&intType{bits: 0, unsigned: true})
	Register[uint8](&intType{bits: 8, unsigned: true})
	Register[uint16](&intType{bits: 16, unsigned: true})
	Register[uint32](&intType{bits: 32, unsigned: true})
	Register[uint64](&intType{bits: 64, unsigned: true})
}

type strType string

func (s *strType) ToStr(in any, tag reflect.StructTag) string {
	if reflect.ValueOf(in).Kind() == reflect.Pointer {
		return *(in.(*string))
	}
	return in.(string)

}

func (s *strType) FromStr(input string, tag reflect.StructTag) (any, error) {
	return input, nil
}

type boolType bool

func (b *boolType) ToStr(in any, tag reflect.StructTag) string {
	if reflect.ValueOf(in).Kind() == reflect.Pointer {
		return fmt.Sprint(*(in.(*bool)))
	}
	return fmt.Sprint(in)
}
func (b *boolType) FromStr(input string, tag reflect.StructTag) (any, error) {
	return strconv.ParseBool(input)
}

type floatType struct {
	len int
}

func (f *floatType) ToStr(in any, tag reflect.StructTag) string {
	//if in is a pointer, convert it to the value
	val := reflect.ValueOf(in)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	return fmt.Sprint(val.Interface())
}
func (f *floatType) FromStr(s string, tag reflect.StructTag) (any, error) {
	f64, err := strconv.ParseFloat(s, f.len)
	if err != nil {
		return 0, err
	}
	switch f.len {
	case 32:
		return float32(f64), nil

	}
	return f64, nil
}

type intType struct {
	bits     int
	unsigned bool
}

func (i *intType) ToStr(in any, tag reflect.StructTag) string {
	val := reflect.ValueOf(in)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}
	return fmt.Sprint(val.Interface())
}

func (i *intType) FromStr(s string, tag reflect.StructTag) (any, error) {
	text := strings.TrimSpace(s)
	if !i.unsigned {
		n, err := strconv.ParseInt(text, 0, i.bits)
		if err != nil {
			return nil, err
		}
		switch i.bits {
		case 0:
			return int(n), nil
		case 8:
			return int8(n), nil
		case 16:
			return int16(n), nil
		case 32:
			return int32(n), nil
		case 64:
			return int64(n), nil
		}
	} else {
		n, err := strconv.ParseUint(text, 0, i.bits)
		if err != nil {
			return nil, err
		}
		switch i.bits {
		case 0:
			return uint(n), nil
		case 8:
			return uint8(n), nil
		case 16:
			return uint16(n), nil
		case 32:
			return uint32(n), nil
		case 64:
			return uint64(n), nil
		}
	}
	return nil, fmt.Errorf("not a supported type")
}
