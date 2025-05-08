[![Go package](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml/badge.svg)](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/myflags)](https://pkg.go.dev/github.com/hujun-open/myflags/v2)
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

1. define all commands/flags in a struct, each command is a sub-struct with tag `action`, the value of the tag specifies a struct method gets called when the corresponding command is entered in CLI. 
    - a root command method could be optionally specified when creating filler with `WithRootMethod`
    - command could be nested, e.g. subcommand could have its subcommand, which is nested struct with `action` tag
2. create a `Filler`, and call `Fill` method with the struct variable with default value. 
3. call one of cobra's command execute method like `Filler.Execute`

Following is an example:
https://github.com/hujun-open/myflags/blob/2635352af91e5628122b0e6077ac4fba0fd20619/example/main.go#L1-L83

the created flags:
```
.\cptool summaryhelp
  = cptool <Arg1> [flags]
    <Arg1>: arg for root command
      default:""
    --backupaddrlist: backup server address list
    -c, --configfile: working profile
        default:default.conf
    --svraddr: server address to download the archive
        default:<nil>
    = cptool compress
      -l, --loop: number of compress iterations
        default:0x20
      --profile: compress profile
      --skip:
        default:false
      = cptool compress dry
      = cptool compress zipfile <FileName> <ArchiveName>
        <FileName>: input file name
          default:"defaultzip.file"
        <ArchiveName>: output archive name
          default:""
      = cptool compress zipfolder <FolderName> <ArchiveName> <CreationTime>
        <FolderName>: input folder name
          default:""
        <ArchiveName>: output archive name
          default:""
        <CreationTime>: creation time
          default:"2025 02 Jan 03:04"
    = cptool extract <InputFile> <OutputFolder>
      <InputFile>: input archive file
        default:""
      <OutputFolder>: output folder
        default:""
    = cptool completion
      = cptool completion bash
        --no-descriptions: disable completion descriptions
                default:false
      = cptool completion fish [flags]
        --no-descriptions: disable completion descriptions
                default:false
      = cptool completion powershell [flags]
        --no-descriptions: disable completion descriptions
                default:false
      = cptool completion zsh [flags]
        --no-descriptions: disable completion descriptions
                default:false
    = cptool docgen
      --output: output folder
        default:./
      = cptool docgen manpage
        --section: manpage section
                default:3
        --title: manpage title
      = cptool docgen markdown
    = cptool help [command]
    = cptool summaryhelp [flags]
      -h, --help: help for summaryhelp
        default:false
```
some parsing results:
```
.\cptool --svraddr 1.1.1.1 --backupaddrlist 2.2.2.2,2001:dead::1 compress -l 3  zipfile  input1 out.zip         
zipfile &{ConfigFile:default.conf SvrAddr:1.1.1.1 BackupAddrList:[2.2.2.2 2001:dead::1] Arg1: Compress:{Loop:3 Profile: Skip:false NoFlag: DryRun:{} ZipFolder:{FolderName: ArchiveName: CreationTime:2025-01-02 03:04:05 +0000 UTC} ZipFile:{FileName:input1 ArchiveName:out.zip}} Extract:{InputFile: OutputFolder:}}

.\cptool compress zipfolder folder1 out.zip "2030 01 Jun 13:01" -l 99
zipfolder &{ConfigFile:default.conf SvrAddr:<nil> BackupAddrList:[] Arg1: Compress:{Loop:99 Profile: Skip:false NoFlag: DryRun:{} ZipFolder:{FolderName:folder1 ArchiveName:out.zip CreationTime:2030-06-01 13:01:00 +0000 UTC} ZipFile:{FileName:defaultzip.file ArchiveName:}} Extract:{InputFile: OutputFolder:}}
```


## Struct Field Tags
Following struct field tags are supported:

- noun: mark the field as postional argument, the value is the index; e.g. 1 means first argument, 2 is 2nd argument ..etc
- skipflag: skip the field for flagging
- alias: use the specified alias as the name of the parameter
- short: use the specified string as the shorthand parameter name
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

## Positional Argument
Positional arguments are the struct field with "noun" tag, the value of the tag is the postional index, start from 1. positional argument only get parsed with a command, which means in case of root command positional argument only get parsed when filler is created with `WithRootMethod`.

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
