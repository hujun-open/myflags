package main

import (
	"fmt"

	"github.com/hujun-open/myflags"
)

type ZipCLI struct {
	ConfigFile string `usage:"working profile"`
	Compress   struct {
		Profile   string
		Skip      bool     `alias:"s"`   //use "s" as the parameter name
		NoFlag    string   `skipflag:""` //ignore this field for flagging
		DryRun    struct{} `usage:"dry run, doesn't actually create any file"`
		ZipFolder struct {
			FolderName string `alias:"folder" usage:"specify folder name"`
		} `usage:"zip a folder"`
		ZipFile struct {
			FileName string `alias:"f" usage:"specify file name"`
		} `usage:"zip a file"`
	} `usage:"to compress things"`
	Extract struct {
		InputFile string `usage:"input zip file"`
	} `usage:"to unzip things"`
	Help struct{} `usage:"help"`
}

func main() {
	filler := myflags.NewFiller("zipcli", "a zip command")
	zipcli := ZipCLI{
		ConfigFile: "default.conf",
	}
	zipcli.Compress.ZipFile.FileName = "defaultzip.file"
	err := filler.Fill(&zipcli)
	if err != nil {
		panic(err)
	}
	acts, err := filler.Parse()
	if err != nil {
		panic(err)
	}
	fmt.Println("parsed actions", acts)
	fmt.Printf("%+v\n", zipcli)
}
