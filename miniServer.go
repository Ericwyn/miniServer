package main

import (
	"flag"
	"fmt"
	"github.com/Ericwyn/MiniServer/service"
	"github.com/Ericwyn/MiniServer/service/linux"
	"github.com/Ericwyn/MiniServer/service/win"
	"github.com/Ericwyn/MiniServer/utils"
	"os"
	"runtime"
)

var port = flag.String("p", "10010", "http listen port \n设置 http 服务器运行端口")
var kPort = flag.Int("k", -1, "kill the miniServer running on this port \n杀死在某个端口运行的 miniServer")
var killFlag = flag.Bool("kl", false, "kill the all running miniServer \n杀死所有的 miniServer 程序")
var list = flag.Bool("l", false, "list the status of all running miniServer \n列出当前运行的所有 miniServer 程序")
var dir = flag.String("d", "null", "the dir path the miniServer listen, default is current path \n设置http服务器鉴定目录，默认为当前目录")
var ver = flag.Bool("v", false, "version message \n版本信息查看")

var logo = "              _       _ _____                          \n" +
	"   ____ ___  (_)___  (_) ___/___  ______   _____  _____\n" +
	"  / __ `__ \\/ / __ \\/ /\\__ \\/ _ \\/ ___/ | / / _ \\/ ___/\n" +
	" / / / / / / / / / / /___/ /  __/ /   | |/ /  __/ /    \n" +
	"/_/ /_/ /_/_/_/ /_/_//____/\\___/_/    |___/\\___/_/     \n\n"
var versionStr = logo +
	"                v1.2 - 2023.02.20 \n" +
	"   @Ericwyn https://github.com/Ericwyn/miniServer\n"

type MiniServerCmd struct {
	SystemName  string // 系统名称
	KillPortFun func(port int)
	KillAllFun  func()
	ListFun     func()
}

func main() {
	flag.Parse()

	if *ver {
		fmt.Println(versionStr)
		os.Exit(0)
	}

	ipAddrArr := utils.GetOrSelectIPv4Addr()

	// 相对路径转绝对路径
	//absDirPath := *dir
	//if absDirPath != "null" && !filepath.IsAbs(absDirPath) {
	//	absDirPath, _ = filepath.Abs(absDirPath)
	//}

	var cmd service.MiniServerCmd
	if runtime.GOOS == "windows" {
		cmd = win.WinMiniServerCmd
	} else {
		cmd = linux.LinuxMiniServerCmd
	}

	if *kPort != -1 {
		if cmd.KillPortFun(*kPort) {
			fmt.Println("成功停止运行端口号为", *port, "的 miniServer 进程")
		} else {
			fmt.Println("无法停止运行端口号为", *port, "的 miniServer 进程")
		}
		os.Exit(0)
	} else if *killFlag {
		cmd.KillAllFun()
		os.Exit(0)
	} else if *list {
		fmt.Println("运行端口" + "\t" + "进程id" + "\t\t" + "监听位置")
		for _, thread := range cmd.ListFun() {
			fmt.Println(thread.Port + "\t\t" + thread.Pid + "\t\t" + thread.DirPath)
		}
	} else {
		fmt.Println(versionStr)
		cmd.Run(*dir, *port, ipAddrArr)
	}

	//if runtime.GOOS == "linux" {
	//	if *kPort != -1 {
	//		if utils.KillAPort(*kPort) {
	//			fmt.Println("成功停止运行端口号为", *port, "的 miniServer 进程")
	//		} else {
	//			fmt.Println("无法停止运行端口号为", *port, "的 miniServer 进程")
	//		}
	//		os.Exit(0)
	//	} else if *killFlag {
	//		utils.KillAllProcess()
	//		os.Exit(0)
	//	} else if *list {
	//		fmt.Println("运行端口" + "\t" + "进程id" + "\t\t" + "监听位置")
	//		for _, thread := range utils.GetAllRunning() {
	//			fmt.Println(thread.Port + "\t\t" + thread.Pid + "\t\t" + thread.DirPath)
	//		}
	//	} else {
	//		if *dir == "null" {
	//			fmt.Println("未使用 -d 指定监听路径")
	//			runInNewProcress(ipAddrArr)
	//		} else {
	//			dirPath := strings.Replace(*dir, "\\", "/", -1)
	//			fmt.Println("监听:" + dirPath)
	//			h := http.FileServer(http.Dir(dirPath))
	//			ports := ":" + *port
	//			fmt.Println("服务启动在" + ports)
	//			for _, ip := range ipAddrArr {
	//				fmt.Println("http://" + ip + ports)
	//			}
	//			err2 := http.ListenAndServe(ports, h)
	//			if err2 != nil {
	//				log.Fatal("ListenAndServe: ", err2)
	//			}
	//		}
	//	}
	//} else if runtime.GOOS == "windows" {
	//	if *kPort != -1 {
	//		if win.KillAPortWin(*kPort) {
	//			fmt.Println("成功停止运行端口号为", *port, "的 miniServer 进程")
	//		} else {
	//			fmt.Println("无法停止运行端口号为", *port, "的 miniServer 进程")
	//		}
	//		os.Exit(0)
	//	} else if *killFlag {
	//		win.KillAllProcessWin()
	//		os.Exit(0)
	//	} else if *list {
	//		fmt.Println("运行端口" + "\t" + "进程id" + "\t\t" + "监听位置")
	//		for _, thread := range win.GetAllRunningWin() {
	//			fmt.Println(thread.Port + "\t\t" + thread.Pid + "\t\t 当前系统无法获取监听位置")
	//		}
	//	} else {
	//		if *dir == "null" {
	//			fmt.Println("未使用 -d 指定监听路径")
	//			runInNewProcress(ipAddrArr)
	//		} else {
	//			dirPath := strings.Replace(*dir, "\\", "/", -1)
	//			fmt.Println("监听:" + dirPath)
	//			h := http.FileServer(http.Dir(dirPath))
	//			ports := ":" + *port
	//			fmt.Println("服务启动在" + ports)
	//			for _, ip := range ipAddrArr {
	//				fmt.Println("http://" + ip + ports)
	//			}
	//			err2 := http.ListenAndServe(ports, h)
	//			if err2 != nil {
	//				log.Fatal("ListenAndServe: ", err2)
	//			}
	//		}
	//	}
	//}

}

// 在新的线程里面开启
//func runInNewProcress(ipAddrArr []string) {
//	var cmd *exec.Cmd
//	if *port != "10010" {
//		fmt.Println("使用命令:\n\t", utils.GetBinPath()+"/miniServer", "-d", utils.GetCurPath(), "-p", *port, "\n创建新进程")
//		for _, ip := range ipAddrArr {
//			fmt.Println("服务启动在 ", "http://"+ip+":"+*port)
//		}
//		cmd = exec.Command(utils.GetBinPath()+"/miniServer", "-d", utils.GetCurPath(), "-p", *port)
//	} else {
//		fmt.Println("使用命令:\n\t", utils.GetBinPath()+"/miniServer", "-d", utils.GetCurPath(), "\n创建新进程")
//		fmt.Println("服务运行在: 10010 端口")
//		for _, ip := range ipAddrArr {
//			fmt.Println("http://" + ip + ":10010")
//		}
//		cmd = exec.Command(utils.GetBinPath()+"/miniServer", "-d", utils.GetCurPath())
//	}
//	_, err := cmd.Output()
//	if err != nil {
//		fmt.Println("创建新线程失败或遭退出, :", err)
//	}
//}
