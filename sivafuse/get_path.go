package sivafuse

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getSivaPath(name string) (ok bool, fsPath, sivaPath string) {
	p := strings.Split(name, "/")

	for i, s := range p {
		if len(s) > 5 && s[len(s)-5:] == ".siva" {
			return true, filepath.Join(p[:i+1]...), filepath.Join(p[i+1:]...)
		}
	}

	return false, name, ""
}

var pathTypes = []string{
	"branch",
	"tag",
	"commit",
}

func getGitPath(name string) (ok bool, pathType, ref, path string) {
	ok = false
	pathType = ""
	ref = ""
	path = ""

	if name != "" && name[0:1] == "/" {
		name = name[1:]
	}

	p := strings.Split(name, "/")
	base := p[0]
	baseLen := len(base)

	if baseLen < 2 || (base[:1] != "_" && base[baseLen-1:] != "_") {
		return
	}

	for _, t := range pathTypes {
		// TODO: these strings and lengths should be pregenerated
		str := "_" + t + "_"
		strLen := len(str)

		if baseLen == strLen && base == str {
			ok = true
			pathType = t

			if len(p) > 1 {
				ref = p[1]
				if len(p) > 2 {
					path = strings.Join(p[2:], "/")
				}
			}

			return
		}
	}

	return
}

type fileInfo struct {
	name string
	size int64
	dir  bool
}

// NewFileInfo creates a new fileInfo object
func NewFileInfo(name string, size int64, dir bool) os.FileInfo {
	return &fileInfo{
		name: name,
		size: size,
		dir:  dir,
	}
}

func (p *fileInfo) Name() string {
	return p.name
}

func (p *fileInfo) IsDir() bool {
	return p.dir
}

func (p *fileInfo) ModTime() time.Time {
	return time.Now()
}

func (p *fileInfo) Mode() os.FileMode {
	return 0500
}

func (p *fileInfo) Sys() interface{} {
	return nil
}

func (p fileInfo) Size() int64 {
	return p.size
}

func getTypeFileInfo(name string) os.FileInfo {
	return NewFileInfo("_"+name+"_", 0, true)
}

func getPathTypesFileInfo() []os.FileInfo {
	info := make([]os.FileInfo, len(pathTypes))

	for i, dir := range pathTypes {
		info[i] = getTypeFileInfo(dir)
	}

	return info
}
