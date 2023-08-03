package temp

import (
	"os"
	"path"
	"util/logs"
)

const tempdir = ".opensca-tempdir"

func init() {
	pwd, _ := os.Getwd()
	os.RemoveAll(path.Join(pwd, tempdir))
}

// DoInTempDir 在临时目录中执行
func DoInTempDir(do func(tempdir string)) {
	pwd, err := os.Getwd()
	if err != nil {
		logs.Warn(err)
	}

	dir, err := os.MkdirTemp(path.Join(pwd, tempdir), "")
	if err != nil {
		logs.Warn(err)
	}

	defer os.RemoveAll(dir)
	do(dir)
}
