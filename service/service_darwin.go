//go:build darwin

package service

import (
	"fmt"
	"log"
	"strconv"
)

var Cmd = MiniServerCmd{
	SystemName:  "darwin",
	KillPortFun: killPort,
	KillAllFun:  killAll,
	ListFun:     ListServer,
	Run:         Run,
}

func killPort(port int) bool {
	for _, server := range ListServer() {
		if server.Port != strconv.Itoa(port) {
			continue
		}
		if server.Process != nil {
			return server.Process.Kill() == nil
		}
		return KillPort(server.Pid)
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false
}

func killAll() {
	for _, server := range ListServer() {
		if server.Process == nil {
			continue
		}
		if err := server.Process.Kill(); err != nil {
			log.Println("关闭端口为 " + server.Port + " 的进程失败: " + err.Error())
			continue
		}
		fmt.Println("成功关闭端口为", server.Port, "的进程")
	}
}
