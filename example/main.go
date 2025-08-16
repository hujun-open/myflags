package main

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/hujun-open/cobra"
	"github.com/hujun-open/myflags/v2"
	_ "github.com/hujun-open/myflags/v2/types"
)

type ZipCLI struct {
	ConfigFile     string       `short:"c" usage:"working profile"`
	SvrAddr        net.IP       `usage:"server address to download the archive" complete:"SvrComplete"` //using completer method SvrComplete for shell completion
	BackupAddrList []netip.Addr `short:"b" usage:"backup server address list"`
	IntList        []float32    `short:"i" usage:"list of numbers"`
	StrList        []string     `usage:"list of string"`
	Compress       struct {
		Loop      uint   `base:"16" short:"l" usage:"number of compress iterations"`
		Profile   string `usage:"compress profile" choices:"p1,p2,p3"` //using choices for shell compeltion
		Skip      bool
		NoFlag    string   `skipflag:""` //ignore this field for flagging
		DryRun    struct{} `alias:"dry" usage:"dry run, doesn't actually create any file" action:"Dry"`
		ZipFolder struct {
			FolderName   string    `noun:"1" usage:"input folder name"`   //first positional argument for command zipfolder
			ArchiveName  string    `noun:"2" usage:"output archive name"` //2nd postional argument for command zipfolder
			CreationTime time.Time `noun:"3" usage:"creation time" layout:"2006 02 Jan 15:04"`
		} `usage:"zip a folder" action:"ZipFolder"`
		ZipFile struct {
			FileName    string `noun:"1" usage:"input file name"`     //first positional argument for command zipfile
			ArchiveName string `noun:"2" usage:"output archive name"` //2nd postional argument for command zipfile
		} `usage:"zip a file" action:"ZipFile"`
	} `usage:"to compress things" action:"Comp"`
	Extract struct {
		InputFile    string `noun:"1" usage:"input archive file" complete:"ExtractInputComplete"`
		OutputFolder string `noun:"2" usage:"output folder"`
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

func (zipcli *ZipCLI) RootCMD(cmd *cobra.Command, args []string) {
	fmt.Printf("root %+v\n", zipcli)
}

func (zipcli *ZipCLI) SvrComplete(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"1.1.1.1", "2.2.2.2", "3.3.3.3"}, cobra.ShellCompDirectiveKeepOrder

}

func (zipcli *ZipCLI) ExtractInputComplete(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"ab", "cd", "ef"}, cobra.ShellCompDirectiveKeepOrder

}

func main() {
	//some default values
	zipcli := ZipCLI{
		ConfigFile: "default.conf",
		StrList:    []string{"a", "bb", "ccc"},
	}
	zipcli.BackupAddrList = []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("2.2.2.2")}
	zipcli.Compress.Loop = 0x20
	zipcli.IntList = []float32{1.7, 2.2, 3.3}
	zipcli.Compress.ZipFile.FileName = "defaultzip.file"
	zipcli.Compress.ZipFolder.CreationTime = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	//create a filler with the application name and its description.
	filler := myflags.NewFiller("cptool", "a zip command",
		myflags.WithShellCompletionCMD(),
		myflags.WithDocGenCMD(),
		myflags.WithSummaryHelp(),
		myflags.WithRootMethod(zipcli.RootCMD),
	)

	//call Fill
	err := filler.Fill(&zipcli)
	if err != nil {
		panic(err)
	}
	//call Excute or any other Execute method
	err = filler.Execute()
	if err != nil {
		panic(err)
	}
}
