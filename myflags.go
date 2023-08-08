package myflags

import (
	"bytes"
	"encoding"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// encodingTextMarshaler is the interface includes both encoding.TextMarshaler and encoding.TextUnmarshaler
type encodingTextMarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// RenameFunc is the function to rename the flag for a struct field,
// name is the field name, while parent is the parent struct field name,
// isAct is true when parent is an action
type RenameFunc func(parent, name string, isAct bool) string

// DefaultRenamer is the default renaming function,
// it is parent + "-" + name when isAct is true;
// otherwise return lower case of name
func DefaultRenamer(parent, name string, isAct bool) string {
	if parent != "" && !isAct {
		return strings.ToLower(parent + "-" + name)
	}
	return strings.ToLower(name)
}

const (
	//default flag.ErrorHandling
	DefaultErrHandle = flag.ExitOnError
)

// Filler auto-generates one or multiple flag.FlagSet based on an input struct
type Filler struct {
	fsMap                map[string]*Filler //child fillers, key is the action name
	orderList            []string
	fs                   *flag.FlagSet
	errHandle            flag.ErrorHandling
	optList              []FillerOption
	usage                string //this is the usage string for for overall filler
	renamer              RenameFunc
	translatedActNameMap map[string]string //key is the transalted action name, val is the original field name
}

// FillerOption is an option when creating new Filler
type FillerOption func(filler *Filler)

// WithRenamer returns a FillerOption that specifies the rename function
func WithRenamer(r RenameFunc) FillerOption {
	return func(filler *Filler) {
		filler.renamer = r
	}
}

// WithFlagErrHandling returns a FillerOption thats specifies the flag.ErrorHandling
func WithFlagErrHandling(h flag.ErrorHandling) FillerOption {
	return func(filler *Filler) {
		filler.errHandle = h
	}
}

// NewFiller creates a new Filler,
// fsname is the name for flagset, usage is the overall usage introduction.
// optionally, a list of FillerOptions could be specified.
func NewFiller(fsname, usage string, options ...FillerOption) *Filler {
	r := &Filler{
		errHandle: DefaultErrHandle,
		renamer:   DefaultRenamer,
	}
	for _, o := range options {
		o(r)
	}
	r.fsMap = make(map[string]*Filler)
	r.translatedActNameMap = make(map[string]string)
	r.fs = flag.NewFlagSet(fsname, r.errHandle)
	r.fs.Usage = r.Usage
	r.orderList = []string{}
	r.usage = usage
	r.optList = options

	return r
}

func newInheritFiller(father *Filler, fsname, ousage string) *Filler {
	r := NewFiller(fsname, ousage, father.optList...)
	return r
}

var textEncodingInt = reflect.TypeOf((*encodingTextMarshaler)(nil)).Elem()

const (
	//SkipTag is the struct field tag used to skip flag generation
	SkipTag = "skipflag"
	//AliasTag is the struct field tag used to specify the parameter name iso field name
	AliasTag = "alias"
	//UsageTag is the struct field tag used to specify the usage of the field
	UsageTag = "usage"
	//ActTag is the struct field tag used to specify the field is an action
	ActTag = "action"
)

// Fill filler with struct in
func (filler *Filler) Fill(in any) error {
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return filler.walk(reflect.ValueOf(in), "", "", false)
	} else {
		return fmt.Errorf("only support a pointer to struct, but got %v", t)
	}

}

// GetFlagset returns the flagset used by the filler
func (filler *Filler) GetFlagset() *flag.FlagSet {
	return filler.fs
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
func (filler *Filler) walk(inV reflect.Value, nameprefix, usage string, isAct bool) error {
	fs := filler.fs
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
				//get tags
				if _, exists := fieldT.Tag.Lookup(SkipTag); exists {
					continue
				}
				usage, _ := fieldT.Tag.Lookup(UsageTag)
				fname := fieldT.Name
				if filler.renamer != nil {
					fname = filler.renamer(nameprefix, fname, isAct)
				}
				alias, _ := fieldT.Tag.Lookup(AliasTag)
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
						setTextEncodingType(fs, field, fname, usage)
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
				//check if the field is a struct
				if fieldT.Type.Kind() == reflect.Struct ||
					(fieldT.Type.Kind() == reflect.Pointer && fieldT.Type.Elem().Kind() == reflect.Struct) {
					if _, ok := fieldT.Tag.Lookup(ActTag); ok {
						if _, ok := filler.fsMap[fname]; ok {
							return fmt.Errorf("found struct type field with duplicate name %v", fname)
						}
						filler.fsMap[fname] = newInheritFiller(filler, fname, usage)
						filler.translatedActNameMap[fname] = fieldT.Name
						filler.orderList = append(filler.orderList, fname)

						// flag.NewFlagSet(fieldT.Name, filler.errHandle)

						err = filler.fsMap[fname].walk(field, fname, usage, true)
						if err != nil {
							return err
						}
						continue
					}
				}
				err = filler.walk(field, fname, usage, false)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var ErrInvalidAction = errors.New("unknown action")

// like, PrseArgs, use os.Args as input
func (filler *Filler) Parse() ([]string, error) {
	return filler.ParseArgs(os.Args[1:])
}

func (filler *Filler) getNextActPosState(args []string) (int, error) {
	const (
		stateArgDone = iota
		stateInArg
	)
	state := stateArgDone
	var hasdash bool
	for i, arg := range args {
		hasdash = strings.HasPrefix(arg, "-")
		switch state {
		case stateArgDone:
			if !hasdash {
				if _, ok := filler.fsMap[arg]; ok {
					return i, nil
				} else {
					return -1, fmt.Errorf(`found unrecognized action "%v"`, arg)
				}
			} else {
				//has -
				state = stateInArg
			}
		case stateInArg:
			if !hasdash {
				state = stateArgDone
			} else {
				// has -
				// this could be current arg is bool "like -arg1 -arg2"
			}

		}

	}
	return -1, nil
}

// ParseArgs parse the args, return parsed actions as a slice of string, each is a parsed action name
func (filler *Filler) ParseArgs(args []string) ([]string, error) {
	parsedActions := []string{}
	var nextActPos int = -1
	var nextAct string
	var err error
	errHanlder := func(inerr error) {
		if inerr != nil {
			switch filler.errHandle {
			case flag.ExitOnError:
				fmt.Println("-----?", err)
				os.Exit(2)
			case flag.PanicOnError:
				panic(err)
			}
		}

	}
	nextActPos, err = filler.getNextActPosState(args)
	if err != nil {
		errHanlder(err)
		return nil, err
	}
	if nextActPos >= 0 {
		nextAct = args[nextActPos]
	}
	endPos := len(args)
	if nextActPos >= 0 {
		endPos = nextActPos
	}
	err = filler.fs.Parse(args[:endPos])
	if err != nil {
		return nil, err
	}
	if nextActPos >= 0 {
		if nextFiller, ok := filler.fsMap[nextAct]; !ok {
			err = fmt.Errorf("%w: %v", ErrInvalidAction, nextAct)
			errHanlder(err)
			return nil, err
		} else {
			parsedActions = append(parsedActions, filler.translatedActNameMap[nextAct])
			acts, err := nextFiller.ParseArgs(args[endPos+1:])
			if err != nil {
				errHanlder(err)
				return nil, err
			}
			parsedActions = append(parsedActions, acts...)
		}
	}
	return parsedActions, nil
}

// GetActUsage returns filler's child action usage specified by actname,
// actname should be the field name before renaming;
// return "" if not found
func (filler *Filler) GetActUsage(actname string) string {
	for rn, n := range filler.translatedActNameMap {
		if n == actname {
			return filler.fsMap[rn].UsageStr("")
		}
	}
	return ""
}

// UsageStr return a usage string for the filler and its descendant fillers (a.k.a actions)
func (filler *Filler) UsageStr(prefix string) string {
	step := "  "
	indent := prefix + step
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, filler.usage)
	filler.fs.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(buf, "%v- %v: %v\n", indent, f.Name,
			// reflect.Indirect(reflect.ValueOf(f.Value)).Kind(),
			f.Usage)
		if f.DefValue != "" {
			fmt.Fprintf(buf, "%v\tdefault:%v\n", indent, f.DefValue)
		}
	})
	for _, childname := range filler.orderList {
		child := filler.fsMap[childname]
		fmt.Fprintf(buf, "%v= %v: ", indent, childname)
		fmt.Fprint(buf, child.UsageStr(indent))
	}
	return buf.String()
}

// Usage print the string returned by UsageStr
func (filler *Filler) Usage() {
	fmt.Println(filler.UsageStr(""))
}
