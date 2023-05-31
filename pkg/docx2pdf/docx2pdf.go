package docx2pdf

import (
	"dmHC_Runner/pkg/hctool"
	"dmHC_Runner/pkg/sftptool"
	"fmt"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/document/convert"
	"regexp"
	"strings"
)

func WordToPDF(myhost sftptool.HostInfo, localdir string, detail int) error {
	var docxname string
	var pdfname string
	filter := regexp.MustCompile(`.docx`)
	for _, f := range *myhost.FLST {
		if filter.FindString(f) != "" {
			docxname = hctool.SmartPathJoin(myhost.OS, localdir, f)
			doc, err := document.Open(docxname)
			if err != nil {
				return fmt.Errorf("open Docx File %s Failed", docxname)
			}
			defer doc.Close()
			pdfname = strings.TrimSuffix(docxname, ".docx") + ".pdf"
			c := convert.ConvertToPdf(doc)
			err = c.WriteToFile(pdfname)
			if err != nil {
				return fmt.Errorf("write PDF File %s Failed", pdfname)
			}
		}
	}
	if detail > 0 {
		fmt.Printf(">>>>>> Converting Docx File %s <<<<<<<\n", docxname)
	}
	return nil
}
