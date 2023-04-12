package sftptool

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"rtRunner/pkg/cfgparser"
	"rtRunner/pkg/hctool"
	"rtRunner/pkg/ostool"
	"time"
)

type HostInfo struct {
	Customer  string
	Appname   string
	LocalDir  string
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

func HostInit(host string, localdir string, mycfg cfgparser.Cfile, myhost *HostInfo) {
	myhost.LocalDir = localdir
	myhost.Customer = mycfg.GetStrVal(host, "customer")
	myhost.Appname = mycfg.GetStrVal(host, "appname")
	myhost.IP = mycfg.GetStrVal(host, "address")
	myhost.SSH_PORT = mycfg.GetIntVal(host, "ssh_port")
	myhost.DB_PORT = mycfg.GetIntVal(host, "db_port")
	myhost.DM_HOME = mycfg.GetStrVal(host, "dm_home")
	myhost.OS = mycfg.GetStrVal(host, "os")
	myhost.CPU = mycfg.GetStrVal(host, "cpu")
	myhost.SSH_USR = mycfg.GetStrVal(host, "ssh_usr")
	myhost.SSH_PWD = url.PathEscape(mycfg.GetStrVal(host, "ssh_pwd"))
	myhost.DB_USR = mycfg.GetStrVal(host, "db_usr")
	myhost.DB_PWD = url.PathEscape(mycfg.GetStrVal(host, "db_pwd"))
	myhost.HCFILE = hctool.DmHC_Sel(myhost.OS, myhost.CPU)
	myhost.RemoteDIR = mycfg.GetStrVal(host, "remotedir")
	myhost.SimpleNo = mycfg.GetIntVal(host, "simple")
	myhost.CFILE = fmt.Sprintf("conf_%s_%s_%d.ini", myhost.IP, myhost.OS, myhost.SimpleNo)
	myhost.FLST = &[]string{myhost.HCFILE, myhost.CFILE}
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
		if !ostool.Exists(filepath.Join(myhost.LocalDir, f)) {
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

func Upload(client *sftp.Client, myhost HostInfo, detail int) {
	//defer client.Close()

	flst := []string{myhost.HCFILE, myhost.CFILE}

	for _, f := range flst {
		if detail > 0 {
			fmt.Printf(">>>>>> Sending File %s <<<<<<<\n", f)
		}
		remoteFile := hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f)
		tarFile, err := client.Create(remoteFile)
		if err != nil {
			panic("Remote File " + remoteFile + " Create Failed! " + err.Error())
		}
		err = client.Chmod(remoteFile, os.FileMode(0755))
		if err != nil {
			panic("Remote File " + remoteFile + " Permission Change Failed! " + err.Error())
		}
		defer tarFile.Close()
		srcFile, err := os.Open(filepath.Join(myhost.LocalDir, f))
		if err != nil {
			panic("Local File " + srcFile.Name() + " Open Failed! " + err.Error())
		}
		defer srcFile.Close()
		buf := make([]byte, 2048)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			_, err := tarFile.Write(buf[:n])
			if err != nil {
				panic("Upload Write " + tarFile.Name() + " Failed! " + err.Error())
			}
		}
	}
}

func RunCmd(client *ssh.Client, cmd string, detail int) {
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

func Download(client *sftp.Client, localDir string, myhost HostInfo, detail int) {
	//defer client.Close()
	tmplst, err := client.ReadDir(myhost.RemoteDIR)
	if err != nil {
		panic("Read Remote Directory " + myhost.RemoteDIR + " Failed! " + err.Error())
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
		panic("Local Directory " + localDir + " Create Failed! " + err.Error())
	}

	for _, f := range downlst {
		if detail > 0 {
			fmt.Printf(">>>>>> Receiving File %s <<<<<<<\n", f)
		}
		srcFile, err := client.OpenFile(hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f), os.O_RDONLY)
		if err != nil {
			panic("Remote File " + srcFile.Name() + " Open Failed! " + err.Error())
		}
		defer srcFile.Close()

		tarFile, err := os.Create(hctool.SmartPathJoin(myhost.OS, localDir, f))
		if err != nil {
			panic("Local File " + tarFile.Name() + " Create Failed! " + err.Error())
		}
		defer tarFile.Close()

		buf := make([]byte, 2048)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			_, err := tarFile.Write(buf[:n])
			if err != nil {
				panic("Download Write " + tarFile.Name() + " Failed! " + err.Error())
			}
		}
	}
}

func RunHC(client *ssh.Client, myhost HostInfo, detail int) {
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

func RemoveHC(client *sftp.Client, myhost HostInfo, detail int) {
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

func ConfGen(myhost HostInfo, dir string) {
	srccfg := cfgparser.Cfile{Path: filepath.Join(dir, fmt.Sprintf("conf_%s.ini", myhost.OS))}
	srccfg.Initialize()
	srccfg.SetStrVal("baseinfo", "customer", myhost.Customer)
	srccfg.SetStrVal("baseinfo", "appname", myhost.Appname)
	srccfg.SetStrVal("database", "dmhome", myhost.DM_HOME)
	srccfg.SetStrVal("database", "svrname", fmt.Sprintf("127.0.0.1:%d", myhost.DB_PORT))
	srccfg.SetStrVal("database", "username", myhost.DB_USR)
	srccfg.SetStrVal("database", "password", myhost.DB_PWD)
	if myhost.SimpleNo == 1 {
		srccfg.SetStrVal("report", "dbinfo_top", "0")
		srccfg.SetStrVal("report", "dbinfo_log", "0")
		srccfg.SetStrVal("report", "secinfo", "0")
	}
	srccfg.SaveFile(filepath.Join(dir, myhost.CFILE))
}
