package myflags

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// encodingTextMarshaler is the interface includes both encoding.TextMarshaler and encoding.TextUnmarshaler
type encodingTextMarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// RenameFunc is the function to rename the flag for a struct field,
// name is the field name, while parent is the parent struct field name
type RenameFunc func(parent, name string) string

// DefaultRenamer is the default renaming function,
// it is strings.ToLower(prefix + name)
func DefaultRenamer(parent, name string) string {
	return strings.ToLower(parent + name)
}

// Filler could fill a flagset with a struct with Fill() method
type Filler struct {
	renamer RenameFunc
}

// FillerOption is an option when creating new Filler
type FillerOption func(filler *Filler)

// WithRenamer returns a FillerOption that specifies the rename function
func WithRenamer(r RenameFunc) FillerOption {
	return func(filler *Filler) {
		filler.renamer = r
	}
}

// NewFiller creates a new Filler,
// optionally, a list of FillerOptions could be specified.
func NewFiller(options ...FillerOption) *Filler {
	r := &Filler{
		renamer: DefaultRenamer,
	}
	for _, o := range options {
		o(r)
	}
	return r
}

var textEncodingInt = reflect.TypeOf((*encodingTextMarshaler)(nil)).Elem()

// Fill fs with struct in
func (filler *Filler) Fill(fs *flag.FlagSet, in any) error {
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return filler.walk(fs, reflect.ValueOf(in), "", "")
	} else {
		return fmt.Errorf("only support a pointer to struct, but got %v", t)
	}

}

func setStandardFlagType(fs *flag.FlagSet, ref reflect.Value, name, usage string) {
	switch ref.Elem().Kind() {
	case reflect.String:
		casted := ref.Interface().(*string)
		fs.StringVar(casted, name, *casted, usage)
	case reflect.Int:
		casted := ref.Interface().(*int)
		fs.IntVar(casted, name, *casted, usage)
	case reflect.Uint:
		casted := ref.Interface().(*uint)
		fs.UintVar(casted, name, *casted, usage)
	case reflect.Bool:
		casted := ref.Interface().(*bool)
		fs.BoolVar(casted, name, *casted, usage)
	case reflect.Int64:
		casted := ref.Interface().(*int64)
		fs.Int64Var(casted, name, *casted, usage)
	case reflect.Uint64:
		casted := ref.Interface().(*uint64)
		fs.Uint64Var(casted, name, *casted, usage)
	case reflect.Float64:
		casted := ref.Interface().(*float64)
		fs.Float64Var(casted, name, *casted, usage)
	}
}

func setTextEncodingType(fs *flag.FlagSet, ref reflect.Value, name, usage string) {
	casted := ref.Interface().(encodingTextMarshaler)
	fs.TextVar(casted, name, casted, usage) //This requires go 1.19+
}

func isFlagSupportedKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float64, reflect.Int,
		reflect.Int64,
		reflect.String,
		reflect.Bool, reflect.Uint,
		reflect.Uint64:
		return true
	}
	return false
}

// in must be a pointer to struct
func (filler *Filler) walk(fs *flag.FlagSet, inV reflect.Value, nameprefix, usage string) error {
	var err error
	if inV.Kind() != reflect.Pointer {
		inV = inV.Addr()
	}
	inT := inV.Type()
	if inV.IsNil() {
		//if inV is a nil pointer, initialize it
		inV.Set(reflect.New(inT.Elem()))

	}
	ElemK := inV.Elem().Kind()

	//check if it implements EncodingTextMarshaler inteface
	if inT.Implements(textEncodingInt) {
		setTextEncodingType(fs, inV, nameprefix, usage)
		return nil
	}
	//these are kinds directly supported by flag module
	if isFlagSupportedKind(ElemK) {
		setStandardFlagType(fs, inV, nameprefix, usage)
		return nil
	}
	switch ElemK {
	case reflect.Struct:
		//a struct
		for i := 0; i < inV.Elem().NumField(); i++ {
			// fmt.Println("walk into ", inT.Elem().Field(i).Name, inT.Elem().Field(i).Type)
			field := inV.Elem().Field(i)
			fieldT := inT.Elem().Field(i)
			if fieldT.IsExported() {
				//only handle exported field
				//get usage
				usage, _ := fieldT.Tag.Lookup("usage")
				fname := fieldT.Name
				fname = filler.renamer(nameprefix, fname)
				alias, _ := fieldT.Tag.Lookup("alias")
				if alias != "" {
					fname = alias
				}
				if field.Kind() == reflect.Pointer {
					if field.IsNil() {
						//initilize the nil pointer
						field.Set(reflect.New(fieldT.Type.Elem()))
					}
				}
				//check if it is a registered type
				f := getFactory(field.Interface())
				if f != nil {
					if field.Kind() != reflect.Pointer {
						field = field.Addr()
					}
					f(fs, field, fieldT.Tag, fname, usage)
					continue
				}
				//check if it implements textMarshal
				if fieldT.Type.Kind() == reflect.Pointer {
					if fieldT.Type.Implements(textEncodingInt) {
						//pointer to textmarshale
						setTextEncodingType(fs, field.Addr(), fname, usage)
						continue
					}
				} else {
					if reflect.PointerTo(fieldT.Type).Implements(textEncodingInt) {
						//textmarshale
						setTextEncodingType(fs, field.Addr(), fname, usage)
						continue
					}
				}

				//check if it is a slice/array of registered types
				switch fieldT.Type.Kind() {
				case reflect.Slice, reflect.Array:
					process := false
					if globalRegistry.GetViaType(fieldT.Type.Elem()) != nil {
						process = true
					}

					if fieldT.Type.Elem().Kind() == reflect.Pointer {
						if fieldT.Type.Elem().Implements(textEncodingInt) {
							//list of pointer to textmarshalce
							process = true
						}
					} else {
						if reflect.PointerTo(fieldT.Type.Elem()).Implements(textEncodingInt) {
							//list of textmarshalce
							process = true
						}
					}
					if process {
						err = processList(fs, field.Addr(), fieldT.Tag, fname, usage)
						if err != nil {
							return err
						}
						continue
					} else {
						return fmt.Errorf("%v is a slice/array of unsupported type %v", fieldT.Name, fieldT)
					}

				}
				err = filler.walk(fs, field, fname, usage)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
