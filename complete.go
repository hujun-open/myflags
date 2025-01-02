package myflags

import (
	"bytes"
	"fmt"

	"github.com/hujun-open/cobra"
)

type ShellType string

const (
	ShellBash ShellType = "bash"
)

func (filler *Filler) GenCompletionScript(cmd *cobra.Command, args []string) {
	buf := bytes.NewBuffer([]byte{})
	var err error
	shell := ShellBash
	if len(args) > 0 {
		//use default bash
		shell = ShellType(args[0])
	}
	switch shell {
	case ShellBash:
		err = filler.GenBashCompletionV2(buf, false)
		if err != nil {
			fmt.Printf("failed to generate BASHv2 completion script, %v\n", err)
			return
		}
		fmt.Println(buf.String())
		return
	default:
		fmt.Printf("unsupported shell type %v\n", shell)
		return
	}

}
