package myflags

import (
	"encoding"
	"errors"

	// "flag"
	"fmt"
	"reflect"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"
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
	fs                   *flag.FlagSet
	errHandle            flag.ErrorHandling
	optList              []FillerOption
	usage                string //this is the usage string for for overall filler
	renamer              RenameFunc
	translatedActNameMap map[string]string //key is the transalted action name, val is the original field name
	cobraCMD             *cobra.Command
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
	r.usage = usage
	r.optList = options
	r.cobraCMD = &cobra.Command{
		Use:   fsname,
		Short: usage,
	}

	return r
}

func newInheritFiller(father *Filler, fsname, ousage string, method runMethod) *Filler {
	r := NewFiller(fsname, ousage, father.optList...)
	r.cobraCMD.Run = method
	father.cobraCMD.AddCommand(r.cobraCMD)
	return r
}

var textEncodingInt = reflect.TypeOf((*encodingTextMarshaler)(nil)).Elem()

const (
	//SkipTag is the struct field tag used to skip flag generation
	SkipTag = "skipflag"
	//AliasTag is the struct field tag used to specify the flag name iso field name
	AliasTag = "alias"
	//ShorthandTag is the struct field tag used to specify shorthand name for the flag
	ShorthandTag = "short"
	//UsageTag is the struct field tag used to specify the usage of the field
	UsageTag = "usage"
	//ActTag is the struct field tag used to specify the field is an action
	ActTag = "action"
	//ActMethodTag is the method name for the action
	ActMethodTag = "method"
	//RequiredTag indicate the flag is mandatory required
	RequiredTag = "required"
)

type runMethod func(cmd *cobra.Command, args []string)

// if last is false, add the new act as first one in orderList
func (filler *Filler) addNewAct(smallAct, bigAct, usage string, method runMethod, last bool) {

	filler.fsMap[smallAct] = newInheritFiller(filler, smallAct, usage, method)
	filler.translatedActNameMap[smallAct] = bigAct

}

func (filler *Filler) PrintDebug() {
	fmt.Println("actions:")
	if len(filler.fsMap) > 0 {
		for action, child := range filler.fsMap {
			fmt.Printf("action:%v\n", action)
			child.PrintDebug()
		}
	}
}

// Fill filler with struct in
func (filler *Filler) Fill(in any) error {
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		err := filler.walk(reflect.ValueOf(in), reflect.ValueOf(in), "", "", "", true)
		if err != nil {
			return err
		}
		// if filler.withCompletionCMD {
		// 	filler.addNewAct("complete", CompleteCMDName, "generate completion script", filler.GenCompletionScript, false)
		// }
		return nil
	} else {
		return fmt.Errorf("only support a pointer to struct, but got %v", t)
	}

}

// GetFlagset returns the flagset used by the filler
func (filler *Filler) GetFlagset() *flag.FlagSet {
	return filler.fs
}

func setStandardFlagType(fs *flag.FlagSet, ref reflect.Value, name, short, usage string) {
	switch ref.Elem().Kind() {
	case reflect.String:
		casted := ref.Interface().(*string)
		if strings.TrimSpace(short) != "" {
			fs.StringVarP(casted, name, short, *casted, usage)
		} else {
			fs.StringVar(casted, name, *casted, usage)
		}
	case reflect.Int:
		casted := ref.Interface().(*int)
		if strings.TrimSpace(short) != "" {
			fs.IntVarP(casted, name, short, *casted, usage)
		} else {
			fs.IntVar(casted, name, *casted, usage)
		}

	case reflect.Uint:
		casted := ref.Interface().(*uint)
		if strings.TrimSpace(short) != "" {
			fs.UintVarP(casted, name, short, *casted, usage)
		} else {
			fs.UintVar(casted, name, *casted, usage)
		}
	case reflect.Bool:
		casted := ref.Interface().(*bool)
		if strings.TrimSpace(short) != "" {
			fs.BoolVarP(casted, name, short, *casted, usage)
		} else {
			fs.BoolVar(casted, name, *casted, usage)
		}

	case reflect.Int64:
		casted := ref.Interface().(*int64)
		if strings.TrimSpace(short) != "" {
			fs.Int64VarP(casted, name, short, *casted, usage)
		} else {
			fs.Int64Var(casted, name, *casted, usage)
		}
	case reflect.Uint64:
		casted := ref.Interface().(*uint64)
		if strings.TrimSpace(short) != "" {
			fs.Uint64VarP(casted, name, short, *casted, usage)
		} else {
			fs.Uint64Var(casted, name, *casted, usage)
		}
	case reflect.Float32:
		casted := ref.Interface().(*float32)
		if strings.TrimSpace(short) != "" {
			fs.Float32VarP(casted, name, short, *casted, usage)
		} else {
			fs.Float32Var(casted, name, *casted, usage)
		}

	case reflect.Float64:
		casted := ref.Interface().(*float64)
		if strings.TrimSpace(short) != "" {
			fs.Float64VarP(casted, name, short, *casted, usage)
		} else {
			fs.Float64Var(casted, name, *casted, usage)
		}
	}
}

func setTextEncodingType(fs *flag.FlagSet, ref reflect.Value, name, short, usage string) {
	casted := ref.Interface().(encodingTextMarshaler)
	if strings.TrimSpace(short) != "" {
		fs.TextVarP(casted, name, short, casted, usage)
		return
	}
	fs.TextVar(casted, name, casted, usage)
}

func isFlagSupportedKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float64, reflect.Float32, reflect.Int,
		reflect.Int64,
		reflect.String,
		reflect.Bool, reflect.Uint,
		reflect.Uint64:
		return true
	}
	return false
}

// getMethod return method value specified by the name if current has it, if not, then return root's method with the same name
func getMethod(root, current reflect.Value, name string) reflect.Value {
	cur := current
	if cur.Kind() != reflect.Pointer {
		cur = current.Addr()
	}
	methodVal := cur.MethodByName(name)
	if methodVal.IsValid() {
		return methodVal
	}
	r := root
	if r.Kind() != reflect.Pointer {
		r = root.Addr()
	}
	return r.MethodByName(name)

}

// in must be a pointer to struct
// NOTE: there are following methods to register a flag
// - setStandardFlagType
// - setTextEncodingType
// - simpleType.process
func (filler *Filler) walk(root, inV reflect.Value, nameprefix, short, usage string, isAct bool) error {
	fs := filler.fs
	requiredFlags := []string{}
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
	defer func() {
		if isAct {
			filler.cobraCMD.PersistentFlags().AddFlagSet(fs)
			for _, f := range requiredFlags {
				cobra.MarkFlagRequired(fs, f)
			}
		}
	}()

	//check if it implements EncodingTextMarshaler inteface
	if inT.Implements(textEncodingInt) {

		setTextEncodingType(fs, inV, nameprefix, short, usage)
		return nil
	}
	//these are kinds directly supported by flag module
	if isFlagSupportedKind(ElemK) {
		setStandardFlagType(fs, inV, nameprefix, short, usage)
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
				fshort, ok := fieldT.Tag.Lookup(ShorthandTag)
				if ok {
					if len(fshort) > 1 {
						return fmt.Errorf("%v's %v tag can't be more than one letter long", fieldT.Name, ShorthandTag)
					}
				}
				if _, ok := fieldT.Tag.Lookup(RequiredTag); ok {
					requiredFlags = append(requiredFlags, fname)
				}

				if field.Kind() == reflect.Pointer {
					if field.IsNil() {
						//initilize the nil pointer
						field.Set(reflect.New(fieldT.Type.Elem()))
					}
				}
				//check if it is a registered type, a.k.a simpleType
				f := getFactory(field.Interface())
				if f != nil {
					if field.Kind() != reflect.Pointer {
						field = field.Addr()
					}
					f(fs, field, fieldT.Tag, fname, fshort, usage)
					continue
				}
				//check if it implements textMarshal
				if fieldT.Type.Kind() == reflect.Pointer {
					if fieldT.Type.Implements(textEncodingInt) {
						//pointer to textmarshale
						setTextEncodingType(fs, field, fname, fshort, usage)
						continue
					}
				} else {
					if reflect.PointerTo(fieldT.Type).Implements(textEncodingInt) {
						//textmarshale
						setTextEncodingType(fs, field.Addr(), fname, fshort, usage)
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
						if methodName, ok := fieldT.Tag.Lookup(ActMethodTag); !ok {
							return fmt.Errorf("action %v doesn't have %v tag", fieldT.Name, ActMethodTag)
						} else {
							methodVal := getMethod(root, field, methodName)
							if !methodVal.IsValid() {
								return fmt.Errorf("action %v's method %v not found", fieldT.Name, methodName)
							}

							// filler.fsMap[fname] = newInheritFiller(filler, fname, usage)
							// filler.translatedActNameMap[fname] = fieldT.Name
							// filler.orderList = append(filler.orderList, fname)

							filler.addNewAct(fname, fieldT.Name, usage, methodVal.Interface().(func(*cobra.Command, []string)), true)

							// flag.NewFlagSet(fieldT.Name, filler.errHandle)

							err = filler.fsMap[fname].walk(root, field, fname, fshort, usage, true)
							if err != nil {
								return err
							}
							continue

						}
					}
				}
				err = filler.walk(root, field, fname, fshort, usage, false)
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
// func (filler *Filler) Parse() ([]string, error) {
// 	cmd, args, err := filler.cobraCMD.Traverse(os.Args[1:])
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("cobra", cmd.Use, args)

// 	return filler.ParseArgs(os.Args[1:])
// }

type isBoolInt interface {
	IsBoolFlag() bool
}

func (filler *Filler) getNextActPosState(args []string) (int, error) {
	const (
		stateArgDone = iota
		stateInArg
	)
	state := stateArgDone
	var hasdash bool
L1:
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
				//check if it is argname=xxx format
				_, _, found := strings.Cut(arg[1:], "=")
				if found {
					//yes
					state = stateArgDone
					continue L1
				} else {
					//no,meaing it is just "-argname" check if this is boolvar
					isBool := false
					filler.fs.VisitAll(func(f *flag.Flag) {
						if f.Name == arg[1:] {
							if s, ok := f.Value.(isBoolInt); ok {
								if s.IsBoolFlag() {
									isBool = true
								}
							}
						}
					})
					if isBool {
						state = stateArgDone
						continue L1
					}
				}
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
// func (filler *Filler) ParseArgs(args []string) ([]string, error) {
// 	parsedActions := []string{}
// 	var nextActPos int = -1
// 	var nextAct string
// 	var err error
// 	errHanlder := func(inerr error) {
// 		if inerr != nil {
// 			switch filler.errHandle {
// 			case flag.ExitOnError:
// 				fmt.Println("-----?", err)
// 				os.Exit(2)
// 			case flag.PanicOnError:
// 				panic(err)
// 			}
// 		}

// 	}
// 	nextActPos, err = filler.getNextActPosState(args)
// 	if err != nil {
// 		errHanlder(err)
// 		return nil, err
// 	}
// 	if nextActPos >= 0 {
// 		nextAct = args[nextActPos]
// 	}
// 	endPos := len(args)
// 	if nextActPos >= 0 {
// 		endPos = nextActPos
// 	}
// 	err = filler.fs.Parse(args[:endPos])
// 	if err != nil {
// 		return nil, err
// 	}
// 	if nextActPos >= 0 {
// 		if nextFiller, ok := filler.fsMap[nextAct]; !ok {
// 			err = fmt.Errorf("%w: %v", ErrInvalidAction, nextAct)
// 			errHanlder(err)
// 			return nil, err
// 		} else {
// 			parsedActions = append(parsedActions, filler.translatedActNameMap[nextAct])
// 			acts, err := nextFiller.ParseArgs(args[endPos+1:])
// 			if err != nil {
// 				errHanlder(err)
// 				return nil, err
// 			}
// 			parsedActions = append(parsedActions, acts...)
// 		}
// 	}
// 	return parsedActions, nil
// }

// // GetActUsage returns filler's child action usage specified by actname,
// // actname should be the field name before renaming;
// // return "" if not found
// func (filler *Filler) GetActUsage(actname string) string {
// 	for rn, n := range filler.translatedActNameMap {
// 		if n == actname {
// 			return filler.fsMap[rn].UsageStr("")
// 		}
// 	}
// 	return ""
// }

// // UsageStr return a usage string for the filler and its descendant fillers (a.k.a actions)
// func (filler *Filler) UsageStr(prefix string) string {
// 	step := "  "
// 	indent := prefix + step
// 	buf := new(bytes.Buffer)
// 	fmt.Fprintln(buf, filler.usage)
// 	filler.fs.VisitAll(func(f *flag.Flag) {
// 		fmt.Fprintf(buf, "%v- %v: %v\n", indent, f.Name,
// 			// reflect.Indirect(reflect.ValueOf(f.Value)).Kind(),
// 			f.Usage)
// 		if f.DefValue != "" {
// 			fmt.Fprintf(buf, "%v\tdefault:%v\n", indent, f.DefValue)
// 		}
// 	})
// 	for _, childname := range filler.orderList {
// 		child := filler.fsMap[childname]
// 		fmt.Fprintf(buf, "%v= %v: ", indent, childname)
// 		fmt.Fprint(buf, child.UsageStr(indent))
// 	}
// 	return buf.String()
// }

// // Usage print the string returned by UsageStr
// func (filler *Filler) Usage() {
// 	fmt.Println(filler.UsageStr(""))
// }

func (filler *Filler) Exec() error {
	return filler.cobraCMD.Execute()
}

func (filler *Filler) GetCobraCMD() *cobra.Command {
	return filler.cobraCMD
}
