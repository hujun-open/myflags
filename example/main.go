package main

import (
	"fmt"

	"github.com/hujun-open/cobra"
	"github.com/hujun-open/myflags/v2"
)

type ZipCLI struct {
	ConfigFile string `short:"c" usage:"working profile" required:""`
	Compress   struct {
		Loop      uint   `base:"16" short:"l" usage:"number of compress iterations"`
		Profile   string `usage:"compress profile" choices:"p1,p2,p3"`
		Skip      bool
		NoFlag    string   `skipflag:""` //ignore this field for flagging
		DryRun    struct{} `alias:"dry" usage:"dry run, doesn't actually create any file" action:"Dry"`
		ZipFolder struct {
			FolderName string `usage:"specify folder name"`
		} `usage:"zip a folder" action:"ZipFolder"`
		ZipFile struct {
			FileName string `usage:"specify file name"`
		} `usage:"zip a file" action:"ZipFile"`
	} `usage:"to compress things" action:"Comp"`
	Extract struct {
		InputFile string `usage:"input zip file"`
	} `usage:"to unzip things" action:"Extr"`
}

func (zipcli *ZipCLI) Dry(cmd *cobra.Command, args []string) {
	fmt.Printf("dryrun %+v\n", zipcli)
}

func (zipcli *ZipCLI) ZipFolder(cmd *cobra.Command, args []string) {
	fmt.Printf("zipfolder %+v\n", zipcli)
}
func (zipcli *ZipCLI) ZipFile(cmd *cobra.Command, args []string) {
	fmt.Printf("zipfile %+v\n", zipcli)
}
func (zipcli *ZipCLI) Extr(cmd *cobra.Command, args []string) {
	fmt.Printf("extract %+v\n", zipcli)
}

func (zipcli *ZipCLI) Comp(cmd *cobra.Command, args []string) {
	fmt.Printf("compress %+v\n", zipcli)
}

func main() {
	//create a new filler with the application name and its description.
	filler := myflags.NewFiller("cptool", "a zip command")
	//some default values
	zipcli := ZipCLI{
		ConfigFile: "default.conf",
	}
	zipcli.Compress.Loop = 0x20
	zipcli.Compress.ZipFile.FileName = "defaultzip.file"
	//call Fill
	err := filler.Fill(&zipcli)
	if err != nil {
		panic(err)
	}
	//call Execute to fill zipcli with parsed values from input
	err = filler.Execute()
	if err != nil {
		panic(err)
	}
}
