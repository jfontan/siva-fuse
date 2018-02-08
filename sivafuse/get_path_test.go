package sivafuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type gitExample struct {
	Example  string
	Ok       bool
	PathType string
	Ref      string
	Path     string
}

var gitExamples = []gitExample{
	{
		Example:  "/this/is/a/path",
		Ok:       false,
		PathType: "",
		Ref:      "",
		Path:     "",
	},
	{"/_commit_/76683487299ab8", true, "commit", "76683487299ab8", ""},
	{"/_tag_/v0.6.5/some/path", true, "tag", "v0.6.5", "some/path"},
	{"/_branch_/master/other/path", true, "branch", "master", "other/path"},
	{"/_branch/master/other/path", false, "", "", ""},
}

func TestGetGitPath(t *testing.T) {
	for _, e := range gitExamples {
		ok, pathType, ref, path := getGitPath(e.Example)

		got := gitExample{
			Example:  e.Example,
			Ok:       ok,
			PathType: pathType,
			Ref:      ref,
			Path:     path,
		}

		assert.Equal(t, e, got)
	}
}
