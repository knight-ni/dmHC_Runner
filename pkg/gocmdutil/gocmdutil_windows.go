package gocmdutil

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"syscall"
)

func RunCMD(command string) ([]byte, error) {
	var opBytes []byte
	var err error
	var stdout io.ReadCloser
	cmd := exec.Command("cmd", "/c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.New(err.Error())
	}

	if opBytes, err = ioutil.ReadAll(stdout); err != nil {
		return nil, errors.New(err.Error())
	}
	err = stdout.Close()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return opBytes, nil
}
