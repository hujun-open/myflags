package myflags

import (
	"encoding"
	"reflect"
)

type textMarshalConverter struct {
	unmarshaller encoding.TextUnmarshaler
}

func (tmc *textMarshalConverter) ToStr(in any, tag reflect.StructTag) string {
	buf, _ := in.(encodingTextMarshaler).MarshalText()
	return string(buf)
}

func (tmc *textMarshalConverter) FromStr(input string, tag reflect.StructTag) (any, error) {
	err := tmc.unmarshaller.UnmarshalText([]byte(input))
	if err != nil {
		return nil, err
	}
	val := reflect.ValueOf(tmc.unmarshaller)
	return val.Elem().Interface(), nil

}
