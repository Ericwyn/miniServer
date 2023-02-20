package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func CmdRun(name string, arg string) string {
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

	return string(cmdOutput)
}

func RunInNewProcess(port string, dirPath string, ipAddrArr []string) {
	var cmd *exec.Cmd
	if port != "10010" {
		fmt.Println("使用命令:\n\t", GetBinPath()+"/miniServer", "-d", dirPath, "-p", port, "\n创建新进程")
		for _, ip := range ipAddrArr {
			fmt.Println("服务启动在 ", "http://"+ip+":"+port)
		}
		cmd = exec.Command(GetBinPath()+"/miniServer", "-d", dirPath, "-p", port)
	} else {
		fmt.Println("使用命令:\n\t", GetBinPath()+"/miniServer", "-d", dirPath, "\n创建新进程")
		fmt.Println("服务运行在: 10010 端口")
		for _, ip := range ipAddrArr {
			fmt.Println("http://" + ip + ":10010")
		}
		cmd = exec.Command(GetBinPath()+"/miniServer", "-d", dirPath)
	}
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("创建新线程失败或遭退出, :", err)
	}
}
