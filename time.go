package myflags

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func init() {
	Register[time.Time](&timeConvertor{})
	Register[time.Duration](new(durationType))
}

// DefaultTimeLayout is the default layout string to parse time, following golang time.Parse() format,
// can be overridden per field by field tag "layout". Default value is "2006-01-02 15:04:05", which is
// the same as time.DateTime in Go 1.20
var DefaultTimeLayout = "2006-01-02 15:04:05"

type timeConvertor struct{}

func (tc *timeConvertor) ToStr(in any, tag reflect.StructTag) string {
	layout, _ := tag.Lookup("layout")
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return in.(time.Time).Format(layout)
}

func (tc *timeConvertor) FromStr(s string, tag reflect.StructTag) (any, error) {
	layout, _ := tag.Lookup("layout")
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return time.Parse(layout, strings.TrimSpace(s))
}

type durationType time.Duration

func (d *durationType) ToStr(in any, tag reflect.StructTag) string {
	return fmt.Sprint(in)
}
func (d *durationType) FromStr(s string, tag reflect.StructTag) (any, error) {
	return time.ParseDuration(s)
}
