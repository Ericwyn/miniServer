package service

import (
	"fmt"
	"github.com/Ericwyn/MiniServer/utils"
	"github.com/shirou/gopsutil/v3/process"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// 跨平台公有实现

func KillPort(pid string) bool {
	processes, err := process.Processes()
	if err != nil {
		log.Fatalln("无法获取 process 列表")
		return false
	}
	for _, proc := range processes {
		if fmt.Sprint(proc) == pid {
			err := proc.Kill()
			if err != nil {
				log.Fatalln("kill pid : " + pid + " error")
				return false
			}
			return true
		}
	}
	return false
}

func ListServer() []Process {
	var resultList []Process

	processes, err := process.Processes()
	if err != nil {
		return nil
	}
	for _, proc := range processes {
		var err error
		name, err := proc.Name()
		if err != nil {
			continue
		}
		if !strings.Contains(name, "miniServer") {
			continue
		}

		cmdline, err := proc.CmdlineSlice()
		if err != nil {
			continue
		}
		cmdLineStr := fmt.Sprint(cmdline)
		// 正式执行的命令行最末尾会有个 trueRun 参数
		if !strings.Contains(cmdLineStr, "trueRun") {
			continue
		}
		resProc := Process{
			Process: proc,
		}

		for i := 0; i < len(cmdline); i++ {
			if cmdline[i] == "-d" {
				resProc.DirPath = cmdline[i+1]
				i = i + 1
			}
			if cmdline[i] == "-p" {
				resProc.Pid = fmt.Sprint(proc.Pid)
				resProc.Port = cmdline[i+1]
				i = i + 1
			}
		}
		resultList = append(resultList, resProc)
	}
	return resultList
}

func Run(dirPath string, port string, ipAddrArr []string) {
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
