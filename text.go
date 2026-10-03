package myflags

import (
	"encoding"
	"reflect"
)

type textMarshalConverter struct {
	elem reflect.Type
}

func (tmc *textMarshalConverter) ToStr(in any, tag reflect.StructTag) string {
	buf, _ := in.(encodingTextMarshaler).MarshalText()
	return string(buf)
}

func (tmc *textMarshalConverter) FromStr(input string, tag reflect.StructTag) (any, error) {
	fresh := reflect.New(tmc.elem)
	err := fresh.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(input))
	if err != nil {
		return nil, err
	}
	return fresh.Elem().Interface(), nil
}
