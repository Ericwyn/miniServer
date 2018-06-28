package main
import (
	"log"
	"net/http"
	"fmt"
	"path/filepath"
	"os"
	"strings"
)
func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dirPath := strings.Replace(dir, "\\", "/", -1)
	fmt.Println(dirPath)
	h := http.FileServer(http.Dir(dirPath))
	fmt.Println("服务启动在 10010")
	err2 := http.ListenAndServe(":10010", h)
	if err2 != nil {
		log.Fatal("ListenAndServe: ", err2)
	}
}