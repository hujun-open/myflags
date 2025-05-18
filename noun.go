package myflags

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/hujun-open/cobra"
	"golang.org/x/exp/slices"
)

// parseNoun is set as command's PreRunE method so that it could parse args for the nouns
func (filler *Filler) parseNoun(args []string) error {
	var inV reflect.Value
	var ok bool
	var index uint

	for i, arg := range args {
		index = uint(i + 1)
		if inV, ok = filler.nounVals[index]; !ok {
			return fmt.Errorf("noun id %d not found in struct definition", i)
		}

		if inV.Kind() != reflect.Pointer {
			inV = inV.Addr()
		}
		//check if it is a registered type, a.k.a simpleType
		conv := globalRegistry.GetViaInterface(inV.Interface())
		if conv != nil {
			val, err := conv.FromStr(arg, filler.nounFields[index].Tag)
			if err != nil {
				return err
			}
			inV.Elem().Set(reflect.ValueOf(val))
			continue

		}
		//check if it implements UnmarshalText
		if txtUnmarshal, ok := inV.Interface().(encoding.TextUnmarshaler); ok {
			err := txtUnmarshal.UnmarshalText([]byte(arg))
			if err != nil {
				return err
			}
			continue
		}
		//check if it is arrary/slice
		switch filler.nounVals[index].Kind() {
		case reflect.Slice, reflect.Array:
			process := false
			if globalRegistry.GetViaType(filler.nounVals[index].Type().Elem()) != nil {
				process = true
			}

			if filler.nounVals[index].Type().Elem().Kind() == reflect.Pointer {
				if filler.nounVals[index].Type().Elem().Implements(textEncodingInt) {
					//list of pointer to textmarshalce
					process = true
				}
			} else {
				if reflect.PointerTo(filler.nounVals[index].Type().Elem()).Implements(textEncodingInt) {
					//list of textmarshalce
					process = true
				}
			}
			if process {
				newlist, err := getListType(inV, filler.nounFields[index].Tag)
				if err != nil {
					return err
				}
				err = newlist.Set(arg)
				if err != nil {
					return err
				}
				continue
			}

		}
	}
	return nil
}

// getActNounCompleter return a completer that completes one or multiple nouns
func (filler *Filler) getActNounCompleter() cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		filler.parseNoun(args) //no need to check return error since this is just for completion
		index := len(args) + 1
		if cf, ok := filler.nounCompleters[uint(index)]; ok {
			return cf(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// SetPreRun set f as filler.Commmand.PreRun method;
// don't set filler.Command.PreRun directly
func (filler *Filler) SetPreRun(f func(cmd *cobra.Command, args []string)) {
	newf := func(cmd *cobra.Command, args []string) {
		if len(filler.nounVals) > 0 {
			filler.parseNoun(args)
		}
		f(cmd, args)
	}
	filler.PreRun = newf
}

// SetPreRun set f as filler.Commmand.PreRunE method;
// don't set filler.Command.PreRunE directly
func (filler *Filler) SetPreRunE(f func(cmd *cobra.Command, args []string) error) {
	newf := func(cmd *cobra.Command, args []string) error {
		if len(filler.nounVals) > 0 {
			err := filler.parseNoun(args)
			if err != nil {
				return err
			}
		}
		return f(cmd, args)
	}
	filler.PreRunE = newf
}

func (filler *Filler) getSortedNounIDs() []uint {
	idList := []uint{}
	for key := range filler.nounFields {
		idList = append(idList, key)
	}
	slices.Sort(idList)
	return idList
}

func (filler *Filler) getUse() string {
	idList := filler.getSortedNounIDs()
	useStr := filler.Use
	for _, id := range idList {
		useStr += fmt.Sprintf(" <%v>", filler.nounFields[id].Name)
	}
	return useStr

}

// this replaces cobra default usageFunc to add noun part
func (filler *Filler) usageFunc(c *cobra.Command) error {
	fmt.Print("Usage:")
	if c.Runnable() {
		fmt.Printf("\n  %s", c.UseLine())
	}

	if c.HasAvailableSubCommands() {
		fmt.Printf("\n  %s [command]", c.CommandPath())
	}
	if len(filler.nounFields) > 0 {
		fmt.Printf("\n\nArguments:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		for _, id := range filler.getSortedNounIDs() {
			defStr := fmt.Sprintf("%s", filler.nounVals[id].Interface())
			conv := globalRegistry.GetViaInterface(filler.nounVals[id].Interface())
			if conv != nil {
				defStr = conv.ToStr(filler.nounVals[id].Interface(), filler.nounFields[id].Tag)
			}
			fmt.Fprintf(w, "\n  <%v> %v\t%v (default \"%s\")",
				filler.nounFields[id].Name,
				filler.nounFields[id].Type,
				filler.nounFields[id].Tag.Get(UsageTag),
				defStr,
			)
		}
		w.Flush()
	}
	if len(c.Aliases) > 0 {
		fmt.Printf("\n\nAliases:\n")
		fmt.Printf("  %s", c.NameAndAliases())
	}
	if c.HasExample() {
		fmt.Printf("\n\nExamples:\n")
		fmt.Printf("%s", c.Example)
	}
	if c.HasAvailableSubCommands() {
		cmds := c.Commands()
		if len(c.Groups()) == 0 {
			fmt.Printf("\n\nAvailable Commands:")
			for _, subcmd := range cmds {
				if subcmd.IsAvailableCommand() || subcmd.Name() == "help" {
					fmt.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
				}
			}
		} else {
			for _, group := range c.Groups() {
				fmt.Printf("\n\n%s", group.Title)
				for _, subcmd := range cmds {
					if subcmd.GroupID == group.ID && (subcmd.IsAvailableCommand() || subcmd.Name() == "help") {
						fmt.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
			if !c.AllChildCommandsHaveGroup() {
				fmt.Printf("\n\nAdditional Commands:")
				for _, subcmd := range cmds {
					if subcmd.GroupID == "" && (subcmd.IsAvailableCommand() || subcmd.Name() == "help") {
						fmt.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
		}
	}
	if c.HasAvailableLocalFlags() {
		fmt.Printf("\n\nFlags:\n")
		fmt.Print(trimRightSpace(c.LocalFlags().FlagUsages()))
	}
	if c.HasAvailableInheritedFlags() {
		fmt.Printf("\n\nGlobal Flags:\n")
		fmt.Print(trimRightSpace(c.InheritedFlags().FlagUsages()))
	}
	if c.HasHelpSubCommands() {
		fmt.Printf("\n\nAdditional help topics:")
		for _, subcmd := range c.Commands() {
			if subcmd.IsAdditionalHelpTopicCommand() {
				fmt.Printf("\n  %s %s", rpad(subcmd.CommandPath(), subcmd.CommandPathPadding()), subcmd.Short)
			}
		}
	}
	if c.HasAvailableSubCommands() {
		fmt.Printf("\n\nUse \"%s [command] --help\" for more information about a command.", c.CommandPath())
	}
	fmt.Println()
	return nil
}
func rpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}
func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
