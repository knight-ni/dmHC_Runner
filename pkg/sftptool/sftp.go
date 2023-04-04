package sftptool

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"rtRunner/pkg/hctool"
	"time"
)

func SshConnect(user string, password string, host string, port int) (*ssh.Client, error) {
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
	auth = append(auth, ssh.Password(password))
	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//ssh.FixedHostKey(hostKey),
	}
	// connet to sshtool
	addr = fmt.Sprintf("%s:%d", host, port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return sshClient, nil
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

func Upload(client *sftp.Client, remoteDir string, hcfile string, cfile string, detail int64) {
	var (
		err error
	)
	//defer client.Close()
	hctool.DmHC_Chk(hcfile)

	flst := []string{hcfile, cfile}
	err = client.MkdirAll(remoteDir)
	if err != nil {
		panic("Remote Directory Create Failed! " + err.Error())
	}
	err = client.Chmod(remoteDir, os.FileMode(0755))
	if err != nil {
		panic("Remote Directory Permission Change Failed! " + err.Error())
	}

	for _, f := range flst {
		if detail > 0 {
			fmt.Printf(">>>>>> Sending File %s <<<<<<<\n", f)
		}
		remoteFile := path.Join(remoteDir, f)
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

	if err := session.Run(cmd); err != nil {
		fmt.Errorf(
			"%s has failed: [%w] %s",
			cmd,
			err,
			session.Stderr.(*bytes.Buffer).String(),
		)
	}
	result := session.Stdout.(*bytes.Buffer).String()
	if detail > 0 {
		fmt.Println(result)
	}
}

func Download(client *sftp.Client, remoteDir string, localDir string, detail int64) {
	//defer client.Close()
	flst, err := client.ReadDir(remoteDir)
	if err != nil {
		panic("Get Directory Failed " + err.Error())
	}
	var downlst []string
	filter := regexp.MustCompile(`.docx|.xlsx|.log`)
	for _, v := range flst {
		if v.IsDir() {
			continue
		} else {
			fname := filter.FindString(v.Name())
			if fname != "" {
				downlst = append(downlst, v.Name())
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
		srcFile, err := client.OpenFile(path.Join(remoteDir, f), os.O_RDONLY)
		if err != nil {
			panic("Remote File Open Failed! " + err.Error())
		}
		defer srcFile.Close()

		tarFile, err := os.Create(filepath.Join(localDir, f))
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

func RunHC(client *ssh.Client, remoteDir string, hcfile string, cfile string, detail int64) {
	cmd := "cd " + remoteDir + " && ./" + hcfile + " " + cfile
	if detail > 0 {
		fmt.Printf(">>>>>> Collecting Info <<<<<<<\n")
	}
	RunCmd(client, cmd, detail)
}

func RemoveHC(client *sftp.Client, remoteDir string, detail int64) {
	flst, err := client.ReadDir(remoteDir)
	for _, f := range flst {
		if detail > 0 {
			fmt.Printf(">>>>>> Removing File %s <<<<<<<\n", f.Name())
		}
		err := client.Remove(path.Join(remoteDir, f.Name()))
		if err != nil {
			panic("Removing File Error:" + err.Error())
		}
	}
	err = client.RemoveDirectory(remoteDir)
	if err != nil {
		panic("Removing Remote Directory Failed:" + err.Error())
	}
}
