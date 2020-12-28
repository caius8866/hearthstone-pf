package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const BlockSymbol = "hearthstone"

var ErrDupRun = errors.New("重复执行")
var ErrPermissionDenied = errors.New("无权限")

type HearthstonePF struct {
	PFConfPath string
	ports      []uint16
	out        []byte
}

func NewHearthstonePF(confPath string, ports []uint16) *HearthstonePF {
	return &HearthstonePF{
		PFConfPath: confPath,
		ports:      ports,
		out:        make([]byte, 0),
	}
}

func (h *HearthstonePF) Enable() error {
	exist, err := h.blockSymbolInConf()
	if err != nil {
		return err
	}

	if !exist {
		err = h.writeBlockLines()
		if err != nil {
			return err
		}
	}

	err = h.runCmd()
	if err != nil {
		return err
	}

	return nil
}

func (h *HearthstonePF) PFRules() ([]byte, error) {
	cmd := exec.Command("/bin/bash", "-c", "sudo pfctl -s rules")
	return cmd.Output()
}

func (h *HearthstonePF) Disable() error {
	exist, err := h.blockSymbolInConf()
	if err != nil {
		return err
	}

	if exist {
		err = h.delBlockLines()
		if err != nil {
			return err
		}
	}

	err = h.runCmd()
	if err != nil {
		return err
	}

	return nil
}

func (h *HearthstonePF) runCmd() error {
	shell := fmt.Sprintf("sudo pfctl -ef %s", h.PFConfPath)
	dupErrStr := "pf already enabled"

	cmd := exec.Command("/bin/bash", "-c", shell)
	out, err := cmd.CombinedOutput()
	h.out = out

	logger.Debugf("%s", out)

	if err == nil {
		return nil
	}

	//TODO 解析错误,目前只解析常见错误
	var errStr = strings.ToLower(string(out))
	var errMap = map[string]error{
		"permission denied": ErrPermissionDenied,
		dupErrStr:           ErrDupRun,
	}

	for s, e := range errMap {
		if strings.Contains(errStr, s) {
			return e
		}
	}

	return nil
}

func (h *HearthstonePF) blockSymbolInConf() (exist bool, err error) {
	f, err := os.Open(h.PFConfPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err != nil && err != io.EOF {
			return false, err
		}

		if err == io.EOF {
			break
		}

		if strings.Contains(string(line), BlockSymbol) {
			exist = true
			break
		}
	}

	return
}

func (h *HearthstonePF) writeBlockLines() error {
	f, err := os.OpenFile(h.PFConfPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	var portStrList []string
	for _, port := range h.ports {
		portStrList = append(portStrList, fmt.Sprintf("%d", port))
	}

	blockStr := fmt.Sprintf("block out quick proto tcp from any to any port {%s} #%s\n",
		strings.Join(portStrList, ","), BlockSymbol)

	_, err = f.Write([]byte(blockStr))
	if err != nil {
		return err
	}

	return nil
}

func (h *HearthstonePF) delBlockLines() error {
	fb, err := ioutil.ReadFile(h.PFConfPath)
	if err != nil {
		return err
	}

	var lines [][]byte
	var buf = bytes.NewBuffer(fb)

	for {
		line, err := buf.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		if strings.Contains(string(line), BlockSymbol) {
			continue
		}

		lines = append(lines, line)
	}

	b := bytes.Join(lines, []byte{})
	err = ioutil.WriteFile(h.PFConfPath, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (h *HearthstonePF) GetCmdOut() []byte {
	return h.out
}
