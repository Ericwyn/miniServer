package main

import (
	"flag"
	"fmt"
	"github.com/Ericwyn/MiniServer/conf"
	"github.com/Ericwyn/MiniServer/service"
	"github.com/Ericwyn/MiniServer/utils"
	"os"
)

var port = flag.String("p", "10010", "http listen port \n设置 http 服务器运行端口")
var kPort = flag.Int("k", -1, "kill the miniServer running on this port \n杀死在某个端口运行的 miniServer")
var killFlag = flag.Bool("kl", false, "kill the cross running miniServer \n杀死所有的 miniServer 程序")
var list = flag.Bool("l", false, "list the status of cross running miniServer \n列出当前运行的所有 miniServer 程序")
var dir = flag.String("d", "null", "the dir path the miniServer listen, default is current path \n设置http服务器鉴定目录，默认为当前目录")
var ver = flag.Bool("v", false, "version message \n版本信息查看")

func main() {
	flag.Parse()

	if *ver {
		fmt.Println(conf.VersionStr)
		os.Exit(0)
	}

	ipAddrArr := utils.GetOrSelectIPv4Addr()

	var cmd = service.Cmd

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
		fmt.Println(conf.VersionStr)
		cmd.Run(*dir, *port, ipAddrArr)
	}
}
