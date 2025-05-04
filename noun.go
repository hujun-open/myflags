package myflags

import (
	"encoding"
	"reflect"

	"github.com/hujun-open/cobra"
)

func (filler *Filler) parseNoun(args []string) error {
	if len(args) == 0 {
		args = []string{""}
	}

	inV := filler.nounVal

	if filler.nounVal.Kind() != reflect.Pointer {
		inV = inV.Addr()
	}
	//check if it is a registered type, a.k.a simpleType
	conv := globalRegistry.GetViaInterface(inV.Interface())
	if conv != nil {
		val, err := conv.FromStr(args[0], filler.nounTag)
		if err != nil {
			return err
		}
		inV.Elem().Set(reflect.ValueOf(val))
		return nil

	}
	//check if it implements UnmarshalText
	if txtUnmarshal, ok := inV.Interface().(encoding.TextUnmarshaler); ok {
		return txtUnmarshal.UnmarshalText([]byte(args[0]))

	}
	//check if it is arrary/slice
	switch filler.nounVal.Kind() {
	case reflect.Slice, reflect.Array:
		process := false
		if globalRegistry.GetViaType(filler.nounVal.Type().Elem()) != nil {
			process = true
		}

		if filler.nounVal.Type().Elem().Kind() == reflect.Pointer {
			if filler.nounVal.Type().Elem().Implements(textEncodingInt) {
				//list of pointer to textmarshalce
				process = true
			}
		} else {
			if reflect.PointerTo(filler.nounVal.Type().Elem()).Implements(textEncodingInt) {
				//list of textmarshalce
				process = true
			}
		}
		if process {
			newlist, err := getListType(inV, filler.nounTag)
			if err != nil {
				return err
			}
			return newlist.Set(args[0])
		}

	}

	return nil
}

// SetPreRun set f as filler.Commmand.PreRun method;
// don't set filler.Command.PreRun directly
func (filler *Filler) SetPreRun(f func(cmd *cobra.Command, args []string)) {
	newf := func(cmd *cobra.Command, args []string) {
		if filler.nounVal.IsValid() {
			filler.parseNoun(args)
		}
		f(cmd, args)
	}
	filler.PreRun = newf
}

// SetPreRun set f as filler.Commmand.PreRunE method;
// don't set filler.Command.PreRunE directly
func (filler *Filler) SetPreRunE(f func(cmd *cobra.Command, args []string) error) {
	newf := func(cmd *cobra.Command, args []string) error {
		if filler.nounVal.IsValid() {
			err := filler.parseNoun(args)
			if err != nil {
				return err
			}
		}
		return f(cmd, args)
	}
	filler.PreRunE = newf
}
