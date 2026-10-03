package myflags

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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

func (docgen *genDoc) DoMarkdown(cmd *cobra.Command, args []string) error {
	err := doc.GenMarkdownTree(cmd.Root(), docgen.Output)
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}
	return nil
}

func (docgen *genDoc) DoManpage(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("failed to generate manpage: %w", err)
	}
	return nil
}

func defGenDoc() *genDoc {
	r := new(genDoc)
	r.Output = "./"
	r.Manpage.Section = 3
	return r
}

// DocgenCMDName is the optional command to generate docs
const DocgenCMDName = "docgen"

func newDocFiller() (*Filler, error) {
	f := NewFiller(DocgenCMDName, "generate docs")
	if err := f.Fill(defGenDoc()); err != nil {
		return nil, err
	}
	return f, nil
}

// include doc generation command
func WithDocGenCMD() FillerOption {
	return func(filler *Filler) {
		filler.includeDocGenCMD = true
	}
}
