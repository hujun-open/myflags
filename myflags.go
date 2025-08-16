package myflags

import (
	"bytes"
	"cmp"
	"encoding"
	"strconv"

	// "flag"
	"fmt"
	"reflect"
	"strings"

	flag "github.com/hujun-open/pflag"
	"golang.org/x/exp/slices"

	"github.com/hujun-open/cobra"
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
	*cobra.Command
	fsMap              map[string]*Filler //child fillers, key is the action name
	fs                 *flag.FlagSet
	errHandle          flag.ErrorHandling
	optList            []FillerOption
	usage              string //this is the usage string for for overall filler
	renamer            RenameFunc
	includeDocGenCMD   bool
	includeSummaryHelp bool
	nounVals           map[uint]reflect.Value //the noun field values
	nounFields         map[uint]reflect.StructField
	nounCompleters     map[uint]cobra.CompletionFunc
}

// FillerOption is an option when creating new Filler
type FillerOption func(filler *Filler)

// WithRenamer returns a FillerOption that specifies the rename function
func WithRenamer(r RenameFunc) FillerOption {
	return func(filler *Filler) {
		filler.renamer = r
	}
}

func WithShellCompletionCMD() FillerOption {
	return func(filler *Filler) {
		filler.CompletionOptions.DisableDefaultCmd = false
	}
}

// WithFlagErrHandling returns a FillerOption thats specifies the flag.ErrorHandling
func WithFlagErrHandling(h flag.ErrorHandling) FillerOption {
	return func(filler *Filler) {
		filler.errHandle = h
	}
}

// WithRootMethod set f as the Run method of the root command
func WithRootMethod(f RunMethod) FillerOption {
	return func(filler *Filler) {
		filler.Command.Root().Run = f
	}
}

// WithRootArgCompletionMethod set f as ValidArgsFunction of the root command
func WithRootArgCompletionMethod(f cobra.CompletionFunc) FillerOption {
	return func(filler *Filler) {
		filler.Command.Root().ValidArgsFunction = f
	}
}

// NewFiller creates a new Filler,
// name is the name for the command, usage is the overall usage introduction.
// optionally, a list of FillerOptions could be specified.
// Note: the name must be same as the name of application executable, otherwise shell completion won't work.
func NewFiller(name, usage string, options ...FillerOption) *Filler {
	r := &Filler{
		errHandle: DefaultErrHandle,
		renamer:   DefaultRenamer,
		Command: &cobra.Command{
			Use:   name,
			Short: usage,
		},
		nounVals:       make(map[uint]reflect.Value),
		nounFields:     make(map[uint]reflect.StructField),
		nounCompleters: make(map[uint]cobra.CompletionFunc),
	}
	r.Command.CompletionOptions.DisableDefaultCmd = true
	for _, o := range options {
		o(r)
	}
	r.fsMap = make(map[string]*Filler)
	r.fs = flag.NewFlagSet(name, r.errHandle)
	r.usage = usage
	r.optList = options

	return r
}

func newInheritFiller(father *Filler, fsname, ousage string, method RunMethod) *Filler {
	r := NewFiller(fsname, ousage, father.optList...)
	r.Run = method
	r.ValidArgsFunction = r.getActNounCompleter()
	father.AddCommand(r.Command)
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
	//ActTag is the struct field tag used to specify the field (of type struct) is an action, the value is the name of the method to run
	//method name is looked up in the field struct first, then in the root struct
	ActTag = "action"
	//RequiredTag indicate the flag is mandatory required
	RequiredTag = "required"
	//ValidValuesTag is a list of valid values for the field, separated by comma, used for completion
	ValidValuesTag = "choices"
	//NounTag marks a field as noun for a command, there is only one field will be treated as noun per (sub-)struct
	NounTag = "noun"
	//CompleteTag specify a method as cobra.CompleteFunc
	CompleteTag = "complete"
)

// RunMethod is the type of function could be used as cobra.Command.Run
type RunMethod func(cmd *cobra.Command, args []string)

// DefRunMethod is function that does nothing, it is assigned if value of ActTag is empty string
var DefRunMethod RunMethod = func(cmd *cobra.Command, args []string) {}

// if last is false, add the new act as first one in orderList
func (filler *Filler) addNewAct(act, usage string, method RunMethod) {
	filler.fsMap[act] = newInheritFiller(filler, act, usage, method)
}

const summaryStep = "  "

// SummaryFillerUsageStr returns a usage string that include usage of flags, child commands and their children
func SummaryFillerUsageStr(filler *Filler, prefix string) string {
	cmd := filler.Command
	indent := prefix + summaryStep
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v= %v\n", indent, cmd.UseLine())
	// nounBuf := new(bytes.Buffer)
	// w := tabwriter.NewWriter(nounBuf, 0, 0, 4, ' ', 0)
	// if filler.Runnable() {
	// 	fmt.Fprintf(buf, "  %s\n", filler.UseLine())
	// }

	// if filler.HasAvailableSubCommands() {
	// 	fmt.Printf("  %s [command]\n", filler.CommandPath())
	// }
	for _, id := range filler.getSortedNounIDs() {
		defStr := fmt.Sprintf("%s", filler.nounVals[id].Interface())
		conv := globalRegistry.GetViaInterface(filler.nounVals[id].Interface())
		if conv != nil {
			defStr = conv.ToStr(filler.nounVals[id].Interface(), filler.nounFields[id].Tag)
		}
		fmt.Fprintf(buf, "  %v<%v>: %v\n",
			indent,
			filler.nounFields[id].Name,
			filler.nounFields[id].Tag.Get(UsageTag),
		)
		fmt.Fprintf(buf, "    %vdefault:\"%v\"\n", indent, defStr)
	}
	// w.Flush()
	// fmt.Fprint(buf, nounBuf.String())

	cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
		if f.Shorthand == "" {
			fmt.Fprintf(buf, "  %v--%v: %v\n", indent, f.Name, f.Usage)
		} else {
			fmt.Fprintf(buf, "  %v-%v, --%v: %v\n", indent, f.Shorthand, f.Name, f.Usage)
		}

		if f.DefValue != "" {
			fmt.Fprintf(buf, "  %v\tdefault:%v\n", indent, f.DefValue)
		}
	})

	for _, childFiller := range getSortedMapVals(filler.fsMap) {

		// fmt.Fprintf(buf, "%v= %v: ", indent, childFiller.Use)
		fmt.Fprint(buf, SummaryFillerUsageStr(childFiller, indent))

	}
	return buf.String()

}

// SummaryHelpCMDName is the name of command print summary help
const SummaryHelpCMDName = "summaryhelp"

// SummaryHelpCMD returns a command `summaryhelp` that print summary help  that include usage of flags, child commands and their children
func (filler *Filler) SummaryHelpCMD() *cobra.Command {
	return &cobra.Command{
		Use:   SummaryHelpCMDName,
		Short: "help in summary",
		Run: func(c *cobra.Command, args []string) {
			fmt.Print(SummaryFillerUsageStr(filler, ""))
			for _, cmd := range filler.Commands() {
				if _, ok := filler.fsMap[cmd.Name()]; !ok {
					fmt.Print(SummaryUsageStr(cmd, summaryStep))
				}
			}

		},
	}

}

// WithSummaryHelp adds a command `summaryhelp` that print summary help  that include usage of flags, child commands and their children
func WithSummaryHelp() FillerOption {
	return func(filler *Filler) {
		filler.includeSummaryHelp = true
	}
}

// Fill filler with struct in
func (filler *Filler) Fill(in any) error {
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		err := filler.walk(reflect.ValueOf(in), reflect.ValueOf(in), "", true)
		if err != nil {

			return err
		}
		if filler.includeDocGenCMD {
			filler.AddCommand(docFiller.Command)
		}
		if filler.includeSummaryHelp {
			filler.AddCommand(filler.SummaryHelpCMD())
		}

		if len(filler.Commands()) == 0 {
			if !filler.Runnable() {
				filler.Run = DefRunMethod //this is needed otherwise, the "-h" or "--help" won't work
			}
		}
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

	case reflect.Int8:
		casted := ref.Interface().(*int8)
		if strings.TrimSpace(short) != "" {
			fs.Int8VarP(casted, name, short, *casted, usage)
		} else {
			fs.Int8Var(casted, name, *casted, usage)
		}
	case reflect.Int16:
		casted := ref.Interface().(*int16)
		if strings.TrimSpace(short) != "" {
			fs.Int16VarP(casted, name, short, *casted, usage)
		} else {
			fs.Int16Var(casted, name, *casted, usage)
		}
	case reflect.Int32:
		casted := ref.Interface().(*int32)
		if strings.TrimSpace(short) != "" {
			fs.Int32VarP(casted, name, short, *casted, usage)
		} else {
			fs.Int32Var(casted, name, *casted, usage)
		}
	case reflect.Uint64:
		casted := ref.Interface().(*uint64)
		if strings.TrimSpace(short) != "" {
			fs.Uint64VarP(casted, name, short, *casted, usage)
		} else {
			fs.Uint64Var(casted, name, *casted, usage)
		}
	case reflect.Uint8:
		casted := ref.Interface().(*uint8)
		if strings.TrimSpace(short) != "" {
			fs.Uint8VarP(casted, name, short, *casted, usage)
		} else {
			fs.Uint8Var(casted, name, *casted, usage)
		}
	case reflect.Uint16:
		casted := ref.Interface().(*uint16)
		if strings.TrimSpace(short) != "" {
			fs.Uint16VarP(casted, name, short, *casted, usage)
		} else {
			fs.Uint16Var(casted, name, *casted, usage)
		}
	case reflect.Uint32:
		casted := ref.Interface().(*uint32)
		if strings.TrimSpace(short) != "" {
			fs.Uint32VarP(casted, name, short, *casted, usage)
		} else {
			fs.Uint32Var(casted, name, *casted, usage)
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
	case reflect.Float64, reflect.Float32,
		reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.String,
		reflect.Bool,
		reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
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

func isSupportedKind(inK reflect.Kind) bool {
	switch inK {
	case reflect.Invalid, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
		return false
	}
	return true
}

// in must be a pointer to struct
// NOTE: there are following methods to register a flag
// - setStandardFlagType
// - setTextEncodingType
// - simpleType.process
func (filler *Filler) walk(root, inV reflect.Value, nameprefix string, isAct bool) (ferr error) {
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
	validFlagValsMap := make(map[string]string)
	flagCompleteMethodMap := make(map[string]cobra.CompletionFunc)
	// flagCompleteFuncMap:=make(map[string]cobra.com)
	defer func() {
		if isAct {
			//updateing usage
			if len(filler.nounFields) > 0 {
				filler.Use = filler.getUse()
			}
			filler.SetUsageFunc(filler.usageFunc)
			//check if noun id is squential
			idList := filler.getSortedNounIDs()
			for i, id := range idList {
				if i != int(id-1) {
					ferr = fmt.Errorf("unsequential noun id, expect %d, but get %d", i+1, id)
				}
			}
			if len(idList) > 0 {
				filler.Command.Args = cobra.MaximumNArgs(len(idList))
			}

			filler.PersistentFlags().AddFlagSet(fs)
			for _, f := range requiredFlags {
				derr := cobra.MarkFlagRequired(fs, f)
				if derr != nil {
					ferr = fmt.Errorf("failed to mark flag %v required, %w", f, derr)
					return
				}
			}
			for flagname, valstr := range validFlagValsMap {
				helper := newValidFlagValues(valstr)
				derr := filler.Command.RegisterFlagCompletionFunc(flagname, helper.complete)
				if derr != nil {
					ferr = fmt.Errorf("failed to register flag completion function for %v, %w", flagname, derr)
					return
				}

			}
			for flagname, m := range flagCompleteMethodMap {
				derr := filler.Command.RegisterFlagCompletionFunc(flagname, m)
				if derr != nil {
					ferr = fmt.Errorf("failed to register complete function for flag %v, %w", flagname, err)
					return
				}
			}
		}
	}()

	// //check if it implements EncodingTextMarshaler inteface
	// if inT.Implements(textEncodingInt) {
	// 	setTextEncodingType(fs, inV, nameprefix, short, usage)
	// 	return nil
	// }
	// //these are kinds directly supported by flag module
	// if isFlagSupportedKind(ElemK) {
	// 	setStandardFlagType(fs, inV, nameprefix, short, usage)
	// 	return nil
	// }

	switch ElemK {
	case reflect.Struct:
		//a struct
		for i := 0; i < inV.Elem().NumField(); i++ {
			// fmt.Println("walk into ", inT.Elem().Field(i).Name, inT.Elem().Field(i).Type)
			field := inV.Elem().Field(i)
			fieldT := inT.Elem().Field(i)
			isNoun := false
			if fieldT.IsExported() {
				//only handle exported field

				//skip unsupported kind
				if !isSupportedKind(fieldT.Type.Kind()) {
					continue
				}
				//get tags
				if _, exists := fieldT.Tag.Lookup(SkipTag); exists {
					continue
				}
				//check complete method
				var completeMethod cobra.CompletionFunc
				if flagCompleteName, exists := fieldT.Tag.Lookup(CompleteTag); exists {
					if _, exists := fieldT.Tag.Lookup(ValidValuesTag); exists {
						return fmt.Errorf("%v tag and %v tag can't be set at the same time", CompleteTag, ValidValuesTag)
					}
					flagCompleteName = strings.TrimSpace(flagCompleteName)
					if flagCompleteName != "" {
						methodVal := getMethod(root, field, flagCompleteName)
						if !methodVal.IsValid() {
							return fmt.Errorf("action %v's method %v not found", fieldT.Name, flagCompleteName)
						}
						completeMethod = methodVal.Interface().(cobra.CompletionFunc)

					}
				}
				validFlagVals, _ := fieldT.Tag.Lookup(ValidValuesTag)
				//check if it is noun field
				var nounID uint64 = 1
				if ids, exists := fieldT.Tag.Lookup(NounTag); exists {
					isNoun = true

					if strings.TrimSpace(ids) != "" {
						nounID, ferr = strconv.ParseUint(ids, 10, 64)

						if ferr != nil {
							ferr = fmt.Errorf("%v's noun id is not a uint, %w", fieldT.Name, ferr)
							return ferr
						}
						if nounID == 0 {
							ferr = fmt.Errorf("%v's noun id is 0, but the noun id starts from 1", fieldT.Name)
							return
						}
					}
					if _, ok := filler.nounVals[uint(nounID)]; ok {
						ferr = fmt.Errorf("%v'has duplicate noun id", fieldT.Name)
						return
					}
					filler.nounVals[uint(nounID)] = field
					filler.nounFields[uint(nounID)] = fieldT
					filler.Command.PreRunE = func(cmd *cobra.Command, args []string) error {
						return filler.parseNoun(args)
					}
					if completeMethod == nil {
						//if there is no CompleteTag, check for ValidValuesTag value
						if strings.TrimSpace(validFlagVals) != "" {
							helper := newValidFlagValues(validFlagVals)
							completeMethod = helper.complete
						}
					}
					filler.nounCompleters[uint(nounID)] = completeMethod
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

				//check if the field is an action struct
				if fieldT.Type.Kind() == reflect.Struct ||
					(fieldT.Type.Kind() == reflect.Pointer && fieldT.Type.Elem().Kind() == reflect.Struct) {
					if methodName, ok := fieldT.Tag.Lookup(ActTag); ok { //if this is an act struct
						if _, ok := filler.fsMap[fname]; ok {
							return fmt.Errorf("found struct type field with duplicate name %v", fname)
						}
						var m RunMethod = DefRunMethod
						methodName = strings.TrimSpace(methodName)
						if methodName != "" {
							methodVal := getMethod(root, field, methodName)
							if !methodVal.IsValid() {
								return fmt.Errorf("action %v's method %v not found", fieldT.Name, methodName)
							}
							m = methodVal.Interface().(func(*cobra.Command, []string))
						}

						filler.addNewAct(fname, usage, m)
						err = filler.fsMap[fname].walk(root, field, fname, true)

						if err != nil {
							return err
						}
						continue
					}
				}

				//from now on, the field is NOT action struct

				//flagFunc gets called for each valid flag
				flagFunc := func(islist bool) error {
					if islist {
						return nil
					}
					if validFlagVals != "" {
						validFlagValsMap[fname] = validFlagVals
					}
					if completeMethod != nil {
						if !isNoun {
							flagCompleteMethodMap[fname] = completeMethod
						}
					}
					return nil

				}

				//check if it is a registered type, a.k.a simpleType
				f := getFactory(field.Interface())
				if f != nil {
					if field.Kind() != reflect.Pointer {
						field = field.Addr()
					}
					f(fs, field, fieldT.Tag, fname, fshort, usage)
					err = flagFunc(false)
					if err != nil {
						return fmt.Errorf("there is error processing flag %v, %w", fname, err)
					}
					continue
				}
				//check if it implements textMarshal
				if fieldT.Type.Kind() == reflect.Pointer {
					if fieldT.Type.Implements(textEncodingInt) {
						//pointer to textmarshale
						setTextEncodingType(fs, field, fname, fshort, usage)
						err = flagFunc(false)
						if err != nil {
							return fmt.Errorf("there is error processing flag %v, %w", fname, err)
						}
						continue
					}
					//these are kinds directly supported by flag module
					if isFlagSupportedKind(field.Elem().Kind()) {
						setStandardFlagType(fs, field, fname, fshort, usage)
						err = flagFunc(false)
						if err != nil {
							return fmt.Errorf("there is error processing flag %v, %w", fname, err)
						}
						continue
					}
				} else {
					if reflect.PointerTo(fieldT.Type).Implements(textEncodingInt) {
						//textmarshale
						setTextEncodingType(fs, field.Addr(), fname, fshort, usage)
						err = flagFunc(false)
						if err != nil {
							return fmt.Errorf("there is error processing flag %v, %w", fname, err)
						}
						continue
					}
					//these are kinds directly supported by flag module
					if isFlagSupportedKind(fieldT.Type.Kind()) {
						setStandardFlagType(fs, field.Addr(), fname, fshort, usage)
						err = flagFunc(false)
						if err != nil {
							return fmt.Errorf("there is error processing flag %v, %w", fname, err)
						}
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
						err = processList(fs, field.Addr(), fieldT.Tag, fname, fshort, usage)
						if err != nil {
							return err
						}
						err = flagFunc(true)
						if err != nil {
							return fmt.Errorf("there is error processing flag %v, %w", fname, err)
						}
						continue
					} else {
						return fmt.Errorf("%v is a slice/array of unsupported type %v", fieldT.Name, fieldT)
					}

				}

				//non-act struct or
				err = filler.walk(root, field, fname, false)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// IsOwnAction check if the cobra.Command.ExecuteC() returned command cmd is intended for cobra's own action, like help, completion command or --version flag with root command
// completionCMDName and helpCMDName specifies corresponding completion and help command name,
// "completion" and "help" are used if they are empty string.
// it also return true if cmd is the root command and skipRootCMD is true,
func IsOwnAction(cmd *cobra.Command, completionCMDName, helpCMDName string, skipRootCMD bool) bool {
	switch cmd.Name() {
	case cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd, DocgenCMDName, SummaryHelpCMDName:
		return true
	}

	if cmd.Root() == cmd {
		if skipRootCMD {
			return true
		} else {
			if f := cmd.Flags().Lookup("version"); f != nil {
				if f.Value.String() == "true" {
					return true
				}
			}
		}
	}
	if completionCMDName == "" {
		completionCMDName = "completion"
	}
	if helpCMDName == "" {
		helpCMDName = "help"
	}
	switch cmd.Name() {
	case completionCMDName, helpCMDName:
		return true
	}
	if cmd.Parent() != nil {
		if cmd.Parent().Name() == completionCMDName {
			return true
		}
	}
	if f := cmd.Flags().Lookup("help"); f != nil {
		if f.Value.String() == "true" {
			return true
		}
	}

	return false

}

// FindFiller return corresponding filler that handles given args
func (filler *Filler) FindFiller(args []string) *Filler {
	cmd, _, _ := filler.Find(args)

	path := "/" + strings.Join(strings.Fields(cmd.CommandPath())[1:], "/")

	return filler.GetChildFiller(path)
}

// GetChildFiller return a child Filler specified by child path,
// which is a string of list of command names separated by "/", start with "/" which represents calling filler's command
// e.g. /act1/act12/act121; return nil if not found or childpath is not valid
func (filler *Filler) GetChildFiller(childpath string) *Filler {
	if childpath[0] != '/' {
		return nil
	}
	if childpath == "/" {
		return filler
	}
	curCMD := filler
	for _, p := range strings.FieldsFunc(childpath, func(c rune) bool { return c == '/' }) {
		found := false
		for _, child := range curCMD.fsMap {
			if child.Command.Name() == p {
				found = true
				curCMD = child
				break
			}
		}
		if !found {
			return nil
		}
	}
	return curCMD
}

// GetChildCommand return a child command specified by child path,
// which is a string of list of command names separated by "/", start with "/" which represents calling filler's command
// e.g. /act1/act12/act121; return nil if not found or childpath is not valid
func (filler *Filler) GetChildCommand(childpath string) *cobra.Command {
	if childpath[0] != '/' {
		return nil
	}
	curCMD := filler.Command
	for _, p := range strings.FieldsFunc(childpath, func(c rune) bool { return c == '/' }) {
		found := false
		for _, child := range curCMD.Commands() {
			if child.Name() == p {
				found = true
				curCMD = child
				break
			}
		}
		if !found {
			return nil
		}
	}
	return curCMD
}

// ExecutePath run filler.ExecuteC() and return the executed command and full command path as slice of string,
// first level sub-command name invoked is the first item in the slice.
func (filler *Filler) ExecuteCMDPath() (*cobra.Command, []string, error) {
	cmd, err := filler.ExecuteC()
	if err != nil {
		return nil, nil, err
	}
	return cmd, strings.Fields(cmd.CommandPath())[1:], nil
}

func getSortedMapVals[K cmp.Ordered, V any](m map[K]V) []V {
	klist := []K{}
	for k := range m {
		klist = append(klist, k)
	}
	slices.Sort(klist)
	r := []V{}
	for _, k := range klist {
		r = append(r, m[k])
	}
	return r
}

// SummaryUsageStr returns a usage string that include usage of flags, child commands and their children
func SummaryUsageStr(cmd *cobra.Command, prefix string) string {

	indent := prefix + summaryStep
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v= %v\n", indent, cmd.UseLine())
	cmd.LocalFlags().VisitAll(func(f *flag.Flag) {
		if f.Shorthand == "" {
			fmt.Fprintf(buf, "  %v--%v: %v\n", indent, f.Name, f.Usage)
		} else {
			fmt.Fprintf(buf, "  %v-%v, --%v: %v\n", indent, f.Shorthand, f.Name, f.Usage)
		}

		if f.DefValue != "" {
			fmt.Fprintf(buf, "  %v\tdefault:%v\n", indent, f.DefValue)
		}
	})

	for _, childCMD := range cmd.Commands() {

		// fmt.Fprintf(buf, "%v= %v: ", indent, childCMD.Name())
		fmt.Fprint(buf, SummaryUsageStr(childCMD, indent))

	}
	return buf.String()

}
