package sftptool

import (
	"bytes"
	"dmHC_Runner/pkg/cfgparser"
	"dmHC_Runner/pkg/hctool"
	"dmHC_Runner/pkg/ostool"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/wxnacy/wgo/file"
	"golang.org/x/crypto/ssh"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type HostInfo struct {
	Customer     string
	Appname      string
	LocalDir     string
	IP           string
	SSH_PORT     int
	DB_PORT      int
	DM_HOME      string
	OS           string
	CPU          string
	SSH_USR      string
	SSH_PWD      string
	DB_USR       string
	DB_PWD       string
	HCFILE       string
	CFILE        string
	FLST         *[]string
	RemoteDIR    string
	RemoteDIROut string
	SimpleNo     int
	RsaFile      string
}

func in(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

func HostInit(host string, localdir string, mycfg cfgparser.Cfile, myhost *HostInfo) {
	myhost.LocalDir = localdir
	myhost.Customer = mycfg.GetStrVal(host, "customer")
	myhost.Appname = mycfg.GetStrVal(host, "appname")
	myhost.IP = mycfg.GetStrVal(host, "address")
	ip := net.ParseIP(myhost.IP)
	if ip == nil {
		panic("Invalid IP Address!")
	}
	myhost.SSH_PORT = mycfg.GetIntVal(host, "ssh_port")
	myhost.DB_PORT = mycfg.GetIntVal(host, "db_port")
	myhost.OS = mycfg.GetStrVal(host, "os")
	if !in(myhost.OS, hctool.Support_OS) {
		panic(fmt.Sprintf("OS Can Only Be In %v", hctool.Support_OS))
	}
	myhost.CPU = mycfg.GetStrVal(host, "cpu")
	if !in(myhost.CPU, hctool.Support_CPU) {
		panic(fmt.Sprintf("CPU Can Only Be In %v", hctool.Support_CPU))
	}
	myhost.SSH_USR = mycfg.GetStrVal(host, "ssh_usr")
	myhost.SSH_PWD = mycfg.GetStrVal(host, "ssh_pwd")
	myhost.DB_USR = mycfg.GetStrVal(host, "db_usr")
	myhost.DB_PWD = url.PathEscape(mycfg.GetStrVal(host, "db_pwd"))
	myhost.HCFILE = hctool.DmHC_Sel(myhost.OS, myhost.CPU)
	myhost.RemoteDIR = mycfg.GetStrVal(host, "remotedir")
	if !SmartIsAbs(myhost.OS, myhost.RemoteDIR) {
		panic("Remote Work Directory Must Be Absolute Path!")
	}
	myhost.DM_HOME = mycfg.GetStrVal(host, "dm_home")
	if !SmartIsAbs(myhost.OS, myhost.DM_HOME) {
		panic("DM HOME Must Be Absolute Path!")
	}
	myhost.SimpleNo = mycfg.GetIntVal(host, "simple")
	if myhost.SimpleNo != 0 && myhost.SimpleNo != 1 {
		panic("SimpleNo Can Only Be 0 or 1!")
	}
	myhost.CFILE = fmt.Sprintf("conf_%s_%s_%d.ini", myhost.IP, myhost.OS, myhost.SimpleNo)
	myhost.FLST = &[]string{}
	myhost.RsaFile = mycfg.GetStrVal(host, "rsa_file")
}

func IsLocalIP(ip string) bool {
	tmpip := net.ParseIP(ip)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil && ipnet.IP.Equal(tmpip) {
				return true
			}
		}
	}
	return false
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

	if file.Exists(myhost.RsaFile) {
		key, err := os.ReadFile(myhost.RsaFile)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		auth = append(auth, ssh.Password(myhost.SSH_PWD))
	}
	clientConfig = &ssh.ClientConfig{
		User:    myhost.SSH_USR,
		Auth:    auth,
		Timeout: 30 * time.Second,
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//ssh.FixedHostKey(hostKey),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
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

func ChkRemoteDirExists(client *sftp.Client, myhost HostInfo) bool {
	var (
		fi  os.FileInfo
		err error
	)

	fi, err = client.Stat(myhost.RemoteDIROut)
	if err != nil {
		return false
	}

	if err != nil && os.IsNotExist(err) {
		panic("Remote Directory Check Failed! " + err.Error())
	} else if err == nil && fi.IsDir() {
		return true
	} else if err == nil && !fi.IsDir() {
		panic("Remote Directory Name Already Been Used By Other File!")
	}
	return false
}

func ChkLocalHostDirExists(myhost HostInfo) bool {
	var (
		fi  os.FileInfo
		err error
	)
	fi, err = os.Stat(myhost.RemoteDIR)
	if err != nil {
		return false
	}

	fi, err = os.Stat(myhost.RemoteDIROut)
	if err != nil {
		return false
	} else if err == nil && fi.IsDir() {
		return true
	} else if err == nil && !fi.IsDir() {
		panic("Remote Directory Name Already Been Used By Other File!")
	}
	return false
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

func MkLocalHostDir(myhost HostInfo) {
	var (
		err error
	)
	err = os.MkdirAll(myhost.RemoteDIROut, 0770)
	if err != nil {
		panic("Remote Directory Create Failed! " + err.Error())
	}
	err = os.Chmod(myhost.RemoteDIROut, os.FileMode(0755))
	if err != nil {
		panic("Remote Directory Permission Change Failed! " + err.Error())
	}
}

func Upload(client *sftp.Client, myhost HostInfo, detail int) {
	//defer client.Close()

	flst := []string{myhost.HCFILE, myhost.CFILE, "dmHC.lic"}

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
		srcFilePath := filepath.Join(myhost.LocalDir, f)
		if !file.Exists(srcFilePath) {
			panic("Local File " + srcFilePath + " Does Not Exists! ")
		}
		srcFile, err := os.Open(srcFilePath)
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
		*myhost.FLST = append(*myhost.FLST, f)
	}
}

func LocalHostUpload(myhost HostInfo, detail int) {
	//defer client.Close()

	flst := []string{myhost.HCFILE, myhost.CFILE, "dmHC.lic"}

	for _, f := range flst {
		if detail > 0 {
			fmt.Printf(">>>>>> Sending File %s <<<<<<<\n", f)
		}
		remoteFile := hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f)
		tarFile, err := os.Create(remoteFile)
		if err != nil {
			panic("Remote File " + remoteFile + " Create Failed! " + err.Error())
		}
		err = os.Chmod(remoteFile, os.FileMode(0755))
		if err != nil {
			panic("Remote File " + remoteFile + " Permission Change Failed! " + err.Error())
		}
		defer tarFile.Close()
		srcFilePath := filepath.Join(myhost.LocalDir, f)
		if !file.Exists(srcFilePath) {
			panic("Local File " + srcFilePath + " Does Not Exists! ")
		}
		srcFile, err := os.Open(srcFilePath)
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
		*myhost.FLST = append(*myhost.FLST, f)
	}
}

func RunLocalHostCmd(cmd string, detail int) {
	var (
		err error
	)

	cmdlst := strings.Split(cmd, "&&")
	dirarg := strings.Join(strings.Split(cmdlst[0], " ")[2:], "")

	err = os.Chdir(dirarg)
	if err != nil {
		panic("Change Directory Failed! " + err.Error())
	}

	cmdstr := strings.Split(cmdlst[1], " ")
	command := exec.Command(".\\"+cmdstr[1], cmdstr[2:]...)

	if err != nil {
		panic("Command Create Failed! " + err.Error())
	}

	command.Stdin = bytes.NewBufferString("")
	command.Stdout = &bytes.Buffer{}
	command.Stderr = &bytes.Buffer{}

	if detail > 0 {
		fmt.Println("Command:%s\n", cmd)
	}
	if err := command.Run(); err != nil {
		panic(
			err.Error() + "\n" + command.Stdout.(*bytes.Buffer).String() + "\n" +
				command.Stderr.(*bytes.Buffer).String(),
		)
	}
	result := command.Stdout.(*bytes.Buffer).String()
	if detail > 0 {
		fmt.Println(result)
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

func LocalHostDownload(localDir string, myhost HostInfo, detail int) {
	//defer client.Close()
	tmplst, err := os.ReadDir(myhost.RemoteDIR)
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
		srcFile, err := os.OpenFile(hctool.SmartPathJoin(myhost.OS, myhost.RemoteDIR, f), os.O_RDONLY, 0644)
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

func RunLocalHostHC(myhost HostInfo, detail int) {
	var cmd string
	if myhost.OS == "linux" {
		cmd = "cd " + myhost.RemoteDIR + " && ./" + myhost.HCFILE + " " + myhost.CFILE
	} else if myhost.OS == "windows" {
		cmd = "cd /d " + myhost.RemoteDIR + " && " + myhost.HCFILE + " " + myhost.CFILE
	}
	if detail > 0 {
		fmt.Printf(">>>>>> Collecting Info <<<<<<<\n")
	}
	RunLocalHostCmd(cmd, detail)
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
		fmt.Printf(">>>>>> Removing Directory %s <<<<<<<\n", myhost.RemoteDIROut)
	}
	err := client.RemoveDirectory(myhost.RemoteDIR)
	if err != nil {
		panic("Removing Remote Directory Failed:" + err.Error())
	}
}

func ChkDirEmpty(client *sftp.Client, myhost HostInfo) bool {
	tmplst, err := client.ReadDir(myhost.RemoteDIR)
	if err != nil {
		panic("Read Remote Directory Failed " + err.Error())
	}
	if len(tmplst) != 0 {
		return false
	}
	return true
}

func ChkLocalHostDirEmpty(myhost HostInfo) bool {
	tmplst, err := os.ReadDir(myhost.RemoteDIR)
	if err != nil {
		panic("Read Remote Directory Failed " + err.Error())
	}
	if len(tmplst) != 0 {
		return false
	}
	return true
}

func ChkRemotePath(myhost *HostInfo) bool {
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
