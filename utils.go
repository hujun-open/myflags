package myflags

import (
	"fmt"
	"reflect"
)

// PrettyStruc returns a pretty formatted string representation of in
func PrettyStruc(in any, prefix string) string {
	inT := reflect.TypeOf(in)
	inV := reflect.ValueOf(in)
	if !inV.IsValid() {
		return "nnil"
	}
	if inT.Kind() == reflect.Pointer {
		inT = inT.Elem()
		inV = inV.Elem()
	}
	if !inV.IsValid() {
		return "nnil"
	}
	if r, ok := inV.Interface().(fmt.Stringer); ok {
		return r.String()
	}
	switch inT.Kind() {
	case reflect.Array, reflect.Slice:
		rs := ""
		for i := 0; i < inV.Len(); i++ {
			rs += fmt.Sprintf("%v,", PrettyStruc(inV.Index(i).Interface(), prefix))
		}
		return rs
	case reflect.Struct:
		rs := ""
		for i := 0; i < inT.NumField(); i++ {
			fieldV := inV.Field(i)
			if !inT.Field(i).IsExported() {
				continue
			}
			if inT.Field(i).Type.Kind() == reflect.Pointer {
				// fmt.Println("field", inT.Field(i).Name, "is a pointer", inT.Field(i).Type)
				fieldV = fieldV.Elem()
			}
			if !fieldV.IsValid() {
				rs += fmt.Sprintf("%v:nil\n", prefix+inT.Field(i).Name)
			} else {
				if r, ok := fieldV.Interface().(fmt.Stringer); ok {
					// fmt.Println("field", inT.Field(i).Name, "use stringer", fieldV.Type())
					rs += fmt.Sprintf("%v:%v\n", prefix+inT.Field(i).Name, r.String())
				} else {
					if fieldV.Kind() == reflect.Struct {
						rs += fmt.Sprintf("%v:\n%v", prefix+inT.Field(i).Name, PrettyStruc(inV.Field(i).Interface(), prefix+"    "))
					} else {
						// rs += fmt.Sprintf("%v:%v\n", prefix+inT.Field(i).Name, fieldV.Interface())
						rs += fmt.Sprintf("%v:%v\n", prefix+inT.Field(i).Name, PrettyStruc(fieldV.Interface(), prefix+"    "))
					}
				}
			}
		}
		return rs
	}
	return fmt.Sprint(inV.Interface())
}
