package myflags

import (
	"flag"
	"fmt"
	"reflect"
)

func getTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return fmt.Sprint(t.PkgPath() + "=>" + t.String())
}

// FromStrFunc is a function convert string s into a specific type T, the tag is the struct field tag, as addtional input.
// see time.go for implementation examples
type FromStrFunc func(s string, tags reflect.StructTag) (any, error)

// ToStrFunc is a function convert in to string, the tag is the struct field tag, as addtional input.
// see time.go for implementation examples
type ToStrFunc func(in any, tags reflect.StructTag) string

type simpleType[T any] struct {
	val     *T
	tags    reflect.StructTag
	fromStr FromStrFunc
	toStr   ToStrFunc
	isBool  bool
}

func newSimpleType[T any](from FromStrFunc, to ToStrFunc, tag reflect.StructTag, isbool bool) simpleType[T] {
	return simpleType[T]{val: new(T), toStr: to, fromStr: from, tags: tag, isBool: isbool}
}

// implment flag.Value interface
func (v *simpleType[T]) String() string {
	if v.val == nil {
		return fmt.Sprint(nil)
	}
	return v.toStr(*v.val, v.tags)
}

// implment flag.Value interface
func (v *simpleType[T]) Set(s string) error {
	r, err := v.fromStr(s, v.tags)
	if err != nil {
		return fmt.Errorf("failed to parse %s into %T, %w", s, *(new(T)), err)
	}
	// if v.val == nil {
	// 	v.val = new(T)
	// }
	*v.val = r.(T)
	return nil
}

// implment flag.Value interface
func (v *simpleType[T]) IsBoolFlag() bool { return v.isBool }

type factory[T any] struct{}

func (f *factory[T]) process(fs *flag.FlagSet, ref reflect.Value, tag reflect.StructTag, name, usage string) {
	// if ref.Type().Elem().Kind()
	isbool := false
	if reflect.TypeOf(*new(T)).Kind() == reflect.Bool {
		isbool = true
	}
	casted := ref.Interface().(*T)
	conv := globalRegistry.GetViaInterface(ref.Interface())
	newval := newSimpleType[T](conv.FromStr, conv.ToStr, tag, isbool)
	newval.SetRef(casted)
	fs.Var(&newval, name, usage)
}

func (v *simpleType[T]) SetRef(t *T) {
	v.val = t
}
