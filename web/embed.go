package web

import _ "embed"

// DirectoryTemplate is the directory page template compiled into the binary.
//
//go:embed directory.html
var DirectoryTemplate []byte

// DirectoryStyles contains the page styles compiled into the binary.
//
//go:embed directory.css
var DirectoryStyles []byte

// DirectoryScript contains the page behavior compiled into the binary.
//
//go:embed directory.js
var DirectoryScript []byte
