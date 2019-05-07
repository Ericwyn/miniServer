package main

import (
	"flag"
	"fmt"

	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var port = flag.String("p", "10010", "http listen port \n设置 http 服务器运行端口")
var kPort = flag.Int("k", -1, "kill the miniServer running on this port \n杀死在某个端口运行的 miniServer")
var killFlag = flag.Bool("kl", false, "kill the all running miniServer \n杀死所有的 miniServer 程序")
var list = flag.Bool("l", false, "list the status of all running miniServer \n列出当前运行的所有 miniServer 程序")
var dir = flag.String("d", "null", "the dir path the miniServer listen, default is current path \n设置http服务器鉴定目录，默认为当前目录")
var ver = flag.Bool("v", false, "version message \n版本信息查看")

var versionStr = "miniServer v1.1; author: @Ericwyn; github.com/Ericwyn/miniServer;"

func main() {
	flag.Parse()

	if *ver {
		fmt.Println(versionStr)
		os.Exit(0)
	}

	if runtime.GOOS == "linux" {
		if *kPort != -1 {
			if killAPort(*kPort) {
				fmt.Println("成功停止运行端口号为", *port, "的 miniServer 进程")
			} else {
				fmt.Println("无法停止运行端口号为", *port, "的 miniServer 进程")
			}
			os.Exit(0)
		} else if *killFlag {
			killAllProcess()
			os.Exit(0)
		} else if *list {
			fmt.Println("运行端口" + "\t" + "进程id" + "\t\t" + "监听位置")
			for _, thread := range getAllRunning() {
				fmt.Println(thread.Port + "\t\t" + thread.Pid + "\t\t" + thread.DirPath)
			}
		} else {
			if *dir == "null" {
				fmt.Println("未使用 -d 指定监听路径")
				runInNewProcress()
			} else {
				dirPath := strings.Replace(*dir, "\\", "/", -1)
				fmt.Println("监听:" + dirPath)
				h := http.FileServer(http.Dir(dirPath))
				ports := ":" + *port
				fmt.Println("服务启动在" + ports)
				err2 := http.ListenAndServe(ports, h)
				if err2 != nil {
					log.Fatal("ListenAndServe: ", err2)
				}
			}
		}
	} else if runtime.GOOS == "windows" {
		if *kPort != -1 {
			if killAPortWin(*kPort) {
				fmt.Println("成功停止运行端口号为", *port, "的 miniServer 进程")
			} else {
				fmt.Println("无法停止运行端口号为", *port, "的 miniServer 进程")
			}
			os.Exit(0)
		} else if *killFlag {
			killAllProcessWin()
			os.Exit(0)
		} else if *list {
			fmt.Println("运行端口" + "\t" + "进程id" + "\t\t" + "监听位置")
			for _, thread := range getAllRunningWin() {
				fmt.Println(thread.Port + "\t\t" + thread.Pid + "\t\t 当前系统无法获取监听位置")
			}
		} else {
			if *dir == "null" {
				fmt.Println("未使用 -d 指定监听路径")
				runInNewProcress()
			} else {
				dirPath := strings.Replace(*dir, "\\", "/", -1)
				fmt.Println("监听:" + dirPath)
				h := http.FileServer(http.Dir(dirPath))
				ports := ":" + *port
				fmt.Println("服务启动在" + ports)
				err2 := http.ListenAndServe(ports, h)
				if err2 != nil {
					log.Fatal("ListenAndServe: ", err2)
				}
			}
		}
	}

}

// 获取执行文件夹
func getCurPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

// 获取可执行文件的绝对路径
func GetBinPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	rst := filepath.Dir(path)
	return rst
}

type Process struct {
	Port    string
	Pid     string
	DirPath string
	Name    string
}

func getAllRunning() []Process {
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
	portRegex := regexp.MustCompile(`:::[0-9]{1,5}`)
	pidRegex := regexp.MustCompile(`[0-9]{1,5}\/miniServer`)
	var resList []Process
	for _, line := range lines {
		if strings.Contains(line, "miniServer") {
			pid := strings.Replace(pidRegex.FindString(line), "/miniServer", "", -1)
			resList = append(resList, Process{
				Port:    strings.Replace(portRegex.FindString(line), ":", "", -1),
				Pid:     pid,
				DirPath: getListenDir(pid),
			})
		}
	}
	return resList
}

func getListenDir(pid string) string {
	var cmdOutput []byte
	var err error
	var cmd *exec.Cmd

	// 执行单个shell命令时, 直接运行即可
	cmd = exec.Command("ps", "-efH")
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println("netstate 命令调用错误")
		os.Exit(1)
	}
	lines := strings.Split(string(cmdOutput), "\n")
	for _, line := range lines {
		if strings.Contains(line, pid) &&
			strings.Contains(line, "miniServer") && strings.Contains(line, "-d") {
			temp := strings.Split(line, "-d")[1]
			if strings.Contains(temp, "-p") {
				temp = strings.Split(temp, "-p")[0]
			}
			return strings.Replace(temp, " ", "", -1)
		}
	}
	return "null"
}

func killProcess(process Process) bool {
	var err error
	var cmd *exec.Cmd
	cmd = exec.Command("kill", process.Pid)
	if _, err = cmd.Output(); err != nil {
		fmt.Println("kill 线程失败, 错误 :", err)
		return false
	}
	fmt.Println("kill 端口", process.Port, "上的 miniServer 进程")
	return true
}

func killAllProcess() {
	for _, th := range getAllRunning() {
		killProcess(th)
	}
}

func killAPort(port int) bool {
	for _, th := range getAllRunning() {
		if th.Port == strconv.Itoa(port) {
			return killProcess(th)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false
}

// 在新的线程里面开启
func runInNewProcress() {
	var cmd *exec.Cmd
	if *port != "10010" {
		fmt.Println("使用命令:\n\t", GetBinPath()+"/miniServer", "-d", getCurPath(), "-p", *port, "\n创建新进程")
		fmt.Println("服务运行在: " + *port + "端口")
		cmd = exec.Command(GetBinPath()+"/miniServer", "-d", getCurPath(), "-p", *port)
	} else {
		fmt.Println("使用命令:\n\t", GetBinPath()+"/miniServer", "-d", getCurPath(), "\n创建新进程")
		fmt.Println("服务运行在: 10010 端口")
		cmd = exec.Command(GetBinPath()+"/miniServer", "-d", getCurPath())
	}
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("创建新线程失败或遭退出, :", err)
	}
}

// ------------------ 为 windows 做的适配

func getAllRunningWin() []Process {

	// get all tasklist
	var process []Process

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

			process = append(process, Process{
				Name: taskMsg[0],
				Pid:  pidTemp,
			})
		}
	}

	taskMap := make(map[string]Process)
	for _, pro := range process {
		taskMap[pro.Pid] = pro
		//fmt.Println(pro.Name + " -- " + pro.Pid)
	}

	// get all task running http port
	output := cmd("netstat", "-ano")

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
	var resList []Process
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

//func GbkToUtf8(s []byte) ([]byte, error) {
//	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
//	d, e := ioutil.ReadAll(reader)
//	if e != nil {
//		return nil, e
//	}
//	return d, nil
//}
//
//func Utf8ToGbk(s []byte) ([]byte, error) {
//	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
//	d, e := ioutil.ReadAll(reader)
//	if e != nil {
//		return nil, e
//	}
//	return d, nil
//}

func cmd(name string, arg string) string {
	var cmdOutput []byte
	var err error
	var cmd *exec.Cmd

	// 执行单个shell命令时, 直接运行即可
	cmd = exec.Command(name, arg)
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println(name + " 命令调用错误")
		os.Exit(1)
	}

	//cmdOutPutUTF8, err := GbkToUtf8(cmdOutput)
	//return string(cmdOutPutUTF8)
	return string(cmdOutput)
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

func killAPortWin(port int) bool {
	for _, th := range getAllRunningWin() {
		if th.Port == strconv.Itoa(port) {
			return killProcessWin(th)
		}
	}
	fmt.Println("不存在运行端口为", port, "的 miniServer 进程")
	return false
}

func killProcessWin(process Process) bool {
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

func killAllProcessWin() {
	for _, th := range getAllRunningWin() {
		killProcessWin(th)
	}
}
