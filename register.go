package myflags

import (
	"flag"
	"reflect"
)

type RegisteredConverters interface {
	ToStr(any, reflect.StructTag) string
	FromStr(string, reflect.StructTag) (any, error)
}

type registry struct {
	list map[string]RegisteredConverters
}

var globalRegistry = registry{list: make(map[string]RegisteredConverters)}

func (r *registry) GetViaInterface(in any) RegisteredConverters {
	t := reflect.TypeOf(in)

	if r, ok := r.list[getTypeName(t)]; ok {
		return r
	}
	return nil
}

func (r *registry) GetViaType(t reflect.Type) RegisteredConverters {
	if r, ok := r.list[getTypeName(t)]; ok {
		return r
	}
	return nil
}

func (r *registry) Register(t string, c RegisteredConverters) {
	r.list[t] = c
}

type factoryHandler func(fs *flag.FlagSet, ref reflect.Value, tag reflect.StructTag, name, usage string)

var factoryRegistry = make(map[string]factoryHandler)

func getFactory(in any) factoryHandler {
	t := reflect.TypeOf(in)
	if r, ok := factoryRegistry[getTypeName(t)]; ok {
		return r
	}
	return nil
}

func Register[T any](c RegisteredConverters) {
	f := factory[T]{}
	tname := getTypeName(reflect.TypeOf(*new(T)))
	factoryRegistry[tname] = f.process
	globalRegistry.Register(tname, c)
}
