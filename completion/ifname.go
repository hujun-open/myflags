package completion

import (
	"net"
	"strings"

	"github.com/hujun-open/cobra"
)

func InterfaceNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	rs := []string{}
	ifs, err := net.Interfaces()
	if err != nil {
		return rs, cobra.ShellCompDirectiveError
	}
	for _, inf := range ifs {
		if strings.HasPrefix(inf.Name, toComplete) {
			rs = append(rs, inf.Name)
		}
	}
	return rs, cobra.ShellCompDirectiveNoFileComp
}
