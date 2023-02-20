package service

import (
	"github.com/shirou/gopsutil/v3/process"
)

type MiniServerCmd struct {
	SystemName  string // 系统名称
	KillPortFun func(port int) bool
	KillAllFun  func()
	ListFun     func() []Process
	Run         func(dirPath string, port string, ipAddrArr []string)
}

type Process struct {
	Port    string
	Pid     string
	DirPath string
	Name    string
	Process *process.Process
}
