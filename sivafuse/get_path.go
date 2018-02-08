package sivafuse

import (
	"path/filepath"
	"strings"
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

	p := strings.Split(name, "/")
	base := p[0]
	baseLen := len(base)

	if base[:1] != "_" && base[baseLen-1:1] != "_" {
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
