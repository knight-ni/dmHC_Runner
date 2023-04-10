package reporter

import (
	"gitee.com/jiewen/gooxml/document"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	dt  = "20060102"
	dt2 = "2006年01月02日"
)

func SimpleGen(localdir string, simpno int) {
	tmplst, err := os.ReadDir(localdir)
	if err != nil {
		panic("Read Local Docx Failed " + err.Error())
	}
	filter := regexp.MustCompile(`.docx`)
	for _, v := range tmplst {
		if v.IsDir() {
			continue
		} else if strings.HasPrefix(v.Name(), "Simple") || strings.HasPrefix(v.Name(), "~") {
			continue
		} else {
			fname := filter.FindString(v.Name())
			if fname != "" {

				doc, err := document.Open(filepath.Join(localdir, v.Name()))
				if err != nil {
					panic("error opening document:" + err.Error())
				}

				//paragraphs := []document.Paragraph{}
				//for _, p := range doc.Paragraphs() {
				//	paragraphs = append(paragraphs, p)
				//}
				if err != nil {
					panic("Invalid Simple Number!")
				}
				doc.X().Body.EG_BlockLevelElts = doc.X().Body.EG_BlockLevelElts[:simpno]
				//
				//dflag := 0
				//for _, p := range paragraphs {
				//	for _, r := range p.Runs() {
				//		if strings.Contains(r.Text(), "前10 大小对象信息") {
				//			dflag += 1
				//		}
				//	}
				//
				//	if dflag > 0 {
				//
				//
				//
				//	}
				//}
				doc.SaveToFile(filepath.Join(localdir, "Simple-"+v.Name()))
			}
		}
	}

}
