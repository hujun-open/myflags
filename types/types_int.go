/*
# int/uint:

  - unmarshal: uses strconv.ParseInt(), which support different base like `0xaf`
  - marshal: base10 output, could be overridden by using using struct field tag `base`, with possible values 2, 8 or 16
*/
package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hujun-open/myflags/v2"
)

func init() {
	myflags.Register[int](&intType{len: 0, isUint: false})
	myflags.Register[int8](&intType{len: 8, isUint: false})
	myflags.Register[int16](&intType{len: 16, isUint: false})
	myflags.Register[int32](&intType{len: 32, isUint: false})
	myflags.Register[int64](&intType{len: 64, isUint: false})
	myflags.Register[uint](&intType{len: 0, isUint: true})
	myflags.Register[uint8](&intType{len: 8, isUint: true})
	myflags.Register[uint16](&intType{len: 16, isUint: true})
	myflags.Register[uint32](&intType{len: 32, isUint: true})
	myflags.Register[uint64](&intType{len: 64, isUint: true})
}

type intType struct {
	len    int
	isUint bool
}

func (i *intType) ToStr(in any, tag reflect.StructTag) string {
	//if in is a pointer, convert it to the value
	val := reflect.ValueOf(in)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	base, _ := tag.Lookup("base")
	fmtstr := "%d"
	switch strings.TrimSpace(base) {
	case "2":
		fmtstr = "0b%b"
	case "8":
		fmtstr = "0o%O"
	case "16":
		fmtstr = "0x%x"
	}
	return fmt.Sprintf(fmtstr, val.Interface())
}

func (i *intType) FromStr(s string, tag reflect.StructTag) (any, error) {
	if !i.isUint {
		n, err := strconv.ParseInt(strings.TrimSpace(s), 0, i.len)
		if err != nil {
			return nil, err
		}
		switch i.len {
		case 0:
			return int(n), nil
		case 8:
			return int8(n), nil
		case 16:
			return int16(n), nil
		case 32:
			return int32(n), nil
		case 64:
			return int64(n), nil
		}

	} else {
		n, err := strconv.ParseUint(strings.TrimSpace(s), 0, i.len)
		if err != nil {
			return nil, err
		}
		switch i.len {
		case 0:
			return uint(n), nil
		case 8:
			return uint8(n), nil
		case 16:
			return uint16(n), nil
		case 32:
			return uint32(n), nil
		case 64:
			return uint64(n), nil
		}

	}
	return nil, fmt.Errorf("not a supported type")
}
