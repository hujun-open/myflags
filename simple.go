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
// see time.go and net.go for implementation examples
type FromStrFunc func(s string, tag reflect.StructTag) (any, error)

type ToStrFunc func(any, reflect.StructTag) string

type simpleType[T any] struct {
	val     *T
	tags    reflect.StructTag
	fromStr FromStrFunc
	toStr   ToStrFunc
}

func newSimpleType[T any](from FromStrFunc, to ToStrFunc, tag reflect.StructTag) simpleType[T] {
	return simpleType[T]{val: new(T), toStr: to, fromStr: from, tags: tag}
}

func (v *simpleType[T]) String() string {
	if v.val == nil {
		return fmt.Sprint(nil)
	}
	return v.toStr(*v.val, v.tags)
}

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

type factory[T any] struct{}

func (f *factory[T]) process(fs *flag.FlagSet, ref reflect.Value, tag reflect.StructTag, name, usage string) {
	// if ref.Type().Elem().Kind()

	casted := ref.Interface().(*T)
	conv := globalRegistry.GetViaInterface(ref.Interface())
	newval := newSimpleType[T](conv.FromStr, conv.ToStr, tag)
	newval.SetRef(casted)
	fs.Var(&newval, name, usage)
}

func (v *simpleType[T]) SetRef(t *T) {
	v.val = t
}