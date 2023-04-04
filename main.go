package main

import (
	"fmt"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"rtRunner/pkg/cfgparser"
	"rtRunner/pkg/hctool"
	"rtRunner/pkg/sftptool"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		runlst     []string
		sshclient  *ssh.Client
		sftpclient *sftp.Client
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
	detail := mycfg.GetIntVal("debug", "show_db_func")
	if runstr != "" {
		runlst = strings.Split(mycfg.GetStrVal("hosts", "hostlist"), "|")
	} else {
		panic("At Least One Host in hostlist!")
	}
	remoteworkdir := mycfg.GetStrVal("hosts", "remotedir")
	localworkdir := mycfg.GetStrVal("hosts", "localdir")
	if !hctool.ChkLocalPath(localworkdir) {
		panic("Local Work Directory Must Be Absolute Path!")
	}
	if !hctool.ChkRemotePath(remoteworkdir) {
		panic("Remote Work Directory Must Be Absolute Path!")
	}
	remoteDir := path.Join(remoteworkdir, time.Now().Format("20060102"))
	localDir := path.Join(localworkdir, time.Now().Format("20060102"))
	for _, host := range runlst {
		hostinfo := mycfg.GetStrVal("hosts", host)
		mydict := hctool.HostParse(hostinfo)
		localHostDir := path.Join(localDir, mydict["ip"]+"_"+mydict["cmode"])
		port, err := strconv.Atoi(mydict["port"])
		if err != nil {
			panic("Invalid Port")
		}
		sshclient, err = sftptool.SshConnect(mydict["usr"], mydict["pwd"], mydict["ip"], port)
		if err != nil {
			panic("SSH Connect Failed:" + err.Error())
		}
		sftpclient, err = sftptool.SftpConnect(sshclient)
		if err != nil {
			panic("SFTP Connect Failed:" + err.Error())
		}
		fmt.Printf(">>>>>> Working on Host:%s <<<<<<<\n", host)
		sftptool.Upload(sftpclient, remoteDir, mydict["hcfile"], mydict["cfile"], detail)
		sftptool.RunHC(sshclient, remoteDir, mydict["hcfile"], mydict["cfile"], detail)
		sftptool.Download(sftpclient, remoteDir, localHostDir, detail)
		sftptool.RemoveHC(sftpclient, remoteDir, detail)
		//sftptool.RemoveHC(sftpclient, remoteworkdir, detail)
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
