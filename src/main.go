package main

import (
	"common/cfgparser"
	"common/docx2pdf"
	"common/hctool"
	"common/sftptool"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func pause() {
	b := make([]byte, 1)
	fmt.Println("Press Any Key To Exit:")
	_, _ = os.Stdin.Read(b)
}

func main() {
	var (
		err        error
		runlst     []string
		sshclient  *ssh.Client
		sftpclient *sftp.Client
		myhost     sftptool.HostInfo
		dir        string
	)
	defer pause()
	mycfg := cfgparser.Cfile{}

	if len(os.Args) > 1 {
		dir, err = filepath.Abs(filepath.Dir(os.Args[1]))
		if err != nil {
			panic("Get Exec Dir Failed!")
		}
		mycfg.Path = os.Args[1]
	} else {
		dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic("Get Exec Dir Failed!")
		}
		mycfg.Path = filepath.Join(dir, "rtConf.ini")
	}
	mycfg.Initialize()

	runstr := mycfg.GetStrVal("base", "hostlist")
	doclear := mycfg.GetIntVal("base", "doclear")
	overwrite := mycfg.GetIntVal("base", "overwrite")
	convpdf := mycfg.GetIntVal("base", "convert2pdf")
	detail := mycfg.GetIntVal("debug", "show_verbose")
	if runstr != "" {
		runlst = strings.Split(runstr, "|")
	} else {
		panic("At Least One Host in hostlist!")
	}

	localDir := filepath.Join(dir, time.Now().Format("20060102"))
	for _, host := range runlst {
		sftptool.HostInit(host, dir, mycfg, &myhost)
		sftptool.ConfGen(myhost, dir)
		defer os.Remove(filepath.Join(dir, myhost.CFILE))
		localHostDir := filepath.Join(localDir, fmt.Sprintf("%s_%d_%d", myhost.IP, myhost.DB_PORT, myhost.SimpleNo))
		myhost.RemoteDIR = hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, time.Now().Format("20060102"))
		myhost.RemoteDIR = hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, "")
		myhost.RemoteDIROut = myhost.RemoteDIR[:len(myhost.RemoteDIR)-1]
		if !sftptool.IsLocalIP(myhost.IP) {
			sshclient, err = sftptool.SshConnect(myhost)
			if err != nil {
				panic("SSH Connect Failed:" + err.Error())
			}
			sftpclient, err = sftptool.SftpConnect(sshclient)
			if err != nil {
				panic("SFTP Connect Failed:" + err.Error())
			}
			fmt.Printf(">>>>>> Working On Instance: %s:%d <<<<<<<\n", myhost.IP, myhost.DB_PORT)

			if !sftptool.ChkRemoteDirExists(sftpclient, myhost) {
				sftptool.MkRemoteDir(sftpclient, myhost)
			} else if !sftptool.ChkDirEmpty(sftpclient, myhost) && overwrite == 0 {
				panic("Remote Directory Not Empty!")
			}
			sftptool.DmHC_Chk(myhost)
			sftptool.Upload(sftpclient, myhost, detail)
			sftptool.RunHC(sshclient, myhost, detail)
			sftptool.Download(sftpclient, localHostDir, myhost, detail)
			if convpdf > 0 {
				err := docx2pdf.WordToPDF(myhost, localHostDir, detail)
				if err != nil {
					fmt.Printf("Convert Docx to PDF Failed! Error:%s\n", err.Error())
				}
			}
			if doclear > 0 {
				//	sftptool.RemoveHC(sftpclient, myhost, detail)
			} else {
				fmt.Printf("Please Remove %s Manully!\n\n", myhost.RemoteDIROut)
			}
			err = sftpclient.Close()
			if err != nil {
				panic("SFTP Connect Failed:" + err.Error())
			}
			err = sshclient.Close()
			if err != nil {
				panic("SSH Connect Failed:" + err.Error())
			}
		} else {
			fmt.Printf(">>>>>> Working On Instance: %s:%d <<<<<<<\n", myhost.IP, myhost.DB_PORT)
			if !sftptool.ChkLocalHostDirExists(myhost) {
				sftptool.MkLocalHostDir(myhost)
			} else if !sftptool.ChkLocalHostDirEmpty(myhost) && overwrite == 0 {
				panic("Remote Directory Not Empty!")
			}
			sftptool.DmHC_Chk(myhost)
			sftptool.LocalHostUpload(myhost, detail)
			sftptool.RunLocalHostHC(myhost, detail)
			sftptool.LocalHostDownload(localHostDir, myhost, detail)
			if convpdf > 0 {
				err := docx2pdf.WordToPDF(myhost, localHostDir, detail)
				if err != nil {
					fmt.Printf("Convert Docx to PDF Failed! Error:%s\n", err.Error())
				}
			}
			if doclear > 0 {
				//	sftptool.RemoveHC(sftpclient, myhost, detail)
			} else {
				fmt.Printf("Please Remove %s Manully!\n\n", myhost.RemoteDIROut)
			}
		}
	}
}
