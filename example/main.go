package main

import (
	"fmt"

	"github.com/hujun-open/myflags"
	"github.com/spf13/cobra"
)

type ZipCLI struct {
	ConfigFile string `usage:"working profile"`
	Compress   struct {
		Loop      uint `base:"16" short:"l" usage:"number of compress iterations" required:""`
		Profile   string
		Skip      bool     `alias:"skip"` //use "skip" as the parameter name
		NoFlag    string   `skipflag:""`  //ignore this field for flagging
		DryRun    struct{} `usage:"dry run, doesn't actually create any file" action:"" method:"Dry"`
		ZipFolder struct {
			FolderName string `alias:"folder" usage:"specify folder name"`
		} `usage:"zip a folder" action:"" method:"ZipFolder"`
		ZipFile struct {
			FileName string `alias:"f" usage:"specify file name"`
		} `usage:"zip a file" action:"" method:"ZipFile"`
	} `usage:"to compress things" action:"" method:"Comp"`
	Extract struct {
		InputFile string `usage:"input zip file"`
	} `usage:"to unzip things" action:"" method:"Extr"`
	Help struct{} `usage:"help" action:"" method:"H"`
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
func (zipcli *ZipCLI) H(cmd *cobra.Command, args []string) {
	fmt.Printf("help %+v\n", zipcli)
}

func main() {
	filler := myflags.NewFiller("cptool", "a zip command")
	//some default values
	zipcli := ZipCLI{
		ConfigFile: "default.conf",
	}
	zipcli.Compress.Loop = 0x20
	zipcli.Compress.ZipFile.FileName = "defaultzip.file"
	err := filler.Fill(&zipcli)
	if err != nil {
		panic(err)
	}
	err = filler.Exec()
	if err != nil {
		panic(err)
	}

	// filler.PrintDebug()
	// acts, err := filler.Parse()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("parsed actions", acts)
	// fmt.Printf("%+v\n", zipcli)
	// if acts[0] == myflags.CompleteCMDName {
	// 	script, err := filler.GenCompletionScript(myflags.ShellBash)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("xxx")
	// 	fmt.Println(script)
	// }
}
