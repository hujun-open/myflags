package myflags

import (
	"strings"

	"github.com/hujun-open/cobra"
)

// validFlagValues is a helper struct used to support ValidValuesTag
type validFlagValues struct {
	values []string
}

// strlist is a list of options, separated by comma
func newValidFlagValues(strlist string) *validFlagValues {
	slist := strings.FieldsFunc(strlist, func(r rune) bool { return r == ',' })
	rlist := []string{}
	for i := range slist {
		if val := strings.TrimSpace(slist[i]); val != "" {
			rlist = append(rlist, val)
		}
	}
	return &validFlagValues{values: rlist}
}

// complete implements cobra flag completion function
func (helper *validFlagValues) complete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return helper.values, cobra.ShellCompDirectiveNoFileComp
}
