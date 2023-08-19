[![Go package](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml/badge.svg)](https://github.com/hujun-open/myflags/actions/workflows/CI.yaml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/myflags)](https://pkg.go.dev/github.com/hujun-open/myflags)
# myflags
myflags is a Golang module to make command line flag easier, it creates command line flags based on fields in a struct, so user doesn't need to manually create the flags one by one. it also supports more types than `flag` module, and easily extensible for new types, user could even add support for existing type without using alias type.

myflags also support multiple and hierarchical actions, each action could have a set of its own parameters and sub-actions, for example a file compression tool `cptool` could have "compress" and "extract" action, and each has different parameters, and "compress" could have sub-actions like "zipfoler" and "zipfile", each then again has different parameters. 

Input wise, parameters of an action has prefix "-", while action doesn't have any prefix.

for example, compress a folder could be command line input like `cptool compress -profile <profile_name> zipfolder -foldername <foldername>`.

## Struct Field Tags
Following struct field tags are supported:

- skipflag: skip the field for flagging
- alias: use the specified alias as the name of the parameter
- usage: the usage string of the parameter
- action: this field is an action 


## Quick Start 
Using myflags is straight forward:

1. define all flags in a struct, each action is a field whose type is another struct.
2. create a `Filler` with the struct, call `Fill` method with the struct variable with default value. 
3. call `flag.Parse()`

Following is an example:
https://github.com/hujun-open/myflags/blob/2fd27463cabdc368b87aecc7addbb42f5535abc6/example/main.go#L1-L45
the created flags:
```
.\cptool.exe -?
a zip command
  - configfile: working profile
        default:default.conf
  = compress: to compress things
    - loop: number of compress iterations
        default:0x20
    - profile:
    - s:
        default:false
    = dryrun: dry run, doesn't actually create any file
    = zipfolder: zip a folder
      - folder: specify folder name
    = zipfile: zip a file
      - f: specify file name
        default:defaultzip.file
  = extract: to unzip things
    - inputfile: input zip file
  = help: help
```
some parsing results:
```
.\cptool.exe -configfile cp.conf compress -profile my.profile -s zipfilder -folder ./bigfolder/
parsed actions [Compress]
{ConfigFile:cp.conf Compress:{Loop:32 Profile:my.profile Skip:true NoFlag: DryRun:{} ZipFolder:{FolderName:} ZipFile:{FileName:defaultzip.file}} Extract:{InputFile:} Help:{}}

.\cptool.exe -configfile cp.conf compress -loop 100 dryrun
parsed actions [Compress DryRun]
{ConfigFile:cp.conf Compress:{Loop:100 Profile: Skip:false NoFlag: DryRun:{} ZipFolder:{FolderName:} ZipFile:{FileName:defaultzip.file}} Extract:{InputFile:} Help:{}}

```



## Supported Types
Base:
- all int/uint types: support `base` tag for the base
- float32/float64
- string
- bool
- time.Duration
- time.Time: supports `layout` tag for time layout string
- All types implement both of following interface:
    - `encoding.TextUnmarshaler`
    - `encoding.TextUnmarshaler`
- All type register via `myflags.Register` function

Note: flag is only created for exported struct field.


In addition to base types, following types are also supported:

- pointer to the base type 
- slice/array of base type
- slice/array of pointer to the base type

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

Check [time.go](time.go), [inttype.go](inttype.go) for examples.

## Bool
myflags use standard Golang module `flag`, [which doesn't support "-flag x" format for bool](https://pkg.go.dev/flag). using "-flag x" for bool could cause silent failure that input parameters after bool don't get parsed.