package gocmdutil

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
)

func RunCMD(command string) ([]byte, error) {
	var opBytes []byte
	var err error
	var stdout io.ReadCloser
	cmd := exec.Command("bash", "-c", command)
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
