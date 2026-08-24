package service

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Ericwyn/MiniServer/conf"
	"github.com/Ericwyn/MiniServer/utils"
	"github.com/shirou/gopsutil/v3/process"
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
		conf.RunDirPath = dirPath
		startFileServer(port, dirPath, ipAddrArr)
	}
}

func startFileServer(port string, dirPath string, ipAddrArr []string) {
	fmt.Println("监听:" + dirPath)
	fmt.Println("-----------------------------------")
	fmt.Println("服务将启动在以下地址")
	for _, ip := range ipAddrArr {
		fmt.Println("http://" + ip + ":" + port)
	}
	fmt.Println("-----------------------------------")

	handler, err := newFileServerHandler(dirPath)
	if err != nil {
		log.Fatal("无法打开监听目录: ", err)
	}

	server := newHTTPServer(":"+port, handler)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("文件服务器启动失败: ", err)
	}
}
