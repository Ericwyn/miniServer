package linux

import (
	"fmt"
	"github.com/Ericwyn/MiniServer/service"
	"github.com/Ericwyn/MiniServer/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var LinuxMiniServerCmd = service.MiniServerCmd{
	SystemName:  "windows",
	KillPortFun: killPort,
	KillAllFun:  killAll,
	ListFun:     listService,
	Run:         linuxRun,
}

func linuxRun(dirPath string, port string, ipAddrArr []string) {
	//absDirPath := *dir
	//if absDirPath != "null" && !filepath.IsAbs(absDirPath) {
	//	absDirPath, _ = filepath.Abs(absDirPath)
	//}
	//
	if dirPath == "null" {
		fmt.Println("未使用 -d 指定监听路径")
		dirPath = utils.GetCurPath()
		utils.RunInNewProcess(port, dirPath, ipAddrArr)
	} else if !filepath.IsAbs(dirPath) {
		dirPath = filepath.Clean(dirPath)
		dirPath, _ = filepath.Abs(dirPath)
		utils.RunInNewProcess(port, dirPath, ipAddrArr)
	} else {
		dirPath = strings.Replace(dirPath, "\\", "/", -1)
		//utils.RunInNewProcess(port, dirPath, ipAddrArr)

		fmt.Println("监听:" + dirPath)
		h := http.FileServer(http.Dir(dirPath))
		ports := ":" + port
		fmt.Println("服务启动在" + ports)
		for _, ip := range ipAddrArr {
			fmt.Println("http://" + ip + ports)
		}
		err2 := http.ListenAndServe(ports, h)
		if err2 != nil {
			log.Fatal("ListenAndServe: ", err2)
		}
	}
}

func killPort(port int) bool {
	for _, th := range listService() {
		if th.Port == strconv.Itoa(port) {
			return killProcess(th)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false

}

func killAll() {
	for _, th := range listService() {
		killProcess(th)
	}

}

func listService() []service.Process {
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
	var resList []service.Process
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
			process := service.Process{}

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

func killProcess(process service.Process) bool {

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
