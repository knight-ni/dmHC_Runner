package docx2pdf

import (
	"dmHC_Runner/pkg/hctool"
	"dmHC_Runner/pkg/sftptool"
	"fmt"
	"github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/document/convert"
	unipdflicense "github.com/unidoc/unipdf/v3/common/license"
	"github.com/wxnacy/wgo/file"
	"regexp"
	"strings"
)

const licfile = "dmHC.lic"

func init() {
	apiKey, err := file.ReadFile(licfile)
	if apiKey == "" {
		fmt.Errorf("missing License")
	}

	err = unipdflicense.SetMeteredKey(apiKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to set metered key: %v\n", err)
		fmt.Printf("Make sure to get a valid key from https://cloud.unidoc.io\n")
		fmt.Printf("If you don't have one - Grab one in the Free Tier at https://cloud.unidoc.io\n")
		panic(err)
	}

	// This example requires both for unioffice and unipdf.
	err = license.SetMeteredKey(apiKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to set metered key: %v\n", err)
		fmt.Printf("Make sure to get a valid key from https://cloud.unidoc.io\n")
		fmt.Printf("If you don't have one - Grab one in the Free Tier at https://cloud.unidoc.io\n")
		panic(err)
	}
}

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
