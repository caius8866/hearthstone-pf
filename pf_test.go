package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func TestDisable(t *testing.T) {
	pf := NewHearthstonePF("./pf.conf", []uint16{
		1119,
		3724,
	})

	err := pf.Disable()

	if err != nil {
		t.Errorf("出错了: %s\n", err.Error())
	}
}

func TestEnable(t *testing.T) {
	pf := NewHearthstonePF("./pf.conf", []uint16{
		1119,
		3724,
	})

	err := pf.Enable()

	if err != nil {
		t.Errorf("出错了: %s\n", err.Error())
	}
}

func TestProgressBar(t *testing.T) {
	//for {
	//	for _, r := range `-\|/` {
	//		fmt.Printf("\r%c", r)
	//		time.Sleep(100 * time.Millisecond)
	//	}
	//}

	shell := fmt.Sprintf("lsof -i:1226")

	cmd := exec.Command("/bin/bash", "-c", shell)
	out, err := cmd.Output()

	if err != nil {
		return
	}

	bs := bytes.Split(out, []byte("\n"))

	if len(bs) <= 1 {
		return
	}


	fmt.Printf("game running\n")
}
