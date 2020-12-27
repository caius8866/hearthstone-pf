package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"time"
)

var logger = logrus.New()

func main() {
	var input CmdInput
	var flagSet = flag.NewFlagSet("hearthstone", flag.ExitOnError)

	flagSet.BoolVar(&input.Enable, "e", false, "炉石传说网络恢复")
	flagSet.BoolVar(&input.Disable, "d", false, "炉石传说网络中断")
	flagSet.BoolVar(&input.Debug, "debug", false, "调试模式")
	flagSet.UintVar(&input.IntervalSeconds, "s", 0, "自动重连间隔(单位秒)")
	flagSet.Usage = usage

	flagSet.Parse(os.Args[1:])

	if (input.Enable && input.Disable) || (!input.Enable && !input.Disable) {
		flagSet.Usage()
		return
	}

	logger.Out = os.Stdout

	handler := NewCmdHandler().
		Register(CmdHandlerTypeDisable, disableHandler).
		Register(CmdHandlerTypeEnable, enableHandler)

	if input.Debug {
		logger.SetLevel(logrus.DebugLevel)
		handler.RegisterPreRunHandler(debugHandler).RegisterPostRunHandler(debugHandler)
	}

	var t CmdHandlerType
	if input.Enable {
		t = CmdHandlerTypeEnable
	} else if input.Disable {
		t = CmdHandlerTypeDisable
	}

	err := handler.Handle(t, input)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}
}

func debugHandler(c CmdInput) error {
	b, err := DefaultHearthstonePF().PFRules()
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return err
	}

	logger.Debugf("%s\n", b)
	return nil
}

func disableHandler(c CmdInput) error {
	err := DefaultHearthstonePF().Disable()
	if err != nil && err != ErrDupRun {
		fmt.Printf("err: %s\n", err.Error())
		return err
	}
	fmt.Println("炉石网络恢复")
	return nil
}

func enableHandler(c CmdInput) error {
	if !hearthstoneIsRunning() {
		return errors.New("炉石客户端未启动")
	}

	pf := DefaultHearthstonePF()

	err := pf.Enable()
	if err != nil && err != ErrDupRun {
		return err
	}

	fmt.Println("炉石网络中断成功")

	if c.IntervalSeconds == 0 {
		return nil
	}

	fmt.Println("正在等待网络恢复")

	time.Sleep(time.Duration(c.IntervalSeconds) * time.Second)

	err = pf.Disable()
	if err != nil && err != ErrDupRun {
		return err
	}

	fmt.Println("炉石网络恢复")
	return nil
}

func usage() {
	usage := []string{
		"参数说明:",
		"-d 	炉石传说断网",
		"-e 	炉石传说网络恢复",
		"-s 	自动重连间隔(单位秒)",
		"-debug  调试模式\n",
		"使用示例(以文件放在下载目录为例):",
		"自动档(断网7秒后恢复网络): sudo ~/Downloads/hearthstone-pf -e -s 7",
		"手动档(断网): 		   sudo ~/Downloads/hearthstone-pf -e",
		"手动档(网络恢复): 	   sudo ~/Downloads/hearthstone-pf -d",
	}
	fmt.Printf("%s\n", strings.Join(usage, "\n"))
}

func hearthstoneIsRunning() bool {
	cmd := exec.Command("/bin/bash", "-c", "lsof -i:1226")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	bs := bytes.Split(out, []byte("\n"))
	if len(bs) <= 1 {
		return false
	}

	return true
}

type CmdHandlerType int

const (
	CmdHandlerTypeEnable CmdHandlerType = iota + 1
	CmdHandlerTypeDisable
)

type CmdHandlerFn func(c CmdInput) error

type CmdInput struct {
	Disable         bool
	Enable          bool
	Debug           bool
	IntervalSeconds uint
}

type CmdHandler struct {
	handlerMap            map[CmdHandlerType]CmdHandlerFn
	preRunMiddlewareList  []CmdHandlerFn
	postRunMiddlewareList []CmdHandlerFn
}

func NewCmdHandler() *CmdHandler {
	return &CmdHandler{
		handlerMap: make(map[CmdHandlerType]CmdHandlerFn),
	}
}

func (h *CmdHandler) RegisterPreRunHandler(f CmdHandlerFn) *CmdHandler {
	h.preRunMiddlewareList = append(h.preRunMiddlewareList, f)
	return h
}

func (h *CmdHandler) RegisterPostRunHandler(f CmdHandlerFn) *CmdHandler {
	h.postRunMiddlewareList = append(h.postRunMiddlewareList, f)
	return h
}

func (h *CmdHandler) Register(t CmdHandlerType, f CmdHandlerFn) *CmdHandler {
	h.handlerMap[t] = f
	return h
}

func (h *CmdHandler) Handle(t CmdHandlerType, c CmdInput) error {
	fn, ok := h.handlerMap[t]
	if !ok {
		return errors.New("handler not found")
	}

	if len(h.preRunMiddlewareList) > 0 {
		for _, preRunFn := range h.preRunMiddlewareList {
			err := preRunFn(c)
			if err != nil {
				return err
			}
		}
	}

	err := fn(c)
	if err != nil {
		return err
	}

	if len(h.postRunMiddlewareList) > 0 {
		for _, postRunFn := range h.postRunMiddlewareList {
			err := postRunFn(c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
