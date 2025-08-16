package myflags_test

import (
	"fmt"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hujun-open/cobra"
	flag "github.com/spf13/pflag"

	"github.com/hujun-open/myflags/v2"
	_ "github.com/hujun-open/myflags/v2/types"
	"golang.org/x/exp/slices"
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
	Arg1 string    `noun:"1" usage:"arg1 is a string"`
	Arg2 time.Time `noun:"2" usage:"arg2 is a time"`
	Sub
	Sub1 Sub `action:""`
	Act1 struct {
		Act1Counter *uint32 `base:"16"`
		Act11       struct {
			Act11Counter int
		} `action:""`
	} `action:""`
	Addr           netip.Addr
	PAddr          *netip.Addr
	BoolVar        bool
	BoolSlice      []bool
	AddrArray      [2]*netip.Addr
	AddrSlice      []*netip.Addr
	AddrNPSlice    []netip.Addr ``
	Time           time.Time    `layout:"2006 02 Jan 15:04"`
	SNL            SubnoList
	ShouldSkipAddr netip.Addr `skipflag:""`
}

type testCase struct {
	input          TestStruct
	errh           flag.ErrorHandling
	Args           []string
	expectedResult TestStruct
	expectedActs   []string
	shouldFail     bool
}

func (tc *testCase) do(t *testing.T) error {
	fflag := tc.errh
	filler := myflags.NewFiller(
		"test", "", myflags.WithFlagErrHandling(fflag),
		myflags.WithRootMethod(func(cmd *cobra.Command, args []string) {}), //without this, root command's positional arg won't get parsed
	)
	err := filler.Fill(&tc.input)
	if err != nil {
		t.Fatal(err)
	}
	filler.SetArgs(tc.Args)
	cmd, err := filler.ExecuteC()

	// parsedActs, err := filler.ParseArgs(tc.Args)
	if err != nil {
		return err
	}
	pathList := strings.Fields(cmd.CommandPath())[1:]
	if !slices.Equal(pathList, tc.expectedActs) {
		return fmt.Errorf("parsed acts %v is different from expected acts %v", pathList, tc.expectedActs)
	}
	if !deepEqual(tc.input, tc.expectedResult) {
		return fmt.Errorf("\n%+v is different from expected:\n%+v", myflags.PrettyStruct(tc.input, ""), myflags.PrettyStruct(tc.expectedResult, ""))
	}
	t.Logf("result:\n%v\n", myflags.PrettyStruct(tc.input, ""))
	t.Logf("expected:\n%v\n", myflags.PrettyStruct(tc.expectedResult, ""))
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
			Args:  []string{"--addr", "1.1.1.1"},
			expectedResult: TestStruct{
				Addr: netip.AddrFrom4([4]byte{1, 1, 1, 1}),
			},
		},
		{ //case 1
			input: TestStruct{},
			Args:  []string{"--addrslice", "1.1.1.1,2001:dead::beef"},
			expectedResult: TestStruct{
				AddrSlice: []*netip.Addr{createPAddr("1.1.1.1"), createPAddr("2001:dead::beef")},
			},
		},
		{ //case 2
			input: TestStruct{},
			Args:  []string{"--boolvar"},
			expectedResult: TestStruct{
				BoolVar: true,
			},
		},

		{ //case 3, should fail
			input: TestStruct{},
			Args:  []string{"--addrslice", "1.1.1.1,1.1.1.2"},
			expectedResult: TestStruct{
				AddrSlice: []*netip.Addr{createPAddr("1.1.1.1")},
			},
			shouldFail: true,
		},
		{ //case 4
			input: TestStruct{},
			Args:  []string{"--boolslice", "true,false"},
			expectedResult: TestStruct{
				BoolSlice: []bool{true, false},
			},
		},
		{ //case 5, nega case
			input: TestStruct{},
			Args:  []string{"--boolslice", "true,false"},
			expectedResult: TestStruct{
				BoolSlice: []bool{true, true},
			},
			shouldFail: true,
		},
		{ //case 6
			input: TestStruct{},
			Args:  []string{"--addrarray", "1.1.1.1,2001:dead::beef"},
			expectedResult: TestStruct{
				AddrArray: [2]*netip.Addr{createPAddr("1.1.1.1"), createPAddr("2001:dead::beef")},
			},
		},
		{ //case 7
			input: TestStruct{},
			Args:  []string{"act1", "--act1counter", "0x99"},
			expectedResult: TestStruct{
				Act1: struct {
					Act1Counter *uint32 `base:"16"`
					Act11       struct {
						Act11Counter int
					} `action:""`
				}{
					Act1Counter: createInt[uint32](0x99),
				},
			},
			expectedActs: []string{"act1"},
		},
		{ //case 8
			input: TestStruct{},
			Args:  []string{"--sub-subpointercounter", "100", "--sub-subcounterslice", "3,4,5"},
			expectedResult: TestStruct{
				Sub: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
			expectedActs: []string{},
		},
		{ //case 9
			input: TestStruct{},
			Args:  []string{"--sub-subfloat64", "100.1"},
			expectedResult: TestStruct{
				Sub: Sub{
					SubFloat64: 100.1,
				},
			},
			expectedActs: []string{},
		},

		{ //case 10
			input: TestStruct{},
			Args:  []string{"sub1", "--subpointercounter", "100", "--subcounterslice", "3,4,5"},
			expectedResult: TestStruct{
				Sub1: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
			expectedActs: []string{"sub1"},
		},
		{ //case 11
			input: TestStruct{},
			Args:  []string{"sub1", "--subpointercounter", "100", "--subcounterslice", "5,4,3"},
			expectedResult: TestStruct{
				Sub1: Sub{
					SubPointerCounter: createInt[uint32](100),
					SubCounterSlice:   []*uint32{createInt[uint32](3), createInt[uint32](4), createInt[uint32](5)},
				},
			},
			expectedActs: []string{"sub1"},
			shouldFail:   true,
		},
		{ //case 12
			input: TestStruct{},
			Args:  []string{"--addrnpslice", "1.1.1.1, 2001:dead::beef "},
			expectedResult: TestStruct{
				AddrNPSlice: []netip.Addr{
					*createPAddr("1.1.1.1"),
					*createPAddr("2001:dead::beef"),
				},
			},
		},
		{ //case 13
			input: TestStruct{},
			Args:  []string{"--snl", "9,10,11"},
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
			Args:  []string{"--paddr", "1.1.3.3"},
			expectedResult: TestStruct{
				PAddr: createPAddr("1.1.3.3"),
			},
		},
		{ //case 15
			input: TestStruct{},
			Args:  []string{"--shouldskipaddr", "1.1.3.3"},
			expectedResult: TestStruct{
				ShouldSkipAddr: *createPAddr("1.1.3.3"),
			},
			shouldFail: true,
		},
		{ //case 16
			input: TestStruct{},
			Args:  []string{"act1", "--act1counter", "0x99", "act11", "--act11counter", "199"},
			expectedResult: TestStruct{
				Act1: struct {
					Act1Counter *uint32 `base:"16"`
					Act11       struct {
						Act11Counter int
					} `action:""`
				}{
					Act1Counter: createInt[uint32](0x99),
					Act11: struct{ Act11Counter int }{
						Act11Counter: 199,
					},
				},
			},
			expectedActs: []string{"act1", "act11"},
		},
		{ //case 17
			input: TestStruct{},
			Args:  []string{"act1", "--act1counter", "99", "act121", "--act1counter", "199"},
			expectedResult: TestStruct{
				Act1: struct {
					Act1Counter *uint32 `base:"16"`
					Act11       struct {
						Act11Counter int
					} `action:""`
				}{
					Act1Counter: createInt[uint32](0x99),
					Act11: struct{ Act11Counter int }{
						Act11Counter: 199,
					},
				},
			},
			expectedActs: []string{"act1"},
			shouldFail:   true,
		},
		{ //case 18, should fail
			input: TestStruct{},
			Args:  []string{"--xxxxx", "1.1.1.1, 1.1.1.2 "},
			expectedResult: TestStruct{
				AddrSlice: []*netip.Addr{createPAddr("1.1.1.1")},
			},
			shouldFail: true,
		},
		{ //case 19, testing 0x
			input: TestStruct{},
			Args:  []string{"act1", "--act1counter", "0x99"},
			expectedResult: TestStruct{
				Act1: struct {
					Act1Counter *uint32 `base:"16"`
					Act11       struct {
						Act11Counter int
					} `action:""`
				}{
					Act1Counter: createInt[uint32](0x99),
				},
			},
			expectedActs: []string{"act1"},
		},
		{ //case 20, action with globla bool
			input: TestStruct{},
			Args:  []string{"--boolvar", "act1", "--act1counter", "0x99"},
			expectedResult: TestStruct{
				BoolVar: true,
				Act1: struct {
					Act1Counter *uint32 `base:"16"`
					Act11       struct {
						Act11Counter int
					} `action:""`
				}{
					Act1Counter: createInt[uint32](0x99),
				},
			},
			expectedActs: []string{"act1"},
			shouldFail:   false,
		},
		{ //case 21, nouns
			input: TestStruct{},
			errh:  flag.PanicOnError,
			Args:  []string{"disk", "2002-03-04 11:22:33"},
			expectedResult: TestStruct{
				Arg1: "disk",
				Arg2: time.Date(2002, 3, 4, 11, 22, 33, 0, time.UTC),
			},
			expectedActs: []string{},
			shouldFail:   false,
		},
		{ //case 22, 2nd noun default
			input: TestStruct{
				Arg2: time.Date(2002, 3, 4, 11, 22, 33, 0, time.UTC),
			},
			errh: flag.PanicOnError,
			Args: []string{"disk"},
			expectedResult: TestStruct{
				Arg1: "disk",
				Arg2: time.Date(2002, 3, 4, 11, 22, 33, 0, time.UTC),
			},
			expectedActs: []string{},
			shouldFail:   false,
		},
		{ //case 23, noun all default
			input: TestStruct{
				Arg1: "defArg1",
				Arg2: time.Date(2002, 3, 4, 11, 22, 33, 0, time.UTC),
			},
			errh: flag.PanicOnError,
			Args: []string{},
			expectedResult: TestStruct{
				Arg1: "defArg1",
				Arg2: time.Date(2002, 3, 4, 11, 22, 33, 0, time.UTC),
			},
			expectedActs: []string{},
			shouldFail:   false,
		},
	}

	for i, c := range caseList {
		// if i != 20 {
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
		} else {
			if c.shouldFail {
				t.Fatalf("case %d should fail but succeed", i)
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
		if typeIn.PkgPath() != "github.com/hujun-open/myflags/v2_test" && typeIn.PkgPath() != "" {
			fmt.Println(typeIn.PkgPath())
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
