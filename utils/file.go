package utils

import "fmt"

// HumanizeFileSize 函数, chatGpt 写的
// 它将文件大小（以字节为单位）转换为易于阅读的字符串格式，例如 "1.23 MB" 或 "512 Bytes"。
// 该函数首先检查文件大小是否小于1KB，如果是，则返回字节数的字符串表示形式。
// 否则，它将循环遍历整个文件大小并找到一个最接近的单位（KB、MB、GB 等）。
// 然后，它将文件大小除以该单位并将结果格式化为带有单位的字符串。
// 例如，如果文件大小为 1500 字节，则将其转换为 "1.46 KB"。
func HumanizeFileSize(fileSizeBytes int64) string {
	const unit = 1024
	if fileSizeBytes < unit {
		return fmt.Sprintf("%d Bytes", fileSizeBytes)
	}
	div, exp := int64(unit), 0
	for n := fileSizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(fileSizeBytes)/float64(div), "KMGTPE"[exp])
}
