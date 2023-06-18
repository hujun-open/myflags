package myflags

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func init() {
	Register[int](&intType{len: 0, isUint: false})
	Register[int8](&intType{len: 8, isUint: false})
	Register[int16](&intType{len: 16, isUint: false})
	Register[int32](&intType{len: 32, isUint: false})
	Register[int64](&intType{len: 64, isUint: false})
	Register[uint](&intType{len: 0, isUint: true})
	Register[uint8](&intType{len: 8, isUint: true})
	Register[uint16](&intType{len: 16, isUint: true})
	Register[uint32](&intType{len: 32, isUint: true})
	Register[uint64](&intType{len: 64, isUint: true})
}

type intType struct {
	len    int
	isUint bool
}

func (i *intType) ToStr(in any, tag reflect.StructTag) string {
	base, _ := tag.Lookup("base")
	fmtstr := "%d"
	switch strings.TrimSpace(base) {
	case "2":
		fmtstr = "%b"
	case "8":
		fmtstr = "%O"
	case "16":
		fmtstr = "%x"
	}
	return fmt.Sprintf(fmtstr, in)
}

func (i *intType) FromStr(s string, tag reflect.StructTag) (any, error) {
	base, _ := tag.Lookup("base")
	baseN := 10
	switch strings.TrimSpace(base) {
	case "2":
		baseN = 2
	case "8":
		baseN = 8
	case "16":
		baseN = 16
	}
	if !i.isUint {
		n, err := strconv.ParseInt(strings.TrimSpace(s), baseN, i.len)
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
		n, err := strconv.ParseUint(strings.TrimSpace(s), baseN, i.len)
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
