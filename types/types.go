// Package types provide myflags support for following golang types:
//
//   - net.HardwareAddr
//   - net.IPNet
//   - net.IP
//   - time.Time
//   - time.Duration
//
// this package could be used by simply importing it, e.g. `import _ "github.com/hujun-open/myflags/types"`
package types

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hujun-open/myflags/v2"
)

// FlagConverter is used to facilitate Register() by create a RegisteredConverters instance with From/To assigned to corresponding function
type FlagConverter struct {
	From myflags.FromStrFunc
	To   myflags.ToStrFunc
}

// ToStr implements RegisteredConverters interface
func (fc *FlagConverter) ToStr(in any, tag reflect.StructTag) string {
	return fc.To(in, tag)
}

// FromStr implements RegisteredConverters interface
func (fc *FlagConverter) FromStr(s string, tag reflect.StructTag) (any, error) {
	return fc.From(s, tag)
}

func init() {
	myflags.Register[net.HardwareAddr](&FlagConverter{
		From: macFromStr,
		To:   macToStr,
	})
	myflags.Register[net.IPNet](&FlagConverter{
		From: ipnetFromStr,
		To:   ipnetToStr,
	})
	myflags.Register[net.IP](&FlagConverter{
		From: ipFromStr,
		To:   ipToStr,
	})
	myflags.Register[time.Time](&FlagConverter{
		From: timeFromStr,
		To:   timeToStr,
	})
	myflags.Register[time.Duration](&FlagConverter{
		From: durationFromStr,
		To:   durationToStr,
	})

}

func macFromStr(text string, tag reflect.StructTag) (any, error) {
	if text == "" {
		return net.HardwareAddr{}, nil
	}
	var r = make([]byte, 6)
	var flist []string
	switch {
	case strings.Contains(text, "-"):
		flist = strings.Split(text, "-")
	case strings.Contains(text, ":"):
		flist = strings.Split(text, ":")
	default:
		return nil, fmt.Errorf("can't find supported MAC format")
	}
	for i, v := range flist {
		x, err := strconv.ParseInt(strings.TrimSpace(v), 16, 64)
		if err != nil {
			return nil, fmt.Errorf("%v is not valid byte value in hex", v)
		}
		if x >= 255 {
			return nil, fmt.Errorf("%v is not valid byte value in hex, should be <256", v)
		}
		r[i] = byte(x)
	}
	return net.HardwareAddr(r), nil

}

// just the net.Hardware.String()
func macToStr(in any, tag reflect.StructTag) string {
	v := in.(net.HardwareAddr)
	return v.String()
}

func ipnetFromStr(text string, tag reflect.StructTag) (any, error) {
	_, r, err := net.ParseCIDR(text)
	return *r, err
}

func ipnetToStr(in any, tag reflect.StructTag) string {
	v := in.(net.IPNet)
	return v.String()
}

func ipFromStr(text string, tag reflect.StructTag) (any, error) {
	addr := net.ParseIP(text)
	if addr == nil {
		return nil, fmt.Errorf("%v is not a valid IP address", text)
	}
	return addr, nil
}

func ipToStr(in any, tag reflect.StructTag) string {
	return in.(net.IP).String()
}

// DefaultTimeLayout is the default layout string to parse time, following golang time.Parse() format,
// can be overridden per field by field tag "layout". Default value is "2006-01-02 15:04:05", which is
// the same as time.DateTime in Go 1.20
var DefaultTimeLayout = "2006-01-02 15:04:05"

func timeToStr(in any, tag reflect.StructTag) string {
	layout, _ := tag.Lookup("layout")
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return in.(time.Time).Format(layout)
}

func timeFromStr(s string, tag reflect.StructTag) (any, error) {
	layout, _ := tag.Lookup("layout")
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return time.Parse(layout, strings.TrimSpace(s))
}

func durationToStr(in any, tag reflect.StructTag) string {
	return fmt.Sprint(in)
}
func durationFromStr(s string, tag reflect.StructTag) (any, error) {
	return time.ParseDuration(s)
}
