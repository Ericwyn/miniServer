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
		killAllThread()
		os.Exit(0)
	}else {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		dirPath := strings.Replace(dir, "\\", "/", -1)
		fmt.Println(dirPath)
		h := http.FileServer(http.Dir(dirPath))
		ports := ":"+*port
		fmt.Println("服务启动在"+ports)
		err2 := http.ListenAndServe(ports, h)
		if err2 != nil {
			log.Fatal("ListenAndServe: ", err2)
		}
	}
}

type Thread struct {
	Port string
	Pid string
}

func getAllRunning() ([]Thread) {
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
	var resList []Thread
	for _,line := range lines{
		if strings.Contains(line,"miniServer") {
			resList = append(resList,Thread{
				Port:strings.Replace(portRegex.FindString(line),":","",-1),
				Pid:strings.Replace(pidRegex.FindString(line),"/miniServer","",-1),
			})
		}
	}
	return resList
}

func killThread(thread Thread) bool {
	var err error
	var cmd *exec.Cmd
	cmd = exec.Command("kill",thread.Pid)
	if _, err = cmd.Output(); err != nil {
		fmt.Println("kill 线程失败, 错误 :",err)
		return false
	}
	fmt.Println("kill 端口",thread.Port,"上的 miniServer 进程")
	return true
}

func killAllThread() {
	for _,th := range getAllRunning(){
		killThread(th)
	}
}

func killAPort(port int) bool {
	for _,th := range getAllRunning(){
		if th.Port == strconv.Itoa(port) {
			return killThread(th)
		}
	}
	fmt.Println("不存在运行端口为",port,"的 miniServer 进程")
	return false
}