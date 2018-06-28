package main
import (
	"log"
	"net/http"
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"flag"
)
func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	port := flag.String("p", "10010", "http listen port")
	dirPath := strings.Replace(dir, "\\", "/", -1)
	fmt.Println(dirPath)
	h := http.FileServer(http.Dir(dirPath))
	flag.Parse()
	ports := ":"+*port
	fmt.Println("服务启动在"+ports)
	err2 := http.ListenAndServe(ports, h)
	if err2 != nil {
		log.Fatal("ListenAndServe: ", err2)
	}
}