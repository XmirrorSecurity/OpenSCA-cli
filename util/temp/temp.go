package temp

import (
	"os"
	"path"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

var tempdir string

func init() {
	p, _ := os.Executable()
	tempdir = path.Join(filepath.Dir(p), ".opensca-tempdir")
	os.RemoveAll(tempdir)
	os.Mkdir(tempdir, os.ModePerm)
}

// DoInTempDir 在临时目录中执行
func DoInTempDir(do func(tempdir string)) {
	dir, err := os.MkdirTemp(tempdir, "")
	if err != nil {
		logs.Warn(err)
	}
	defer os.RemoveAll(dir)
	do(dir)
}
