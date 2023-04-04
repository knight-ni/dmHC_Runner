package hctool

import (
	"fmt"
	"net"
	"path"
	"path/filepath"
	"rtRunner/pkg/ostool"
	"strings"
)

func dmHC_Sel(taros string, tarcpu string) string {
	var hcfile string
	if taros == "linux" {
		hcfile = fmt.Sprintf("dmHC_%s_%s", taros, tarcpu)
	} else {
		hcfile = fmt.Sprintf("dmHC_%s_%s.exe", taros, tarcpu)
	}
	return hcfile
}

func DmHC_Chk(fname string) bool {
	if !ostool.Exists(fname) {
		fmt.Printf("%s does not exist!\n", fname)
		return false
	} else {
		return true
	}
}

func HostParse(hostinfo string) map[string]string {
	tmpstr := strings.Split(hostinfo, "|")
	var mydict = make(map[string]string)
	ip, port, err := net.SplitHostPort(tmpstr[0])
	if err != nil {
		fmt.Println("Invalid IP or Port")
	}
	os := tmpstr[1]
	cpu := tmpstr[2]
	usr := tmpstr[3]
	pwd := tmpstr[4]
	cmode := tmpstr[5]
	hcfile := dmHC_Sel(os, cpu)

	mydict["ip"] = ip
	mydict["port"] = port
	mydict["os"] = os
	mydict["cpu"] = cpu
	mydict["usr"] = usr
	mydict["pwd"] = pwd
	mydict["hcfile"] = hcfile
	mydict["cmode"] = cmode
	mydict["cfile"] = fmt.Sprintf("conf_%s.ini", cmode)
	return mydict
}

func ChkLocalPath(mypath string) bool {
	if filepath.IsAbs(mypath) {
		return true
	}
	return false
}

func ChkRemotePath(mypath string) bool {
	if path.IsAbs(mypath) {
		return true
	}
	return false
}
