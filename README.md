[![Go package](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml/badge.svg)](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/myflags)](https://pkg.go.dev/github.com/hujun-open/myflags)
# myflags
myflags is a Golang module to make creating command line application easy, it built on top of [cobra](https://cobra.dev/), provides following **additional** features:

1.  Instead of creating commands, sub-commands and flags manually, user could simply define all the commands/flags in a struct, myflags automatically creates command/flags based on the struct definition and parsed result get automatically assigned to the struct field that corresponding to the flag. 
    - Some common cobra command attribute likes shorthand name, usage, auto-completion choices could be specified as struct field tags 
2. In addition to the types supported by cobra, myflags provides capability to extend support for new types as flag, user could even provide myflags support for existing types without creating alias type
    - this is achieved by implement `RegisteredConverters` interface
    - `types` sub module provides support for some existing golang types like time.Time
3. support any slice/array as flag with element type that implements `RegisteredConverters` interface
4. optional `summaryhelp` command to print usage for the entire command tree

note: current release is v2, the package path is `"github.com/hujun-open/myflags/v2"`



## Quick Start 
Using myflags is straight forward:

1. define all commands/flags in a struct, each command is a sub-struct with tag "action", the value of the tag specifies a struct method gets called when the corresponding command is entered in CLI.
    - a root command method could be optionally specified when creating filler 
2. create a `Filler`, and call `Fill` method with the struct variable with default value. 
3. call one of cobra's command execute method like `Filler.Execute`

Following is an example:
https://github.com/hujun-open/myflags/blob/2635352af91e5628122b0e6077ac4fba0fd20619/example/main.go#L1-L83

the created flags:
```
.\cptool summaryhelp
a zip command
  --backupaddrlist: backup server address list
  -c, --configfile: working profile
        default:default.conf
  --svraddr: server address to download the archive
        default:<nil>
  = completion: Generate the autocompletion script for the specified shell
    = bash: Generate the autocompletion script for bash
      --no-descriptions: disable completion descriptions
        default:false
    = fish: Generate the autocompletion script for fish
      --no-descriptions: disable completion descriptions
        default:false
    = powershell: Generate the autocompletion script for powershell
      --no-descriptions: disable completion descriptions
        default:false
    = zsh: Generate the autocompletion script for zsh
      --no-descriptions: disable completion descriptions
        default:false
  = compress: to compress things
    -l, --loop: number of compress iterations
        default:0x20
    --profile: compress profile
    --skip:
        default:false
    = dry: dry run, doesn't actually create any file
    = zipfile: zip a file
      --filename: specify file name
        default:defaultzip.file
    = zipfolder: zip a folder
      --foldername: specify folder name
  = docgen: generate docs
    --output: output folder
        default:./
    = manpage: generate manpage doc
      --section: manpage section
        default:3
      --title: manpage title
    = markdown: generate markdown doc
  = extract: to unzip things
    --inputfile: input zip file
  = help: Help about any command
  = summaryhelp: help in summary
    -h, --help: help for summaryhelp
        default:false

```
some parsing results:
```
.\cptool --svraddr 1.1.1.1 --backupaddrlist 2.2.2.2,2001:dead::1 compress -l 3  zipfile           
zipfile &{ConfigFile:default.conf SvrAddr:1.1.1.1 BackupAddrList:[2.2.2.2 2001:dead::1] Compress:{Loop:3 Profile: Skip:false NoFlag: DryRun:{} ZipFolder:{FolderName:} ZipFile:{FileName:defaultzip.file}} Extract:{InputFile:}}
command path is cptool compress zipfile
```
## Struct Field Tags
Following struct field tags are supported:

- skipflag: skip the field for flagging
- alias: use the specified alias as the name of the parameter
- usage: the usage string of the parameter
- action: this field is an action, the value is the method name to run
- required: this field is a mandatory required flag
- choices: a comma separated list of value choices for the field, used for auto completion


## Supported Types
Base:
- all int/uint types
- float32/float64
- string
- bool

provided by `github.com/hujun-open/myflags/v2/types`:

- all int/uint types: support `base` tag for the base
- net.HardwareAddr
- net.IPNet
- net.IP
- time.Time
- time.Duration
- all int/uint types

Others:
- All types implement both of following interface:
    - `encoding.TextUnmarshaler`
    - `encoding.TextUnmarshaler`
- All type register via `myflags.Register` function

Note: flag is only created for exported struct field.


In addition to above types, following types are also supported:

- pointer to the type above
- slice/array of type above
- slice/array of pointer to the type above

for slice/array, use "," as separator. 

myflags also supports following type of struct:

- nested struct, like:
```
type Outer struct {
    Nested struct {
        Name string
    }
}
```

- embeded struct, like:
```
type SubS struct {
    SubName string
}
type Outer struct {
    SubS
}
```

## Flag Naming
By default, the name of created flag is the lowercase of struct field name, in case the field is part of sub-struct, and parent struct is NOT an action, then the name of created flag is "<parent_field_name>-<field_name>";
 
Optionally a renaming function could be supplied when creating the `Filler`, myflags uses the renaming function returned string as the flag name.



## Extension
New type could be supported via `myflags.Register`, which takes a variable implements `myflags.RegisteredConverters` interface. the `myflags.Register` must be called before `myflags.Fill`, typically it should be called in `init()`.

Check [types.go](types/types.go), [types_int.go](types/types_int.go) for examples.

## Builtin Commands
Myflags support following builtin commands:

- `help`: same as cobra help command, always included
- `completion`: generate completion script for varies shell, same as cobra, always included
- `docgen`: doc generation command, currently support markdown and manpage, use cobra's doc generation lib, include via `WithDocGenCMD()`
- `summaryhelp`: print usage for whole command tree, include via `WithSummaryHelp()`


## Bool
myflags support bool flag with following format:

- `--flag` (meaning true)
- `--flag=<true|false>`

`--flag <true|false>` is not supported
