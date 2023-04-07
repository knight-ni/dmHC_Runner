package reporter

import (
	"gitee.com/jiewen/gooxml/document"
	"os"
	"path/filepath"
	"regexp"
)

const (
	dt  = "20060102"
	dt2 = "2006年01月02日"
)

func ReadPara(localdir string) {
	tmplst, err := os.ReadDir(localdir)
	if err != nil {
		panic("Read Local Docx Failed " + err.Error())
	}
	filter := regexp.MustCompile(`.docx`)
	for _, v := range tmplst {
		if v.IsDir() {
			continue
		} else {
			fname := filter.FindString(v.Name())
			if fname != "" {
				doc, err := document.Open(filepath.Join(localdir, v.Name()))
				if err != nil {
					panic(err.Error())
				}
				for _, par := range doc.Paragraphs() { //读取文档类所有段落
					for _, r := range par.Runs() {
						par.RemoveRun(r)
					}
				}
			}
		}
	}

}
