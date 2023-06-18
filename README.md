# myflags
myflags is a Golang module to make command line flag easier, it creates command line flags based on fields in a struct, so user doesn't need to manually create the flags one by one. it also supports more types than `flag` module, and easily extensible for new types, user could even add support for existing type without using alias type.

## Quick Start 
Using myflags is straight forward:

1. define all flags in a struct
2. create a `Filler` with the struct, call `Fill` method with the struct variable with default value. 
3. call `flag.Parse()`

Following is an example:
```
package main

import (
	"flag"
	"fmt"
	"log"
	"net/netip"
	"time"

	"github.com/hujun-open/myflags"
)

type Conf struct {
	Addr      *netip.Addr
	Name      string
	StartTime time.Time `layout:"2006 02 Jan 15:04"` //layout is the time format string
	Max       uint32    `base:"16"` //use base 16
}

func main() {
    //define the default value
	defaultConf := &Conf{
		Max: 15,
	}
	err := myflags.NewFiller().Fill(flag.CommandLine, defaultConf)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	fmt.Printf("%+v\n", *defaultConf)
}

```
the created flags:
```
 .\test.exe -?
flag provided but not defined: -?
Usage of C:\hujun\gomodules\src\test2\test.exe:
  -addr value

  -max value
         (default f)
  -name value
         (default )
  -starttime value
         (default 0001 01 Jan 00:00)

```
some parsing results:
```
 .\test.exe -name tom
{Addr:invalid IP Name:tom StartTime:0001-01-01 00:00:00 +0000 UTC Max:15}

 .\test.exe -name tom -addr 2001:dead::beef -starttime "2023 12 Dec 11:22"
{Addr:2001:dead::beef Name:tom StartTime:2023-12-12 11:22:00 +0000 UTC Max:15}

.\test.exe -name tom -addr 2001:dead::beef -starttime "2023 12 Dec 11:22" -max 2C
{Addr:2001:dead::beef Name:tom StartTime:2023-12-12 11:22:00 +0000 UTC Max:44}
```

## Supported Types
Base:
- all int/uint types: support `base` tag for the base
- float32/float64
- string
- bool
- time.Duration
- time.Time: supports `layout` tag for time layout string
- All types implement both of following interface:
    - `encoding.TextUnmarshaler`
    - `encoding.TextUnmarshaler`
- All type register via `myflags.Register` function

Note: flag is only created for exported struct field.


In addition to base types, following types are also supported:

- pointer to the base type 
- slice/array of base type
- slice/array of pointer to the base type

for slice/array, use "," as separator. 

myflags also supports following type of struct:

- nested struct, like:
```
type Outer struct {
    Nested struct {
        Name string
    }
}
```

- embeded struct, like:
```
type SubS struct {
    SubName string
}
type Outer struct {
    SubS
}
```

## Flag Naming
By default, the name of created flag is the lowercase of struct field name, if the the field is in a sub-struct, then it is parent_field_name+field_name.

Optionally a renaming function could be supplied when creating the `Filler`, myflags uses the renaming function returned string as the flag name.

A optional struct tag "alias" could be used to override above generated name.


## Extension
New type could be supported via `myflags.Register`, which takes a variable implements `myflags.RegisteredConverters` interface. the `myflags.Register` must be called before `myflags.Fill`, typically it should be called in `init()`.

Check [time.go](time.go), [inttype.go](inttype.go) for examples.

