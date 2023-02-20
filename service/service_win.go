//go:build windows

package service

import (
	"fmt"
	"log"
	"strconv"
)

var Cmd = MiniServerCmd{
	SystemName:  "windows",
	KillPortFun: killPort,
	KillAllFun:  killAll,
	ListFun:     ListServer,
	Run:         Run,
}

func killPort(port int) bool {
	for _, th := range ListServer() {
		if th.Port == strconv.Itoa(port) {
			if th.Process != nil {
				err := th.Process.Kill()
				if err != nil {
					return false
				}
				return true
			}
			return KillPort(th.Pid)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false
}

func killAll() {
	for _, th := range ListServer() {
		if th.Process != nil {
			err := th.Process.Kill()
			if err != nil {
				log.Fatalln("关闭端口为 : " + th.Port + " 的进程失败")
			} else {
				fmt.Println("成功关闭端口为 : " + th.Port + " 的进程")
			}
		}
	}
}
