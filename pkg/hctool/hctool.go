package hctool

import (
	"fmt"
	"path/filepath"
)

func DmHC_Sel(taros string, tarcpu string) string {
	var hcfile string
	if taros == "linux" {
		hcfile = fmt.Sprintf("dmHC_%s_%s", taros, tarcpu)
	} else {
		hcfile = fmt.Sprintf("dmHC_%s_%s.exe", taros, tarcpu)
	}
	return hcfile
}

func ChkLocalPath(mypath string) bool {
	if filepath.IsAbs(mypath) {
		return true
	}
	return false
}

func SmartPathJoin(osname string, path1 string, path2 string) string {
	var sep string
	if osname == "linux" {
		sep = "/"
		return path1 + sep + path2
	} else if osname == "windows" {
		sep = "/" //FOR SOME SPECIAL REASON
		return path1 + sep + path2
	}
	panic("Unknown Os!")
}
