package service

import (
	"bytes"
	"html/template"

	"github.com/Ericwyn/MiniServer/conf"
	webassets "github.com/Ericwyn/MiniServer/web"
)

type fileMsgVO struct {
	FileName    string
	FileSize    string
	Href        string
	Category    string
	TypeLabel   string
	LastMod     string
	ModISO      string
	RawSize     int64
	ModUnix     int64
	IsDir       bool
	Previewable bool
}

type breadcrumbVO struct {
	Name    string
	Href    string
	Current bool
}

type directoryPage struct {
	Version     string
	CurrentPath string
	TotalSize   string
	Files       []fileMsgVO
	Breadcrumbs []breadcrumbVO
	ItemCount   int
	FileCount   int
	DirCount    int
	HasParent   bool
	Styles      template.CSS
	Script      template.JS
}

var directoryTemplate = template.Must(
	template.New("directory").Parse(string(webassets.DirectoryTemplate)),
)

func renderHtml(page directoryPage) (string, error) {
	// These are trusted, compile-time resources. Dynamic file data continues to
	// be escaped by html/template before it is inserted into the page.
	page.Version = conf.VersionNum
	page.Styles = template.CSS(string(webassets.DirectoryStyles))
	page.Script = template.JS(string(webassets.DirectoryScript))

	var output bytes.Buffer
	if err := directoryTemplate.Execute(&output, page); err != nil {
		return "", err
	}
	return output.String(), nil
}
