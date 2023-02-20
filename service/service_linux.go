//go:build linux

package service

import (
	"fmt"
	"github.com/Ericwyn/MiniServer/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var Cmd = MiniServerCmd{
	SystemName:  "windows",
	KillPortFun: killPort,
	KillAllFun:  killAll,
	ListFun:     listService,
	Run:         Run,
}

func killPort(port int) bool {
	for _, th := range listService() {
		if th.Port == strconv.Itoa(port) {
			return killProcess(th)
			// linux 不使用 cross 的方法，因为执行起来会很慢
			//return cross.KillPort(th.Pid)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false

}

func killAll() {
	for _, th := range listService() {
		killProcess(th)
		//cross.KillPort(th.Pid)
	}

}

func listService() []Process {
	var cmdOutput []byte
	var err error
	var cmd *exec.Cmd

	// 执行单个shell命令时, 直接运行即可
	cmd = exec.Command("netstat", "-nlp")
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println("netstate 命令调用错误")
		os.Exit(1)
	}
	lines := strings.Split(string(cmdOutput), "\n")

	//portRegex := regexp.MustCompile(`:::[0-9]{1,5}`)
	//pidRegex := regexp.MustCompile(`[0-9]{1,5}\/miniServer`)
	var resList []Process
	for _, line := range lines {
		if strings.Contains(line, "miniServer") {
			//pid := strings.Replace(pidRegex.FindString(line), "/miniServer", "", -1)
			//resList = append(resList, service.Process{
			//	Port:    strings.Replace(portRegex.FindString(line), ":", "", -1),
			//	Pid:     pid,
			//	DirPath: utils.GetListenDir(pid),
			//})

			for i := 0; i < 5; i++ {
				line = strings.ReplaceAll(line, "  ", " ")
			}
			process := Process{}

			for _, s := range strings.Split(line, " ") {
				if strings.Contains(s, ":::") && !strings.Contains(s, ":*") {
					process.Port = strings.ReplaceAll(s, ":", "")
				}
				if strings.Contains(s, "/miniServer") {
					process.Pid = strings.ReplaceAll(s, "/miniServer", "")
					process.DirPath = utils.GetListenDir(process.Pid)
				}
			}

			if process.Pid != "" && process.Port != "" {
				resList = append(resList, process)
			}
		}
	}
	return resList
}

func killProcess(process Process) bool {

	//// second arg is the signal number
	//arg1 := os.Args[1]
	//val1, _ := strconv.ParseInt(arg1, 10, 32)
	//signal := int(val1)

	pid, err := strconv.Atoi(process.Pid)
	if err != nil {
		return false
	}

	err = syscall.Kill(pid, 3)
	//var cmd *exec.Cmd
	//cmd = exec.Command("kill", process.Pid)
	if err != nil {
		fmt.Println("kill 线程失败, 错误 :", err)
		return false
	}
	fmt.Println("kill 端口", process.Port, "上的 miniServer 进程")
	return true
}
