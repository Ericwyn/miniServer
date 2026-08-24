package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Ericwyn/MiniServer/utils"
)

const (
	readHeaderTimeout = 15 * time.Second
	idleTimeout       = 2 * time.Minute
	maxHeaderBytes    = 1 << 20
)

var errInvalidRequestPath = errors.New("invalid request path")

type fileServerHandler struct {
	root string
}

func newFileServerHandler(root string) (*fileServerHandler, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	return &fileServerHandler{root: filepath.Clean(root)}, nil
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		// A write timeout would terminate legitimate slow or very large downloads.
		// Active transfers are therefore intentionally allowed to run to completion.
		WriteTimeout: 0,
	}
}

func (h *fileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestPath := r.URL.Path
	if requestPath == "" {
		requestPath = "/"
	}

	filePath, err := h.resolvePath(requestPath)
	if err != nil {
		http.Error(w, "bad request path", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		writeFileError(w, r, err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		writeFileError(w, r, err)
		return
	}

	if info.IsDir() {
		if !strings.HasSuffix(requestPath, "/") {
			redirectPath(w, r, requestPath+"/")
			return
		}
		h.serveDirectory(w, r, file, requestPath)
		return
	}

	if strings.HasSuffix(requestPath, "/") {
		redirectPath(w, r, strings.TrimSuffix(requestPath, "/"))
		return
	}
	if !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	// ServeContent provides standards-compliant GET/HEAD, Content-Length,
	// Last-Modified, conditional requests, and single or multipart byte ranges.
	// Supplying an ETag also makes If-Range safe for download resumption.
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fileETag(info))
	w.Header().Set("Cache-Control", "no-transform")
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (h *fileServerHandler) resolvePath(requestPath string) (string, error) {
	if strings.IndexByte(requestPath, 0) >= 0 {
		return "", errInvalidRequestPath
	}
	for _, segment := range strings.Split(requestPath, "/") {
		if segment == ".." {
			return "", errInvalidRequestPath
		}
	}

	cleanURLPath := path.Clean("/" + requestPath)
	relativePath := strings.TrimPrefix(cleanURLPath, "/")
	target := filepath.Join(h.root, filepath.FromSlash(relativePath))

	rel, err := filepath.Rel(h.root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errInvalidRequestPath
	}
	return target, nil
}

func (h *fileServerHandler) serveDirectory(w http.ResponseWriter, r *http.Request, dir *os.File, requestPath string) {
	entries, err := dir.ReadDir(-1)
	if err != nil {
		http.Error(w, "unable to read directory", http.StatusInternalServerError)
		return
	}

	files := make([]fileMsgVO, 0, len(entries))
	var totalFileSize int64
	fileCount := 0
	dirCount := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		href := url.PathEscape(entry.Name())
		category, typeLabel := classifyEntry(entry.Name(), info.IsDir())
		vo := fileMsgVO{
			FileName:    entry.Name(),
			Href:        href,
			Category:    category,
			TypeLabel:   typeLabel,
			IsDir:       info.IsDir(),
			Previewable: isTextPreviewable(entry.Name(), info.IsDir()),
			LastMod:     info.ModTime().Format("2006-01-02 15:04"),
			ModISO:      info.ModTime().Format(time.RFC3339),
			ModUnix:     info.ModTime().Unix(),
			RawSize:     info.Size(),
		}
		if info.IsDir() {
			vo.FileSize = "文件夹"
			vo.RawSize = -1
			vo.Href += "/"
			dirCount++
		} else {
			vo.FileSize = utils.HumanizeFileSize(info.Size())
			totalFileSize += info.Size()
			fileCount++
		}
		files = append(files, vo)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].FileName < files[j].FileName
	})

	cleanRequestPath := path.Clean("/" + requestPath)
	if cleanRequestPath != "/" {
		cleanRequestPath += "/"
	}
	body, err := renderHtml(directoryPage{
		CurrentPath: cleanRequestPath,
		TotalSize:   utils.HumanizeFileSize(totalFileSize),
		Files:       files,
		Breadcrumbs: buildBreadcrumbs(cleanRequestPath),
		ItemCount:   len(files),
		FileCount:   fileCount,
		DirCount:    dirCount,
		HasParent:   cleanRequestPath != "/",
	})
	if err != nil {
		http.Error(w, "unable to render directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; connect-src 'self'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(body))
}

func buildBreadcrumbs(requestPath string) []breadcrumbVO {
	cleanPath := path.Clean("/" + requestPath)
	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if cleanPath == "/" {
		return []breadcrumbVO{{Name: "ROOT", Href: "/", Current: true}}
	}

	breadcrumbs := make([]breadcrumbVO, 0, len(parts)+1)
	breadcrumbs = append(breadcrumbs, breadcrumbVO{Name: "ROOT", Href: "/"})
	currentHref := "/"
	for index, part := range parts {
		currentHref += url.PathEscape(part) + "/"
		breadcrumbs = append(breadcrumbs, breadcrumbVO{
			Name:    part,
			Href:    currentHref,
			Current: index == len(parts)-1,
		})
	}
	return breadcrumbs
}

func classifyEntry(name string, isDir bool) (category string, typeLabel string) {
	if isDir {
		return "folder", "DIR"
	}

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if extension == "" {
		return "file", "FILE"
	}
	typeLabel = strings.ToUpper(extension)
	if runes := []rune(typeLabel); len(runes) > 8 {
		typeLabel = string(runes[:7]) + "…"
	}

	switch extension {
	case "avif", "bmp", "gif", "heic", "ico", "jpeg", "jpg", "png", "raw", "svg", "tif", "tiff", "webp":
		category = "image"
	case "aac", "avi", "flac", "m4a", "mkv", "mov", "mp3", "mp4", "ogg", "opus", "wav", "webm":
		category = "media"
	case "bash", "c", "cc", "cpp", "css", "go", "h", "hpp", "html", "java", "js", "json", "jsx", "py", "rs", "scss", "sh", "sql", "toml", "ts", "tsx", "xml", "yaml", "yml", "zsh":
		category = "code"
	case "7z", "apk", "bz2", "dmg", "gz", "iso", "rar", "tar", "tgz", "xz", "zip":
		category = "archive"
	case "csv", "doc", "docx", "epub", "log", "md", "pdf", "ppt", "pptx", "rtf", "txt", "xls", "xlsx":
		category = "document"
	default:
		category = "file"
	}
	return category, typeLabel
}

func isTextPreviewable(name string, isDir bool) bool {
	if isDir {
		return false
	}

	baseName := strings.ToLower(filepath.Base(name))
	switch baseName {
	case ".dockerignore", ".editorconfig", ".env", ".gitattributes", ".gitignore",
		"authors", "changelog", "contributors", "copying", "dockerfile", "go.mod",
		"go.sum", "licence", "license", "makefile", "readme":
		return true
	}
	if strings.HasPrefix(baseName, ".env.") {
		return true
	}

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(baseName)), ".")
	switch extension {
	case "bash", "bat", "c", "cc", "cfg", "cjs", "cmd", "conf", "config", "cpp", "cs", "css", "csv", "cxx",
		"dockerignore", "editorconfig", "env", "fish", "gitattributes", "gitignore", "go", "gql", "graphql", "h", "hpp",
		"htm", "html", "ini", "java", "js", "json", "jsonl", "jsx", "kt", "kts", "less", "lock", "log", "markdown",
		"md", "mjs", "mod", "ndjson", "npmrc", "php", "properties", "proto", "ps1", "py", "rb", "rs", "sass", "scala",
		"scss", "sh", "sql", "sum", "svelte", "svg", "swift", "text", "toml", "ts", "tsv", "tsx", "txt", "vue", "xml",
		"yaml", "yml", "yarnrc", "zsh":
		return true
	default:
		return false
	}
}

func redirectPath(w http.ResponseWriter, r *http.Request, target string) {
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func fileETag(info os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, info.ModTime().UnixNano(), info.Size())
}

func writeFileError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case os.IsNotExist(err):
		http.NotFound(w, r)
	case os.IsPermission(err):
		http.Error(w, "permission denied", http.StatusForbidden)
	default:
		http.Error(w, "unable to open file", http.StatusInternalServerError)
	}
}
