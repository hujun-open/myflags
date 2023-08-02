package myflags_test

import (
	"flag"
	"fmt"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hujun-open/myflags"
)

type intTyps interface {
	int | uint | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

func createInt[T intTyps](value int64) *T {
	r := new(T)
	*r = T(value)
	return r
}

type Subno struct {
	Counter uint16
}

type SubnoList []Subno

func (sl SubnoList) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprint(sl)), nil
}
func (sl *SubnoList) UnmarshalText(text []byte) error {
	r := SubnoList{}
	flist := strings.Split(string(text), ",")
	for _, ns := range flist {
		n, err := strconv.ParseUint(ns, 10, 16)
		if err != nil {
			return err
		}
		r = append(r, Subno{
			Counter: uint16(n),
		})
	}
	*sl = r
	return nil
}

type Sub struct {
	SubCounter        uint32
	SubPointerCounter *uint32
	SubFloat64        float64
	SubCounterSlice   []*uint32
}

type TestStruct struct {
	Sub
	Sub1   Sub
	Nested struct {
		NestCounter *uint32 `base:"16"`
	}
	Addr           netip.Addr
	PAddr          *netip.Addr
	BoolVar        bool
	BoolSlice      []bool
	AddrArray      [2]*netip.Addr
	AddrSlice      []*netip.Addr
	AddrNPSlice    []netip.Addr `alias:"anps"`
	Time           time.Time    `layout:"2006 02 Jan 15:04"`
	SNL            SubnoList
	ShouldSkipAddr netip.Addr `skipmarshal:""`
}

type testCase struct {
	input          TestStruct
	Args           []string
	expectedResult TestStruct
	shouldFail     bool
}

func (tc *testCase) do(t *testing.T) error {
	fflag := flag.ExitOnError
	if tc.shouldFail {
		fflag = flag.ContinueOnError
	}
	fs := flag.NewFlagSet("test", fflag)
	filler := myflags.NewFiller()
	filler.Fill(fs, &tc.input)
	err := fs.Parse(tc.Args)
	if err != nil {
		return err
	}
	if !deepEqual(tc.input, tc.expectedResult) {
		return fmt.Errorf("\n%+v is different from expected:\n%+v", myflags.PrettyStruct(tc.input, ""), myflags.PrettyStruct(tc.expectedResult, ""))
	}
	t.Logf("result:\n%v\n", myflags.PrettyStruct(tc.input, ""))
	t.Logf("expected:\n%v\n", myflags.PrettyStruct(tc.expectedResult, ""))
	// fmt.Printf("result:::%+v\n", tc.input)
	// fmt.Printf("expected:::%+v\n", tc.expectedResult)
	return nil
}

func createPAddr(in string) *netip.Addr {
	r := new(netip.Addr)
	*r = netip.MustParseAddr(in)
	return r
}

func TestMyflags(t *testing.T) {
	caseList := []testCase{
		{ //case 0
			input: TestStruct{},
			Args:  []string{"-addr", "1.1.1.1"},
			expectedResult: TestStruct{
				Addr: netip.AddrFrom4([4]byte{1, 1, 1, 1}),
			},
		},
		{ //case 1
			input: TestStruct{},
			Args:  []string{"-addrslice", "1.1.1.1"},
			expectedResult: TestStruct{
				AddrSlice: []*netip.Addr{createPAddr("1.1.1.1")},
			},
		},
		{ //case 2
			input: TestStruct{},
			Args:  []string{"-boolvar"},
			expectedResult: TestStruct{
				BoolVar: true,
			},
		},

		{ //case 3, should fail
			input: TestStruct{},
			Args:  []string{"-addrslice", "1.1.1.1,1.1.1.2"},
			expectedResult: TestStruct{
				AddrSlice: []*netip.Addr{createPAddr("1.1.1.1")},
			},
			shouldFail: true,
		},
		{ //case 4
			input: TestStruct{},
			Args:  []string{"-boolslice", "true,false"},
			expectedResult: TestStruct{
				BoolSlice: []bool{true, false},
			},
		},
		{ //case 5, nega case
			input: TestStruct{},
			Args:  []string{"-boolslice", "true,false"},
			expectedResult: TestStruct{
				BoolSlice: []bool{true, true},
			},
			shouldFail: true,
		},
		{ //case 6
			input: TestStruct{},
			Args:  []string{"-addrarray", "1.1.1.1,1.1.1.2"},
			expectedResult: TestStruct{
				AddrArray: [2]*netip.Addr{createPAddr("1.1.1.1"), createPAddr("1.1.1.2")},
			},
		},
		{ //case 7
			input: TestStruct{},
			Args:  []string{"-nestednestcounter", "99"},
			expectedResult: TestStruct{
				Nested: struct {
					NestCounter *uint32 "base:\"16\""
				}{
					NestCounter: createInt[uint32](0x99),
				},
			},
		},
		{ //case 8
			input: TestStruct{},
			Args:  []string{"-subsubpointercounter", "100", "-subsubcounterslice", "3,4,5"},
			expectedResult: TestStruct{
				Sub: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
		},
		{ //case 9
			input: TestStruct{},
			Args:  []string{"-subsubfloat64", "100.1"},
			expectedResult: TestStruct{
				Sub: Sub{
					SubFloat64: 100.1,
				},
			},
		},

		{ //case 10
			input: TestStruct{},
			Args:  []string{"-sub1subpointercounter", "100", "-sub1subcounterslice", "3,4,5"},
			expectedResult: TestStruct{
				Sub1: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
		},
		{ //case 11
			input: TestStruct{},
			Args:  []string{"-sub1subpointercounter", "100", "-sub1subcounterslice", "5,4,3"},
			expectedResult: TestStruct{
				Sub1: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
			shouldFail: true,
		},
		{ //case 12
			input: TestStruct{},
			Args:  []string{"-anps", "1.1.1.1,2001:dead::beef"},
			expectedResult: TestStruct{
				AddrNPSlice: []netip.Addr{
					*createPAddr("1.1.1.1"),
					*createPAddr("2001:dead::beef"),
				},
			},
		},
		{ //case 13
			input: TestStruct{},
			Args:  []string{"-snl", "9,10,11"},
			expectedResult: TestStruct{
				SNL: []Subno{
					{
						Counter: 9,
					},
					{
						Counter: 10,
					},
					{
						Counter: 11,
					},
				},
			},
		},
		{ //case 14
			input: TestStruct{},
			Args:  []string{"-paddr", "1.1.3.3"},
			expectedResult: TestStruct{
				PAddr: createPAddr("1.1.3.3"),
			},
		},
		{ //case 15
			input: TestStruct{},
			Args:  []string{"-shouldskipaddr", "1.1.3.3"},
			expectedResult: TestStruct{
				ShouldSkipAddr: *createPAddr("1.1.3.3"),
			},
			shouldFail: true,
		},
	}

	for i, c := range caseList {
		// if i != 3 {
		// 	continue
		// }
		t.Logf("testing case %d", i)
		err := c.do(t)
		if err != nil {
			if !c.shouldFail {
				t.Fatalf("case %d failed, %v", i, err)
			} else {
				t.Logf("case %d failed as expected, %v", i, err)
				continue
			}
		}
		t.Logf("testing case %d successfully finished", i)

	}
}

func deepEqual(in, expect any) bool {
	typeIn := reflect.TypeOf(in)
	typeExpect := reflect.TypeOf(expect)
	valIn := reflect.ValueOf(in)
	valExpect := reflect.ValueOf(expect)
	// fmt.Println("in", in, typeIn, "expect", expect, typeExpect, "-"+typeIn.PkgPath()+"-")
	if typeIn != typeExpect {
		return false
	}
	if typeIn.Kind() == reflect.Pointer {
		if valIn == valExpect {
			return true
		}
		if valIn.Elem().IsZero() && valExpect.IsNil() {
			return true
		}
		valIn = valIn.Elem()
		valExpect = valExpect.Elem()
		typeIn = typeIn.Elem()
	}

	switch typeIn.Kind() {
	case reflect.Struct:
		if typeIn.PkgPath() != "github.com/hujun-open/myflags_test" && typeIn.PkgPath() != "" {
			// if !reflect.DeepEqual(valIn.Field(i).Interface(), valExpect.Field(i).Interface()) {
			return fmt.Sprint(valIn.Interface()) == fmt.Sprint(valExpect.Interface())
		}
		for i := 0; i < typeIn.NumField(); i++ {
			// fmt.Printf("field %d %v, %v\n", i, typeIn.Field(i).Name, "=="+typeIn.Field(i).Type.PkgPath()+"==")

			if !deepEqual(valIn.Field(i).Interface(), valExpect.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Array, reflect.Slice:
		if valIn.Len() != valExpect.Len() {
			return false
		}
		for i := 0; i < valIn.Len(); i++ {
			if !deepEqual(valIn.Index(i).Interface(), valExpect.Index(i).Interface()) {
				return false
			}
		}
		return true

	}
	// fmt.Println(3333333333, valIn.Interface(), valExpect.Interface())
	return reflect.DeepEqual(valIn.Interface(), valExpect.Interface())
}
