package main

import (
	"fmt"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"rtRunner/pkg/cfgparser"
	"rtRunner/pkg/hctool"
	"rtRunner/pkg/reporter"
	"rtRunner/pkg/sftptool"
	"runtime"
	"strings"
	"time"
)

func main() {
	var (
		runlst     []string
		sshclient  *ssh.Client
		sftpclient *sftp.Client
		myhost     sftptool.HostInfo
	)
	log.SetLevel(log.ErrorLevel)
	mylog := log.New()
	logfile, err := os.OpenFile("rtRunner.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic("Create Logfile For rtRunner Failed:" + err.Error())
	}
	mylog.SetOutput(logfile)
	mycfg := cfgparser.Cfile{}

	if len(os.Args) > 1 {
		mycfg.Path = os.Args[1]
	} else {
		ostype := runtime.GOOS
		if ostype == "linux" {
			mycfg.Path = "./rtConf.ini"
		} else if ostype == "windows" {
			mycfg.Path = ".\\rtConf.ini"
		}
	}
	mycfg.Initialize(mylog)

	runstr := mycfg.GetStrVal("hosts", "hostlist")
	doclear := mycfg.GetIntVal("hosts", "doclear")
	overwrite := mycfg.GetIntVal("hosts", "overwrite")
	detail := mycfg.GetIntVal("debug", "show_verbose")
	if runstr != "" {
		runlst = strings.Split(mycfg.GetStrVal("hosts", "hostlist"), "|")
	} else {
		panic("At Least One Host in hostlist!")
	}
	localworkdir := mycfg.GetStrVal("hosts", "localdir")
	if !hctool.ChkLocalPath(localworkdir) {
		panic("Local Work Directory Must Be Absolute Path!")
	}
	localDir := filepath.Join(localworkdir, time.Now().Format("20060102"))
	for _, host := range runlst {
		hostinfo := mycfg.GetStrVal("hosts", host)
		sftptool.HostInit(hostinfo, &myhost)
		sftptool.ConfGen(myhost, mylog)
		*myhost.FLST = append(*myhost.FLST, myhost.HCFILE, myhost.CFILE)
		localHostDir := filepath.Join(localDir, myhost.IP)
		if !sftptool.ChkRemotePath(myhost) {
			panic("Remote Work Directory Must Be Absolute Path!")
		}
		myhost.RemoteDIR = hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, time.Now().Format("20060102"))
		myhost.RemoteDIR = hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, "")
		if err != nil {
			panic("Invalid Port")
		}
		sshclient, err = sftptool.SshConnect(myhost)
		if err != nil {
			panic("SSH Connect Failed:" + err.Error())
		}
		sftpclient, err = sftptool.SftpConnect(sshclient)
		if err != nil {
			panic("SFTP Connect Failed:" + err.Error())
		}
		fmt.Printf(">>>>>> Working on Host:%s <<<<<<<\n", myhost.IP)
		sftptool.DmHC_Chk(myhost)
		sftptool.MkRemoteDir(sftpclient, myhost)
		if overwrite == 0 {
			sftptool.ChkDirEmpty(sftpclient, myhost)
		}
		sftptool.Upload(sftpclient, myhost, detail)
		sftptool.RunHC(sshclient, myhost, detail)
		sftptool.Download(sftpclient, localHostDir, myhost, detail)
		reporter.SimpleGen(localHostDir, myhost.SimpleNo)
		err := os.Remove(myhost.CFILE)
		if err != nil {
			return
		}
		if doclear > 0 {
			sftptool.RemoveHC(sftpclient, myhost, detail)
		} else {
			fmt.Printf("Please Remove %s Manully!\n", myhost.RemoteDIR)
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
