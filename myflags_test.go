package myflags_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
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

func newApp(t *testing.T, in any, opts ...myflags.FillerOption) *myflags.Filler {
	t.Helper()
	all := []myflags.FillerOption{
		myflags.WithFlagErrHandling(flag.ContinueOnError),
		myflags.WithRootMethod(func(cmd *cobra.Command, args []string) {}),
	}
	all = append(all, opts...)
	f := myflags.NewFiller("app", "test", all...)
	if err := f.Fill(in); err != nil {
		t.Fatal(err)
	}
	f.SetOut(io.Discard)
	f.SetErr(io.Discard)
	return f
}

func noPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("panicked: %v", rec)
		}
	}()
	fn()
}

func captureStdout(t *testing.T, fn func()) (out string) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	defer func() {
		os.Stdout = old
		_ = w.Close()
		<-done
		_ = r.Close()
		out = buf.String()
	}()
	fn()
	return out
}

func TestFloatParseError(t *testing.T) {
	t.Run("scalar", func(t *testing.T) {
		cfg := struct {
			F float64
		}{F: 1.5}
		f := newApp(t, &cfg)
		f.SetArgs([]string{"--f", "nope"})
		var err error
		noPanic(t, func() { err = f.Execute() })
		if err == nil {
			t.Fatalf("invalid float was accepted, value=%v", cfg.F)
		}
		if cfg.F != 1.5 {
			t.Fatalf("F=%v, want the original 1.5", cfg.F)
		}
	})
	t.Run("slice", func(t *testing.T) {
		cfg := struct {
			Vals []float64
		}{}
		f := newApp(t, &cfg)
		f.SetArgs([]string{"--vals", "1.5,nope"})
		var err error
		noPanic(t, func() { err = f.Execute() })
		if err == nil {
			t.Fatalf("invalid float slice was accepted, value=%v", cfg.Vals)
		}
	})
}

func TestIPNetRejectsInvalidCIDR(t *testing.T) {
	cfg := struct {
		CIDR net.IPNet
	}{}
	f := newApp(t, &cfg)
	f.SetArgs([]string{"--cidr", "not-a-cidr"})
	var err error
	noPanic(t, func() { err = f.Execute() })
	if err == nil {
		t.Fatalf("invalid CIDR was accepted: %v", cfg.CIDR)
	}
}

func TestMACParser(t *testing.T) {
	parse := func(t *testing.T, text string) (net.HardwareAddr, error) {
		t.Helper()
		cfg := struct {
			MAC net.HardwareAddr
		}{}
		f := newApp(t, &cfg)
		f.SetArgs([]string{"--mac", text})
		var err error
		noPanic(t, func() { err = f.Execute() })
		return cfg.MAC, err
	}

	t.Run("ff", func(t *testing.T) {
		mac, err := parse(t, "aa:bb:cc:dd:ee:ff")
		if err != nil {
			t.Fatal(err)
		}
		want := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
		if mac.String() != want.String() {
			t.Fatalf("MAC=%s, want %s", mac, want)
		}
	})
	t.Run("short", func(t *testing.T) {
		mac, err := parse(t, "aa:bb:cc")
		if err == nil {
			t.Fatalf("short MAC was accepted: %s", mac)
		}
	})
	t.Run("long", func(t *testing.T) {
		_, err := parse(t, "11:22:33:44:55:66:77")
		if err == nil {
			t.Fatal("overlong MAC was accepted")
		}
	})
}

type bugNounAct struct {
	Path string `noun:"1" usage:"path"`
}

type bugNounRoot struct {
	Act bugNounAct `action:""`
}

func TestNounCompletionWithoutCompleter(t *testing.T) {
	f := newApp(t, &bugNounRoot{})
	cmd := f.GetChildCommand("/act")
	if cmd == nil || cmd.ValidArgsFunction == nil {
		t.Fatal("act command has no argument completer")
	}
	noPanic(t, func() {
		_, _ = cmd.ValidArgsFunction(cmd, []string{}, "p")
	})
}

type bugReqInner struct {
	Name string `required:"" usage:"name"`
}

type bugReqRoot struct {
	Inner bugReqInner
}

func TestRequiredOnNestedStruct(t *testing.T) {
	cfg := bugReqRoot{}
	f := newApp(t, &cfg)
	fl := f.PersistentFlags().Lookup("inner-name")
	if fl == nil {
		t.Fatal("flag inner-name was not created")
	}
	if _, ok := fl.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Fatal("nested field Inner.Name was not marked required")
	}
	if err := f.Execute(); err == nil {
		t.Fatal("Execute succeeded without required flag inner-name")
	}
}

type bugChoiceInner struct {
	Color string `choices:"red,blue" usage:"color"`
}

type bugChoiceRoot struct {
	Inner bugChoiceInner
}

func TestChoicesOnNestedStruct(t *testing.T) {
	f := newApp(t, &bugChoiceRoot{})
	if f.PersistentFlags().Lookup("inner-color") == nil {
		t.Fatal("flag inner-color was not created")
	}
	fn, ok := f.GetFlagCompletionFunc("inner-color")
	if !ok {
		t.Fatal("choices on nested field Inner.Color were not registered")
	}
	got, _ := fn(f.Command, nil, "")
	if strings.Join(got, ",") != "red,blue" {
		t.Fatalf("completions=%v", got)
	}
}

type bugSliceChoice struct {
	Colors []string `choices:"red,blue" usage:"colors"`
}

func TestChoicesOnSlice(t *testing.T) {
	cfg := bugSliceChoice{}
	f := myflags.NewFiller("app", "test")
	err := f.Fill(&cfg)
	if err == nil {
		t.Fatal("expected an error for choices on a slice")
	}
	if !strings.Contains(err.Error(), "choices") || !strings.Contains(err.Error(), "Colors") {
		t.Fatalf("error = %v", err)
	}
}

type bugBox struct {
	Color string `complete:"Colors" usage:"color"`
}

func (b *bugBox) Colors(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"red"}, cobra.ShellCompDirectiveNoFileComp
}

type bugBoxRoot struct {
	Box bugBox `action:""`
}

func TestCompleteMethodOnNestedStructIsNotUsed(t *testing.T) {
	cfg := bugBoxRoot{}
	f := myflags.NewFiller("app", "test", myflags.WithFlagErrHandling(flag.ContinueOnError))
	err := f.Fill(&cfg)
	if err == nil {
		t.Fatal("expected complete method on the nested struct to be ignored")
	}
	if !strings.Contains(err.Error(), "Colors") {
		t.Fatalf("error = %v", err)
	}
}

type bugBadCompleteRoot struct {
	Color string `complete:"Colors" usage:"color"`
}

func (r *bugBadCompleteRoot) Colors() {}

func TestCompleteMethodWrongSignature(t *testing.T) {
	cfg := bugBadCompleteRoot{}
	f := myflags.NewFiller("app", "test")
	var err error
	noPanic(t, func() { err = f.Fill(&cfg) })
	if err == nil {
		t.Fatal("expected an error for a complete method with the wrong signature")
	}
}

type bugBadActionRoot struct {
	Child struct{} `action:"Go"`
}

func (r *bugBadActionRoot) Go() {}

func TestActionMethodWrongSignature(t *testing.T) {
	cfg := bugBadActionRoot{}
	f := myflags.NewFiller("app", "test")
	var err error
	noPanic(t, func() { err = f.Fill(&cfg) })
	if err == nil {
		t.Fatal("expected an error for an action method with the wrong signature")
	}
}

type bugNestedComplete struct {
	Color string `complete:"Colors" usage:"color"`
}

type bugNestedCompleteRoot struct {
	Nested bugNestedComplete
}

func (r *bugNestedCompleteRoot) Colors(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"red"}, cobra.ShellCompDirectiveNoFileComp
}

func TestCompleteOnNestedStructNotRegistered(t *testing.T) {
	f := newApp(t, &bugNestedCompleteRoot{})
	if f.PersistentFlags().Lookup("nested-color") == nil {
		t.Fatal("flag nested-color was not created")
	}
	fn, ok := f.GetFlagCompletionFunc("nested-color")
	if !ok {
		t.Fatal("complete method for nested field Nested.Color was not registered")
	}
	got, _ := fn(f.Command, nil, "")
	if strings.Join(got, ",") != "red" {
		t.Fatalf("completions=%v", got)
	}
}

func TestArrayRejectsTooManyElements(t *testing.T) {
	cfg := struct {
		Names [2]string
	}{}
	f := newApp(t, &cfg)
	f.SetArgs([]string{"--names", "a,b,c"})
	var err error
	noPanic(t, func() { err = f.Execute() })
	if err == nil {
		t.Fatalf("too many array elements were accepted: %q", cfg.Names)
	}
}

func TestNilPointerNoun(t *testing.T) {
	cfg := struct {
		Name *string `noun:"1" usage:"name"`
	}{}
	f := newApp(t, &cfg)
	f.SetArgs([]string{"hello"})
	var err error
	noPanic(t, func() { err = f.Execute() })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name == nil || *cfg.Name != "hello" {
		t.Fatalf("Name=%v, want hello", cfg.Name)
	}
}

func TestPointerNounUsage(t *testing.T) {
	when := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg := struct {
		When *time.Time `noun:"1" usage:"when"`
	}{When: &when}
	f := newApp(t, &cfg)
	var buf bytes.Buffer
	f.SetOut(&buf)
	f.SetErr(&buf)
	if err := f.UsageFunc()(f.Command); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `default "2020-01-02 03:04:05"`) {
		t.Fatalf("usage=%q", buf.String())
	}

	summary := myflags.SummaryFillerUsageStr(f, "")
	if !strings.Contains(summary, `default:"2020-01-02 03:04:05"`) {
		t.Fatalf("summary=%q", summary)
	}
}

type bugRec struct {
	A string
	B string
}

func (r bugRec) MarshalText() ([]byte, error) {
	return []byte(r.A + "/" + r.B), nil
}

func (r *bugRec) UnmarshalText(text []byte) error {
	a, b, found := strings.Cut(string(text), "/")
	r.A = a
	if found {
		r.B = b
	}
	return nil
}

func TestTextListElementsAreIndependent(t *testing.T) {
	cfg := struct {
		Items []bugRec
	}{}
	f := newApp(t, &cfg)
	f.SetArgs([]string{"--items", "a/b,c"})
	if err := f.Execute(); err != nil {
		t.Fatal(err)
	}
	want := []bugRec{{A: "a", B: "b"}, {A: "c", B: ""}}
	if len(cfg.Items) != len(want) {
		t.Fatalf("got %+v, want %+v", cfg.Items, want)
	}
	for i := range want {
		if cfg.Items[i] != want[i] {
			t.Fatalf("got %+v, want %+v", cfg.Items, want)
		}
	}
}

func TestDocgenFailureIsReturned(t *testing.T) {
	var cfg struct{}
	f := newApp(t, &cfg, myflags.WithDocGenCMD())
	t.Cleanup(func() {
		cmd := f.GetChildCommand("/docgen")
		if cmd == nil {
			return
		}
		if fl := cmd.Flags().Lookup("output"); fl != nil {
			_ = fl.Value.Set("./")
		}
	})
	bad := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(bad, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.SetArgs([]string{"docgen", "markdown", "--output", bad})
	var execErr error
	out := captureStdout(t, func() {
		execErr = f.Execute()
	})
	if execErr == nil {
		t.Fatalf("docgen reported success after generation failed, output=%q", out)
	}
}

func TestDocgenCommandIsPerFiller(t *testing.T) {
	var first, second struct{}
	fa := newApp(t, &first, myflags.WithDocGenCMD())
	fb := newApp(t, &second, myflags.WithDocGenCMD())
	da := fa.GetChildCommand("/docgen")
	db := fb.GetChildCommand("/docgen")
	if da == nil || db == nil {
		t.Fatalf("docgen missing: first=%v second=%v", da != nil, db != nil)
	}
	if da == db {
		t.Fatal("fillers share one docgen command")
	}
	if da.Parent() != fa.Command || db.Parent() != fb.Command {
		t.Fatal("docgen command parent does not match the filler that created it")
	}
}

func TestUsageHonorsSetOut(t *testing.T) {
	cfg := struct {
		Name string `usage:"the name"`
	}{}
	f := newApp(t, &cfg)
	var buf bytes.Buffer
	f.SetOut(&buf)
	f.SetErr(&buf)
	f.SetArgs([]string{"--help"})
	_ = f.Execute()
	if !strings.Contains(buf.String(), "the name") {
		t.Fatalf("help output=%q", buf.String())
	}
}

func TestEmptyChildPath(t *testing.T) {
	var cfg struct{}
	f := newApp(t, &cfg)
	t.Run("filler", func(t *testing.T) {
		var got *myflags.Filler
		noPanic(t, func() { got = f.GetChildFiller("") })
		if got != nil {
			t.Fatalf("GetChildFiller returned %#v", got)
		}
	})
	t.Run("command", func(t *testing.T) {
		var got *cobra.Command
		noPanic(t, func() { got = f.GetChildCommand("") })
		if got != nil {
			t.Fatalf("GetChildCommand returned %#v", got)
		}
	})
}

func TestIntNoun(t *testing.T) {
	in := struct {
		Count int `noun:"1" usage:"count"`
	}{}
	f := newApp(t, &in)
	f.SetArgs([]string{"42"})
	if err := f.Execute(); err != nil {
		t.Fatal(err)
	}
	if in.Count != 42 {
		t.Fatalf("Count=%d, want 42", in.Count)
	}
}
