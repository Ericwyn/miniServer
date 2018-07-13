package main
import (
	"log"
	"net/http"
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"flag"
	"os/exec"
	"regexp"
	"strconv"
)

var port = flag.String("p", "10010", "http listen port")
var kPort = flag.Int("k",-1,"kill the miniServer running on this port")
var killFlag = flag.Bool("kl",false,"kill the all running miniServer")
var list = flag.Bool("l",false,"list the status of all running miniServer")
var dir = flag.String("d","null","the dir path the miniServer listen, default is current path")

func main() {
	flag.Parse()
	if *kPort != -1 {
		if killAPort(*kPort) {
			fmt.Println("成功停止运行端口号为",*port,"的 miniServer 进程")
		}else {
			fmt.Println("无法停止运行端口号为",*port,"的 miniServer 进程")
		}
		os.Exit(0)
	}else if *killFlag {
		killAllProcess()
		os.Exit(0)
	}else if *list {
		fmt.Println("运行端口"+"\t"+"进程id"+"\t\t"+"监听位置")
		for _,thread := range getAllRunning() {
			fmt.Println(thread.Port+"\t\t"+thread.Pid+"\t\t"+thread.DirPath)
		}
	}else {
		if *dir == "null" {
			fmt.Println("未使用 -d 指定监听路径")
			runInNewProcress()
		}else {
			dirPath := strings.Replace(*dir, "\\", "/", -1)
			fmt.Println("监听:"+dirPath)
			h := http.FileServer(http.Dir(dirPath))
			ports := ":"+*port
			fmt.Println("服务启动在"+ports)
			err2 := http.ListenAndServe(ports, h)
			if err2 != nil {
				log.Fatal("ListenAndServe: ", err2)
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
	Port string
	Pid string
	DirPath string
}

func getAllRunning() ([]Process) {
	var cmdOutput []byte
	var err error
	var cmd *exec.Cmd

	// 执行单个shell命令时, 直接运行即可
	cmd = exec.Command("netstat","-nlp")
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println("netstate 命令调用错误")
		os.Exit(1)
	}
	lines := strings.Split(string(cmdOutput),"\n")
	portRegex := regexp.MustCompile(`:::[0-9]{1,5}`)
	pidRegex := regexp.MustCompile(`[0-9]{1,5}\/miniServer`)
	var resList []Process
	for _,line := range lines{
		if strings.Contains(line,"miniServer") {
			pid := strings.Replace(pidRegex.FindString(line),"/miniServer","",-1)
			resList = append(resList, Process{
				Port:strings.Replace(portRegex.FindString(line),":","",-1),
				Pid: pid,
				DirPath:getListenDir(pid),
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
	cmd = exec.Command("ps","-efH")
	if cmdOutput, err = cmd.Output(); err != nil {
		fmt.Println(err)
		fmt.Println("netstate 命令调用错误")
		os.Exit(1)
	}
	lines := strings.Split(string(cmdOutput),"\n")
	for _,line := range lines{
		if strings.Contains(line,pid) &&
			strings.Contains(line, "miniServer") && strings.Contains(line, "-d") {
				temp := strings.Split(line,"-d")[1]
				if strings.Contains(temp,"-p") {
					temp = strings.Split(temp,"-p")[0]
				}
				return strings.Replace(temp," ","",-1)
		}
	}
	return "null"
}

func killProcess(process Process) bool {
	var err error
	var cmd *exec.Cmd
	cmd = exec.Command("kill", process.Pid)
	if _, err = cmd.Output(); err != nil {
		fmt.Println("kill 线程失败, 错误 :",err)
		return false
	}
	fmt.Println("kill 端口", process.Port,"上的 miniServer 进程")
	return true
}

func killAllProcess() {
	for _,th := range getAllRunning(){
		killProcess(th)
	}
}

func killAPort(port int) bool {
	for _,th := range getAllRunning(){
		if th.Port == strconv.Itoa(port) {
			return killProcess(th)
		}
	}
	fmt.Println("不存在运行端口为",port,"的 miniServer 进程")
	return false
}

// 在新的线程里面开启
func runInNewProcress() {
	var cmd *exec.Cmd
	if *port != "10010" {
		fmt.Println("使用命令:\n\t",GetBinPath()+"/miniServer","-d",getCurPath(),"-p",*port,"\n创建新进程")
		fmt.Println("服务运行在: "+*port+"端口")
		cmd = exec.Command(GetBinPath()+"/miniServer","-d",getCurPath(),"-p",*port)
	}else {
		fmt.Println("使用命令:\n\t",GetBinPath()+"/miniServer","-d",getCurPath(),"\n创建新进程")
		fmt.Println("服务运行在: 10010 端口")
		cmd = exec.Command(GetBinPath()+"/miniServer","-d",getCurPath())
	}
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("创建新线程失败或遭退出, :",err)
	}
}