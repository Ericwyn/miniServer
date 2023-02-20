package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetListenDir(pid string) string {
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

// GetBinPath 获取可执行文件的绝对路径
func GetBinPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	rst := filepath.Dir(path)
	return rst
}

// GetCurPath 获取执行文件夹
func GetCurPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
