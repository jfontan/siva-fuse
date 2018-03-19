package sivafuse

import (
	"os"
	"path/filepath"
	"strconv"
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

// Types:
//
// * root
// * message
// * parent
// * tree
func getCommitPath(name string) (ok bool, pathType, path string) {
	ok = false
	pathType = ""
	path = ""

	if name != "" && name[0:1] == "/" {
		name = name[1:]
	}

	if name == "" {
		return true, "root", ""
	}

	p := strings.Split(name, "/")
	base := p[0]

	switch base {
	case "message":
		if len(p) != 1 {
			return
		}
		return true, p[0], ""

	case "parent":
		// Can not have subdirectories
		if len(p) > 2 {
			return
		}

		// Files in the subdirectory are parent indexes (unsigned ints)
		if len(p) > 1 {
			_, err := strconv.ParseUint(p[1], 10, 8)
			if err != nil {
				return
			}
			path = p[1]
		}
		return true, p[0], path

	case "tree":
		if len(p) == 1 {
			return true, p[0], ""
		}

		return true, p[0], filepath.Join(p[1:len(p)]...)
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
