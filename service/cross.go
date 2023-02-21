package service

import (
	"fmt"
	"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/MiniServer/conf"
	"github.com/Ericwyn/MiniServer/utils"
	"github.com/shirou/gopsutil/v3/process"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
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
	//h := http.FileServer(http.Dir(dirPath))
	fmt.Println("-----------------------------------")
	fmt.Println("服务将启动在以下地址")
	for _, ip := range ipAddrArr {
		fmt.Println("http://" + ip + ":" + port)
	}
	fmt.Println("-----------------------------------")

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":"+port, nil)

	//h := http.FileServer(http.Dir(dirPath))
	//err := http.ListenAndServe(":"+port, h)
	if err != nil {
		if err != nil {
			fmt.Println(err)
		}
		log.Fatal("文件服务器启动失败: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	pathParam := r.URL.String()
	if pathParam == "" || strings.Contains(pathParam, "./") {
		pathParam = ""
	}

	filePath := conf.RunDirPath + pathParam
	file := file.OpenFile(filePath)
	if !file.Exits() {
		fmt.Fprintf(w, "404 error")
		return
	}

	if file.IsDir() {
		// 返回 html
		w.Header().Set("Content-Type", "text/html")
		var fileList []fileMsgVO
		var dirList []fileMsgVO
		for _, f := range file.Children() {
			vo := fileMsgVO{
				FileName: f.Name(),
				IsDir:    f.IsDir(),
			}
			if f.IsDir() {
				vo.FileSize = "文件夹"
				dirList = append(dirList, vo)
			} else {
				vo.FileSize = utils.HumanizeFileSize(f.Size())
				fileList = append(fileList, vo)
			}
		}

		sort.Slice(fileList, func(i, j int) bool {
			return strings.Compare(fileList[i].FileName, fileList[j].FileName) < 0
		})
		sort.Slice(dirList, func(i, j int) bool {
			return strings.Compare(dirList[i].FileName, dirList[j].FileName) < 0
		})
		dirList = append(dirList, fileList...)
		html := renderHtml(dirList)

		io.WriteString(w, html)
	} else if file.IsFile() {
		openFile, err := file.Open()
		if err != nil {
			return
		}
		defer openFile.Close()

		// 获取文件信息
		fileInfo, err := openFile.Stat()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 设置响应头
		//w.Header().Set("Content-Disposition", "attachment; filename="+file.Name())
		http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), openFile)
	}
}
