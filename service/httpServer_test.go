package service

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	webassets "github.com/Ericwyn/MiniServer/web"
)

func TestFileResponseHeadersAndMethods(t *testing.T) {
	content := []byte("0123456789")
	server, _ := newTestFileServer(t, map[string][]byte{"sample.bin": content})

	response := doRequest(t, server.Client(), http.MethodGet, server.URL+"/sample.bin?token=kept-out-of-path", nil)
	body := readResponseBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if !bytes.Equal(body, content) {
		t.Fatalf("GET body = %q, want %q", body, content)
	}
	if got := response.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges = %q, want bytes", got)
	}
	if got := response.Header.Get("Content-Length"); got != strconv.Itoa(len(content)) {
		t.Errorf("Content-Length = %q, want %d", got, len(content))
	}
	if got := response.Header.Get("ETag"); got == "" {
		t.Error("ETag is empty")
	}
	if got := response.Header.Get("Last-Modified"); got == "" {
		t.Error("Last-Modified is empty")
	}

	response = doRequest(t, server.Client(), http.MethodHead, server.URL+"/sample.bin", nil)
	body = readResponseBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("HEAD status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if len(body) != 0 {
		t.Fatalf("HEAD returned %d body bytes, want 0", len(body))
	}
	if response.ContentLength != int64(len(content)) {
		t.Errorf("HEAD Content-Length = %d, want %d", response.ContentLength, len(content))
	}

	response = doRequest(t, server.Client(), http.MethodPost, server.URL+"/sample.bin", nil)
	_ = readResponseBody(t, response)
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST status = %d, want %d", response.StatusCode, http.StatusMethodNotAllowed)
	}
	if got := response.Header.Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want GET, HEAD", got)
	}

	response = doRequest(t, server.Client(), http.MethodGet, server.URL+"/missing.bin", nil)
	_ = readResponseBody(t, response)
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("missing file status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestSingleRangeAndConditionalResume(t *testing.T) {
	content := []byte("0123456789")
	server, _ := newTestFileServer(t, map[string][]byte{"sample.bin": content})
	fileURL := server.URL + "/sample.bin"

	tests := []struct {
		name        string
		rangeHeader string
		wantBody    string
		wantRange   string
		wantStatus  int
	}{
		{name: "bounded", rangeHeader: "bytes=2-5", wantBody: "2345", wantRange: "bytes 2-5/10", wantStatus: http.StatusPartialContent},
		{name: "open ended", rangeHeader: "bytes=6-", wantBody: "6789", wantRange: "bytes 6-9/10", wantStatus: http.StatusPartialContent},
		{name: "suffix", rangeHeader: "bytes=-3", wantBody: "789", wantRange: "bytes 7-9/10", wantStatus: http.StatusPartialContent},
		{name: "unsatisfiable", rangeHeader: "bytes=99-100", wantRange: "bytes */10", wantStatus: http.StatusRequestedRangeNotSatisfiable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := doRequest(t, server.Client(), http.MethodGet, fileURL, map[string]string{"Range": test.rangeHeader})
			body := readResponseBody(t, response)
			if response.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
			if got := response.Header.Get("Content-Range"); got != test.wantRange {
				t.Errorf("Content-Range = %q, want %q", got, test.wantRange)
			}
			if test.wantBody != "" && string(body) != test.wantBody {
				t.Errorf("body = %q, want %q", body, test.wantBody)
			}
		})
	}

	response := doRequest(t, server.Client(), http.MethodHead, fileURL, nil)
	_ = readResponseBody(t, response)
	etag := response.Header.Get("ETag")
	if etag == "" {
		t.Fatal("HEAD response did not contain ETag")
	}

	response = doRequest(t, server.Client(), http.MethodGet, fileURL, map[string]string{
		"Range":    "bytes=4-",
		"If-Range": etag,
	})
	body := readResponseBody(t, response)
	if response.StatusCode != http.StatusPartialContent || string(body) != "456789" {
		t.Fatalf("matching If-Range response = (%d, %q), want (206, %q)", response.StatusCode, body, "456789")
	}

	response = doRequest(t, server.Client(), http.MethodGet, fileURL, map[string]string{
		"Range":    "bytes=4-",
		"If-Range": `"stale-etag"`,
	})
	body = readResponseBody(t, response)
	if response.StatusCode != http.StatusOK || !bytes.Equal(body, content) {
		t.Fatalf("stale If-Range response = (%d, %q), want complete 200 response", response.StatusCode, body)
	}

	response = doRequest(t, server.Client(), http.MethodGet, fileURL, map[string]string{"If-None-Match": etag})
	body = readResponseBody(t, response)
	if response.StatusCode != http.StatusNotModified || len(body) != 0 {
		t.Fatalf("If-None-Match response = (%d, %d bytes), want (304, 0 bytes)", response.StatusCode, len(body))
	}
}

func TestMultipartRanges(t *testing.T) {
	server, _ := newTestFileServer(t, map[string][]byte{"sample.txt": []byte("0123456789")})
	response := doRequest(t, server.Client(), http.MethodGet, server.URL+"/sample.txt", map[string]string{
		"Range": "bytes=0-1,8-9",
	})
	defer response.Body.Close()

	if response.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusPartialContent)
	}
	mediaType, params, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}
	if mediaType != "multipart/byteranges" {
		t.Fatalf("media type = %q, want multipart/byteranges", mediaType)
	}

	reader := multipart.NewReader(response.Body, params["boundary"])
	want := []struct {
		contentRange string
		body         string
	}{
		{contentRange: "bytes 0-1/10", body: "01"},
		{contentRange: "bytes 8-9/10", body: "89"},
	}
	for index, expected := range want {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatalf("read multipart part %d: %v", index, err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("read multipart body %d: %v", index, err)
		}
		if got := part.Header.Get("Content-Range"); got != expected.contentRange {
			t.Errorf("part %d Content-Range = %q, want %q", index, got, expected.contentRange)
		}
		if string(body) != expected.body {
			t.Errorf("part %d body = %q, want %q", index, body, expected.body)
		}
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("extra multipart data: %v", err)
	}
}

func TestInterruptedAndConcurrentSegmentedDownload(t *testing.T) {
	content := make([]byte, 4<<20)
	for index := range content {
		content[index] = byte((index*31 + 7) % 251)
	}
	server, _ := newTestFileServer(t, map[string][]byte{"large.bin": content})
	fileURL := server.URL + "/large.bin"

	response := doRequest(t, server.Client(), http.MethodHead, fileURL, nil)
	_ = readResponseBody(t, response)
	etag := response.Header.Get("ETag")

	// Simulate a connection that stopped after receiving only a prefix, then
	// resume at the exact next byte with If-Range validation.
	response = doRequest(t, server.Client(), http.MethodGet, fileURL, nil)
	prefix := make([]byte, 128<<10)
	if _, err := io.ReadFull(response.Body, prefix); err != nil {
		t.Fatalf("read interrupted prefix: %v", err)
	}
	_ = response.Body.Close()

	resumeAt := len(prefix)
	response = doRequest(t, server.Client(), http.MethodGet, fileURL, map[string]string{
		"Range":    fmt.Sprintf("bytes=%d-", resumeAt),
		"If-Range": etag,
	})
	remainder := readResponseBody(t, response)
	resumed := append(append([]byte(nil), prefix...), remainder...)
	if response.StatusCode != http.StatusPartialContent || !bytes.Equal(resumed, content) {
		t.Fatalf("resumed download = status %d and %d bytes, want 206 and %d bytes", response.StatusCode, len(resumed), len(content))
	}

	const segments = 16
	assembled := make([]byte, len(content))
	segmentSize := (len(content) + segments - 1) / segments
	errors := make(chan error, segments)
	var waitGroup sync.WaitGroup

	for segment := 0; segment < segments; segment++ {
		start := segment * segmentSize
		end := start + segmentSize - 1
		if end >= len(content) {
			end = len(content) - 1
		}
		if start > end {
			continue
		}

		waitGroup.Add(1)
		go func(start, end int) {
			defer waitGroup.Done()
			request, err := http.NewRequest(http.MethodGet, fileURL, nil)
			if err != nil {
				errors <- err
				return
			}
			request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			request.Header.Set("If-Range", etag)
			response, err := server.Client().Do(request)
			if err != nil {
				errors <- err
				return
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				errors <- err
				return
			}
			if response.StatusCode != http.StatusPartialContent {
				errors <- fmt.Errorf("range %d-%d status = %d", start, end, response.StatusCode)
				return
			}
			if got, want := response.Header.Get("Content-Range"), fmt.Sprintf("bytes %d-%d/%d", start, end, len(content)); got != want {
				errors <- fmt.Errorf("Content-Range = %q, want %q", got, want)
				return
			}
			if len(body) != end-start+1 {
				errors <- fmt.Errorf("range %d-%d returned %d bytes", start, end, len(body))
				return
			}
			copy(assembled[start:end+1], body)
		}(start, end)
	}

	waitGroup.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if !bytes.Equal(assembled, content) {
		t.Fatal("concurrent range responses did not reassemble to the source file")
	}
}

func TestRangeBeyondFourGiB(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "huge.bin")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	tail := []byte("large-file-tail")
	fileSize := (int64(5) << 30) + int64(len(tail))
	if err := file.Truncate(fileSize); err != nil {
		_ = file.Close()
		t.Skipf("filesystem does not support a sparse test file: %v", err)
	}
	if _, err := file.WriteAt(tail, fileSize-int64(len(tail))); err != nil {
		_ = file.Close()
		t.Fatalf("write sparse file tail: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	handler, err := newFileServerHandler(root)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	response := doRequest(t, server.Client(), http.MethodHead, server.URL+"/huge.bin", nil)
	_ = readResponseBody(t, response)
	if response.ContentLength != fileSize {
		t.Fatalf("large HEAD Content-Length = %d, want %d", response.ContentLength, fileSize)
	}

	start := fileSize - int64(len(tail))
	response = doRequest(t, server.Client(), http.MethodGet, server.URL+"/huge.bin", map[string]string{
		"Range": fmt.Sprintf("bytes=%d-", start),
	})
	body := readResponseBody(t, response)
	if response.StatusCode != http.StatusPartialContent {
		t.Fatalf("large range status = %d, want %d", response.StatusCode, http.StatusPartialContent)
	}
	if got, want := response.Header.Get("Content-Range"), fmt.Sprintf("bytes %d-%d/%d", start, fileSize-1, fileSize); got != want {
		t.Errorf("large Content-Range = %q, want %q", got, want)
	}
	if !bytes.Equal(body, tail) {
		t.Errorf("large range body = %q, want %q", body, tail)
	}
}

func TestDirectoryLinksRedirectsAndPathSafety(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	fileName := "report & # 你好.txt"
	fileContent := []byte("unicode path works")
	if err := os.WriteFile(filepath.Join(root, fileName), fileContent, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("outside root"), 0o644); err != nil {
		t.Fatal(err)
	}

	handler, err := newFileServerHandler(root)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	response := doRequest(t, server.Client(), http.MethodGet, server.URL+"/", nil)
	body := readResponseBody(t, response)
	directoryBodyLength := len(body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("directory status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if !strings.Contains(string(body), "report &amp; # 你好.txt") {
		t.Errorf("directory did not HTML-escape the filename: %s", body)
	}
	if !strings.Contains(string(body), "%23") {
		t.Errorf("directory link did not URL-escape # in filename: %s", body)
	}

	response = doRequest(t, server.Client(), http.MethodGet, server.URL+"/"+url.PathEscape(fileName), nil)
	body = readResponseBody(t, response)
	if response.StatusCode != http.StatusOK || !bytes.Equal(body, fileContent) {
		t.Fatalf("escaped filename response = (%d, %q), want (200, %q)", response.StatusCode, body, fileContent)
	}

	noRedirectClient := *server.Client()
	noRedirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response = doRequest(t, &noRedirectClient, http.MethodGet, server.URL+"/subdir", nil)
	_ = readResponseBody(t, response)
	if response.StatusCode != http.StatusMovedPermanently {
		t.Errorf("directory redirect status = %d, want %d", response.StatusCode, http.StatusMovedPermanently)
	}
	if got := response.Header.Get("Location"); got != "/subdir/" {
		t.Errorf("directory redirect Location = %q, want /subdir/", got)
	}

	request := httptest.NewRequest(http.MethodGet, "http://miniserver.test/%2e%2e/secret.txt", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("traversal status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if strings.Contains(recorder.Body.String(), "outside root") {
		t.Fatal("path traversal exposed a file outside the configured root")
	}

	response = doRequest(t, server.Client(), http.MethodHead, server.URL+"/", nil)
	headBody := readResponseBody(t, response)
	if len(headBody) != 0 {
		t.Errorf("directory HEAD returned %d body bytes, want 0", len(headBody))
	}
	if response.ContentLength != int64(directoryBodyLength) {
		t.Errorf("directory HEAD Content-Length = %d, want %d", response.ContentLength, directoryBodyLength)
	}
}

func TestEmbeddedResponsiveDirectoryFrontend(t *testing.T) {
	server, _ := newTestFileServer(t, map[string][]byte{
		"notes.txt":      []byte("notes"),
		"preview.webp":   []byte("image"),
		"source/main.go": []byte("package main"),
	})

	if len(webassets.DirectoryTemplate) == 0 || len(webassets.DirectoryStyles) == 0 || len(webassets.DirectoryScript) == 0 {
		t.Fatal("one or more embedded web resources are empty")
	}

	response := doRequest(t, server.Client(), http.MethodGet, server.URL+"/", nil)
	body := readResponseBody(t, response)
	page := string(body)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("directory status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'none'") {
		t.Errorf("Content-Security-Policy = %q, want self-contained page policy", got)
	}
	if got := response.Header.Get("Content-Security-Policy"); !strings.Contains(got, "connect-src 'self'") {
		t.Errorf("Content-Security-Policy = %q, want same-origin AJAX permission", got)
	}

	wantFragments := []string{
		"v3.0.0.beta",
		`class="page-shell"`,
		`class="directory-header"`,
		`id="file-search"`,
		`id="theme-toggle"`,
		`id="preview-dialog"`,
		`id="preview-copy"`,
		`border-radius: 24px`,
		`@media (max-width: 640px)`,
		`data-type="TXT"`,
		`data-previewable="true"`,
		`data-previewable="false"`,
		`file-mark-image`,
		`function filterRows()`,
		`var previewLineLimit = 2000`,
		`跳转链接`,
		`直接下载`,
	}
	for _, fragment := range wantFragments {
		if !strings.Contains(page, fragment) {
			t.Errorf("rendered page does not contain embedded frontend fragment %q", fragment)
		}
	}
	if strings.Contains(page, "<script src=") || strings.Contains(page, "<link rel=\"stylesheet\"") {
		t.Fatal("rendered page unexpectedly depends on an external script or stylesheet")
	}
	if strings.Contains(page, `class="hero"`) {
		t.Fatal("rendered page still contains the introductory hero section")
	}
	if strings.Contains(page, "@keyframes") || strings.Contains(page, "animation:") {
		t.Fatal("rendered page still contains loading animations")
	}
}

func TestTextPreviewClassification(t *testing.T) {
	tests := []struct {
		name  string
		isDir bool
		want  bool
	}{
		{name: "notes.txt", want: true},
		{name: "README.md", want: true},
		{name: "main.go", want: true},
		{name: ".env.local", want: true},
		{name: "Dockerfile", want: true},
		{name: "LICENSE", want: true},
		{name: "events.jsonl", want: true},
		{name: "preview.webp", want: false},
		{name: "manual.pdf", want: false},
		{name: "archive.zip", want: false},
		{name: "notes.txt", isDir: true, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isTextPreviewable(test.name, test.isDir); got != test.want {
				t.Errorf("isTextPreviewable(%q, %t) = %t, want %t", test.name, test.isDir, got, test.want)
			}
		})
	}
}

func TestHTTPServerTimeoutsPreserveLongDownloads(t *testing.T) {
	server := newHTTPServer(":0", http.NotFoundHandler())
	if server.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %s, want %s", server.ReadHeaderTimeout, readHeaderTimeout)
	}
	if server.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %s, want %s", server.IdleTimeout, idleTimeout)
	}
	if server.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %s, want no active-transfer deadline", server.WriteTimeout)
	}
	if server.MaxHeaderBytes != maxHeaderBytes {
		t.Errorf("MaxHeaderBytes = %d, want %d", server.MaxHeaderBytes, maxHeaderBytes)
	}
}

func newTestFileServer(t *testing.T, files map[string][]byte) (*httptest.Server, string) {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		filePath := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	handler, err := newFileServerHandler(root)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server, root
}

func doRequest(t *testing.T, client *http.Client, method, requestURL string, headers map[string]string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func readResponseBody(t *testing.T, response *http.Response) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
