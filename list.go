package myflags

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// this is to support slice and array
type listType struct {
	val  reflect.Value //need to be a pointer to slice/array
	tags reflect.StructTag
	conv RegisteredConverters //converter for the element
}

func (list *listType) String() string {
	// fmt.Println("type", typeToStr(slice.val.Interface()))
	r := ""
	if !list.val.IsValid() {
		return ""
	}
	if list.val.Elem().Len() == 0 {
		return r
	}
	for i := 0; i < list.val.Elem().Len(); i++ {
		if list.val.Elem().Index(i).Kind() == reflect.Pointer {
			if list.val.Elem().Index(i).IsNil() {
				r += ","
				continue
			}
		}
		r += list.conv.ToStr(list.val.Elem().Index(i).Interface(), list.tags) + ","

	}
	return r[:len(r)-1]
}

func (list *listType) Set(s string) error {
	//check if the slice's element is pointer
	isElmPointer := list.val.Type().Elem().Elem().Kind() == reflect.Pointer
	isArray := list.val.Type().Elem().Kind() == reflect.Array
	list.val.Elem().SetZero()
	for i, ns := range strings.Split(s, ",") {
		n, err := list.conv.FromStr(ns, list.tags)
		if err != nil {
			return err
		}
		if !isArray {
			//slice
			if !isElmPointer {
				list.val.Elem().Set(reflect.Append(list.val.Elem(), reflect.ValueOf(n)))
			} else {
				newval := reflect.New(list.val.Type().Elem().Elem().Elem())
				newval.Elem().Set(reflect.ValueOf(n))
				list.val.Elem().Set(reflect.Append(list.val.Elem(), newval))
			}
		} else {
			//array
			if !isElmPointer {
				list.val.Elem().Index(i).Set(reflect.ValueOf(n))
			} else {
				newval := reflect.New(list.val.Type().Elem().Elem().Elem())
				newval.Elem().Set(reflect.ValueOf(n))
				list.val.Elem().Index(i).Set(newval)
			}

		}
	}
	return nil
}

func processList(fs *flag.FlagSet, ref reflect.Value, tag reflect.StructTag, name, usage string) error {
	rconv := globalRegistry.GetViaType(ref.Type().Elem().Elem())
	var newval listType
	if rconv == nil {
		var unm reflect.Value
		if ref.Type().Elem().Elem().Implements(textEncodingInt) {
			//list of pointer to textmarshalce
			unm = reflect.New(ref.Type().Elem().Elem().Elem())

		} else {
			if reflect.PointerTo(ref.Type().Elem().Elem()).Implements(textEncodingInt) {
				//list of textmarshalce
				unm = reflect.New(ref.Type().Elem().Elem())
			}
		}
		if unm.IsValid() {
			newval = listType{val: ref, tags: tag,
				conv: &textMarshalConverter{unmarshaller: unm.Interface().(encoding.TextUnmarshaler)}}
		} else {
			return fmt.Errorf("%v is not registered", ref.Type().Elem().Elem())
		}

	} else {
		newval = listType{val: ref, tags: tag, conv: rconv}
	}
	fs.Var(&newval, name, usage)
	return nil
}
