package myflags

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hujun-open/cobra"
	"github.com/hujun-open/cobra/doc"
)

type genDoc struct {
	Output   string `usage:"output folder"`
	Markdown struct {
	} `usage:"generate markdown doc" action:"DoMarkdown"`
	Manpage struct {
		Section uint   `usage:"manpage section"`
		Title   string `usage:"manpage title"`
	} `usage:"generate manpage doc" action:"DoManpage"`
}

func (docgen *genDoc) DoMarkdown(cmd *cobra.Command, args []string) {
	err := doc.GenMarkdownTree(cmd.Root(), docgen.Output)
	if err != nil {
		fmt.Println("failed to generate markdown", err)
	}
}

func (docgen *genDoc) DoManpage(cmd *cobra.Command, args []string) {
	t := cmd.Root().Name()
	if docgen.Manpage.Title != "" {
		t = docgen.Manpage.Title
	}
	header := &doc.GenManHeader{
		Title:   t,
		Section: strconv.Itoa(int(docgen.Manpage.Section)),
	}
	err := doc.GenManTree(cmd.Root(), header, docgen.Output)
	if err != nil {
		fmt.Println("failed to generate manpage", err)
	}
}

func defGenDoc() *genDoc {
	r := new(genDoc)
	r.Output = "./"
	r.Manpage.Section = 3
	return r
}

// DocgenCMDName is the optional command to generate docs
const DocgenCMDName = "docgen"

var genDocSetup *genDoc
var docFiller *Filler

func init() {
	genDocSetup = defGenDoc()
	docFiller = NewFiller(DocgenCMDName, "generate docs")
	err := docFiller.Fill(genDocSetup)
	if err != nil {
		log.Fatal(err)
		docFiller = nil
		return
	}

}

// include doc generation command
func WithDocGenCMD() FillerOption {
	return func(filler *Filler) {
		filler.includeDocGenCMD = true
	}
}
