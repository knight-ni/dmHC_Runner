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

func SimpleGen(localdir string) {
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
					panic("Error Opening Document:" + err.Error())
				}
				doc.X().Body.EG_BlockLevelElts = doc.X().Body.EG_BlockLevelElts[:100]
				doc.SaveToFile(filepath.Join(localdir, "Simple-"+v.Name()))
			}
		}
	}
}
