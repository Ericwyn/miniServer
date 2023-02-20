package win

import (
	"fmt"
	"github.com/Ericwyn/MiniServer/service"
	"github.com/Ericwyn/MiniServer/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var WinMiniServerCmd = service.MiniServerCmd{
	SystemName:  "windows",
	KillPortFun: killPort,
	KillAllFun:  killAll,
	ListFun:     listService,
	Run:         winRun,
}

func winRun(dirPath string, port string, ipAddrArr []string) {
	if dirPath == "null" {
		fmt.Println("未使用 -d 指定监听路径")
		dirPath = utils.GetCurPath()

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
			return killProcessWin(th)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false
}

func killAll() {
	for _, th := range listService() {
		killProcessWin(th)
	}
}

func listService() []service.Process {
	// get all tasklist
	var process []service.Process

	taskList := cmdWithoutArgs("tasklist")
	//fmt.Println(taskList)
	taskLines := strings.Split(taskList, "\n")

	for _, taskLine := range taskLines {
		// 内存占用会显示多少 k , 没有多少 k 的证明这一行不是 task 信息
		if strings.Index(taskLine, "Console") >= 0 || strings.Index(taskLine, "Services") >= 0 {
			splitPre := replaceAllTwoStrToOne(
				replaceAllTwoStrToOne(taskLine, "  ", "{TEMP}"),
				"{TEMP}{TEMP}",
				"{TEMP}",
			)
			//fmt.Println(splitPre)
			taskMsg := strings.Split(splitPre, "{TEMP}")
			pidTemp := strings.ReplaceAll(taskMsg[1], " ", "")
			pidTemp = strings.ReplaceAll(pidTemp, "Services", "")
			pidTemp = strings.ReplaceAll(pidTemp, "Console", "")

			process = append(process, service.Process{
				Name: taskMsg[0],
				Pid:  pidTemp,
			})
		}
	}

	taskMap := make(map[string]service.Process)
	for _, pro := range process {
		taskMap[pro.Pid] = pro
		//fmt.Println(pro.Name + " -- " + pro.Pid)
	}

	// get all task running http port
	output := utils.CmdRun("netstat", "-ano")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Index(line, "ESTABLISHED") >= 0 || strings.Index(line, "LISTENING") >= 0 {
			//fmt.Println(line)
			netstatLineTemp := replaceAllTwoStrToOne(line, "  ", " ")
			netstatLineTemp = strings.Replace(netstatLineTemp, " ", "", 1)
			netstatLineTemp = strings.Replace(netstatLineTemp, "\r", "", -1)
			netstatLineTemp = strings.Replace(netstatLineTemp, "[::]", "0.0.0.0", -1)

			netstatMsg := strings.Split(netstatLineTemp, " ")
			if _, ok := taskMap[netstatMsg[4]]; ok {
				processTemp := taskMap[netstatMsg[4]]
				if processTemp.Port == "" {
					processTemp.Port = strings.Split(netstatMsg[1], ":")[1]
					taskMap[netstatMsg[4]] = processTemp
				}
				//fmt.Println("地址", netstatMsg[1])
			}

		}
	}
	var resList []service.Process
	for _, value := range taskMap {
		if strings.Contains(value.Name, "miniServer.exe") && value.Port != "" {
			resList = append(resList, value)
		}
	}
	return resList

}

func replaceAllTwoStrToOne(str string, oldStr string, newStr string) string {
	if strings.Index(str, oldStr) >= 0 {
		return replaceAllTwoStrToOne(strings.ReplaceAll(str, oldStr, newStr), oldStr, newStr)
	} else {
		return str
	}
}

func cmdWithoutArgs(name string) string {
	var cmdOutput []byte
	var err error
	var cmd *exec.Cmd

	// 执行单个shell命令时, 直接运行即可
	cmd = exec.Command(name)
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println(name + " 命令调用错误")
		os.Exit(1)
	}

	//cmdOutPutUTF8, err := GbkToUtf8(cmdOutput)
	//return string(cmdOutPutUTF8)
	return string(cmdOutput)
}

func killProcessWin(process service.Process) bool {
	var err error
	var cmd *exec.Cmd
	cmd = exec.Command("tskill", process.Pid)
	if _, err = cmd.Output(); err != nil {
		fmt.Println("tskill 线程失败, 错误 :", err)
		return false
	}
	fmt.Println("tskill 端口", process.Port, "上的 miniServer 进程")
	return true
}
