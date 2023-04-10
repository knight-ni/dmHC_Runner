package sftptool

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"net/url"
	"os"
	"regexp"
	"rtRunner/pkg/cfgparser"
	"rtRunner/pkg/hctool"
	"rtRunner/pkg/ostool"
	"strconv"
	"strings"
	"time"
)

type HostInfo struct {
	IP        string
	SSH_PORT  int
	DB_PORT   int
	DM_HOME   string
	OS        string
	CPU       string
	SSH_USR   string
	SSH_PWD   string
	DB_USR    string
	DB_PWD    string
	HCFILE    string
	CFILE     string
	FLST      *[]string
	RemoteDIR string
	SimpleNo  int
}

func HostInit(hostinfo string, myhost *HostInfo) {
	var err error
	tmpstr := strings.Split(hostinfo, "|")
	myhost.IP = tmpstr[0]
	myhost.SSH_PORT, err = strconv.Atoi(tmpstr[1])
	if err != nil {
		panic("Invalid SSH Port Number!")
	}
	myhost.DB_PORT, err = strconv.Atoi(tmpstr[2])
	if err != nil {
		panic("Invalid DB Port Number!")
	}
	myhost.DM_HOME = tmpstr[3]
	myhost.OS = tmpstr[4]
	myhost.CPU = tmpstr[5]
	myhost.SSH_USR = tmpstr[6]
	myhost.SSH_PWD = url.PathEscape(tmpstr[7])
	myhost.DB_USR = tmpstr[8]
	myhost.DB_PWD = url.PathEscape(tmpstr[9])
	myhost.HCFILE = hctool.DmHC_Sel(myhost.OS, myhost.CPU)
	myhost.CFILE = fmt.Sprintf("conf_%s_%s.ini", myhost.IP, myhost.OS)
	myhost.FLST = &[]string{}
	myhost.RemoteDIR = tmpstr[10]
	myhost.SimpleNo, err = strconv.Atoi(tmpstr[11])
	if err != nil {
		panic("Invalid Simple Number!")
	}
}

func SshConnect(myhost HostInfo) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
		//hostKey      ssh.PublicKey
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(myhost.SSH_PWD))
	clientConfig = &ssh.ClientConfig{
		User:            myhost.SSH_USR,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//ssh.FixedHostKey(hostKey),
	}
	// connet to sshtool
	addr = fmt.Sprintf("%s:%d", myhost.IP, myhost.SSH_PORT)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return sshClient, nil
}

func DmHC_Chk(myhost HostInfo) {
	for _, f := range *myhost.FLST {
		if !ostool.Exists(f) {
			panic(fmt.Sprintf("%s does not exist!\n", f))
		}
	}
}

func SftpConnect(client *ssh.Client) (*sftp.Client, error) {
	var (
		sftpClient *sftp.Client
		err        error
	)
	// create sftptool client
	if sftpClient, err = sftp.NewClient(client); err != nil {
		return nil, err
	}
	return sftpClient, nil
}

func MkRemoteDir(client *sftp.Client, myhost HostInfo) {
	var (
		err error
	)
	err = client.MkdirAll(myhost.RemoteDIR)
	if err != nil {
		panic("Remote Directory Create Failed! " + err.Error())
	}
	err = client.Chmod(myhost.RemoteDIR, os.FileMode(0755))
	if err != nil {
		panic("Remote Directory Permission Change Failed! " + err.Error())
	}
}

func Upload(client *sftp.Client, myhost HostInfo, detail int64) {
	//defer client.Close()

	flst := []string{myhost.HCFILE, myhost.CFILE}

	for _, f := range flst {
		if detail > 0 {
			fmt.Printf(">>>>>> Sending File %s <<<<<<<\n", f)
		}
		remoteFile := hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f)
		tarFile, err := client.Create(remoteFile)
		if err != nil {
			panic("Remote File Create Failed! " + err.Error())
		}
		err = client.Chmod(remoteFile, os.FileMode(0755))
		if err != nil {
			panic("Remote File Permission Change Failed! " + err.Error())
		}
		defer tarFile.Close()
		srcFile, err := os.Open(f)
		if err != nil {
			panic("Local File Open Failed! " + err.Error())
		}
		defer srcFile.Close()
		buf := make([]byte, 2048)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			tarFile.Write(buf[:n])
		}
	}
}

func RunCmd(client *ssh.Client, cmd string, detail int64) {
	var (
		err     error
		session *ssh.Session
	)

	session, err = client.NewSession()
	if err != nil {
		panic("SSH Session Create Failed! " + err.Error())
	}

	defer session.Close()
	session.Stdin = bytes.NewBufferString("")
	session.Stdout = &bytes.Buffer{}
	session.Stderr = &bytes.Buffer{}

	if detail > 0 {
		fmt.Println(cmd)
	}
	if err := session.Run(cmd); err != nil {
		panic(
			err.Error() + "\n" + session.Stdout.(*bytes.Buffer).String() + "\n" +
				session.Stderr.(*bytes.Buffer).String(),
		)
	}
	result := session.Stdout.(*bytes.Buffer).String()
	if detail > 0 {
		fmt.Println(result)
	}
}

func Download(client *sftp.Client, localDir string, myhost HostInfo, detail int64) {
	//defer client.Close()
	tmplst, err := client.ReadDir(myhost.RemoteDIR)
	if err != nil {
		panic("Read Remote Directory Failed " + err.Error())
	}
	var downlst []string
	filter := regexp.MustCompile(`.docx|.xlsx|.log`)
	for _, v := range tmplst {
		if v.IsDir() {
			continue
		} else {
			fname := filter.FindString(v.Name())
			if fname != "" {
				downlst = append(downlst, v.Name())
				*myhost.FLST = append(*myhost.FLST, v.Name())
			}
		}
	}
	err = os.MkdirAll(localDir, os.FileMode(0755))
	if err != nil {
		panic("Local Directory Create Failed! " + err.Error())
	}

	for _, f := range downlst {
		if detail > 0 {
			fmt.Printf(">>>>>> Receiving File %s <<<<<<<\n", f)
		}
		srcFile, err := client.OpenFile(hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f), os.O_RDONLY)
		if err != nil {
			panic("Remote File Open Failed! " + err.Error())
		}
		defer srcFile.Close()

		tarFile, err := os.Create(hctool.SmartPathJoin(myhost.OS, localDir, f))
		if err != nil {
			panic("Local File Create Failed! " + err.Error())
		}
		defer tarFile.Close()

		buf := make([]byte, 2048)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			tarFile.Write(buf[:n])
		}
	}
}

func RunHC(client *ssh.Client, myhost HostInfo, detail int64) {
	var cmd string
	if myhost.OS == "linux" {
		cmd = "cd " + myhost.RemoteDIR + " && ./" + myhost.HCFILE + " " + myhost.CFILE
	} else if myhost.OS == "windows" {
		cmd = "cd /d " + myhost.RemoteDIR + " && " + myhost.HCFILE + " " + myhost.CFILE
	}
	if detail > 0 {
		fmt.Printf(">>>>>> Collecting Info <<<<<<<\n")
	}
	RunCmd(client, cmd, detail)
}

func RemoveHC(client *sftp.Client, myhost HostInfo, detail int64) {
	for _, f := range *myhost.FLST {
		fname := hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f)
		if detail > 0 {
			fmt.Printf(">>>>>> Removing File %s <<<<<<<\n", fname)
		}
		err := client.Remove(fname)
		if err != nil {
			panic("Removing File Failed:" + err.Error())
		}

	}
	if detail > 0 {
		fmt.Printf(">>>>>> Removing Directory %s <<<<<<<\n", myhost.RemoteDIR)
	}
	err := client.RemoveDirectory(myhost.RemoteDIR)
	if err != nil {
		panic("Removing Remote Directory Failed:" + err.Error())
	}
}

func ChkDirEmpty(client *sftp.Client, myhost HostInfo) {
	tmplst, err := client.ReadDir(myhost.RemoteDIR)
	if err != nil {
		panic("Read Remote Directory Failed " + err.Error())
	}
	if len(tmplst) != 0 {
		panic("Remote Directory Not Empty!")
	}
}

func ChkRemotePath(myhost HostInfo) bool {
	if SmartIsAbs(myhost.OS, myhost.RemoteDIR) {
		return true
	}
	return false
}

func SmartIsAbs(osname, path string) bool {
	if osname == "linux" {
		return len(path) > 0 && path[0] == '/'
	} else if osname == "windows" {
		return len(path) > 0 && path[1] == ':' && path[2] == '\\'
	}
	return false
}

func ConfGen(myhost HostInfo, mylog *logrus.Logger) {
	srccfg := cfgparser.Cfile{Path: fmt.Sprintf("conf_%s.ini", myhost.OS)}
	srccfg.Initialize(mylog)
	srccfg.SetStrVal("database", "dmhome", myhost.DM_HOME)
	srccfg.SetStrVal("database", "svrname", fmt.Sprintf("127.0.0.1:%d", myhost.DB_PORT))
	srccfg.SetStrVal("database", "username", myhost.DB_USR)
	srccfg.SetStrVal("database", "password", myhost.DB_PWD)
	srccfg.SaveFile(myhost.CFILE)
}
