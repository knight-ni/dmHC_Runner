package main

import (
	"dmHC_Runner/pkg/cfgparser"
	"dmHC_Runner/pkg/hctool"
	"dmHC_Runner/pkg/sftptool"
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
		sshclient, err = sftptool.SshConnect(myhost)
		if err != nil {
			panic("SSH Connect Failed:" + err.Error())
		}
		sftpclient, err = sftptool.SftpConnect(sshclient)
		if err != nil {
			panic("SFTP Connect Failed:" + err.Error())
		}
		fmt.Printf(">>>>>> Working On Instance: %s:%d <<<<<<<\n", myhost.IP, myhost.DB_PORT)

		sftptool.MkRemoteDir(sftpclient, myhost)
		if overwrite == 0 {
			sftptool.ChkDirEmpty(sftpclient, myhost)
		}
		sftptool.DmHC_Chk(myhost)
		sftptool.Upload(sftpclient, myhost, detail)
		sftptool.RunHC(sshclient, myhost, detail)
		sftptool.Download(sftpclient, localHostDir, myhost, detail)

		if doclear > 0 {
			sftptool.RemoveHC(sftpclient, myhost, detail)
		} else {
			fmt.Printf("Please Remove %s Manully!\n\n", myhost.RemoteDIR)
		}
		err = sftpclient.Close()
		if err != nil {
			panic("SFTP Connect Failed:" + err.Error())
		}
		err = sshclient.Close()
		if err != nil {
			panic("SSH Connect Failed:" + err.Error())
		}
	}
}
