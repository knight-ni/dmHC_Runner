package hctool

import (
	"fmt"
	"path/filepath"
	"strings"
)

var Support_CPU = []string{"amd64", "arm64", "ppc64le"}
var Support_OS = []string{"linux", "windows"}

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
	var sep rune
	if osname == "linux" {
		sep = '/'
		if strings.HasSuffix(path1, "/") {
			return path1 + path2
		} else {
			return path1 + string(sep) + path2
		}
	} else if osname == "windows" {
		sep = '\\'
		if strings.HasSuffix(path1, "\\") {
			return path1 + path2
		} else {
			return path1 + string(sep) + path2
		}
	}
	panic("Unknown Os!")
}
