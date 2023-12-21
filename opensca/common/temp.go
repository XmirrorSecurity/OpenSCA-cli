package common

import (
	"os"
	"path/filepath"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

var tempdir = ".temp"

func init() {
	// 在opensca-cli所在目录下创建临时目录
	excpath, _ := os.Executable()
	tempdir = filepath.Join(filepath.Dir(excpath), tempdir)
	os.MkdirAll(tempdir, 0755)
	// 删除24小时前的临时文件(一般是进程意外中断时未被删除的临时文件)
	old := time.Now().Add(-24 * time.Hour)
	entries, err := os.ReadDir(tempdir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(old) {
			os.RemoveAll(filepath.Join(tempdir, entry.Name()))
		}
	}
}

func MkdirTemp(pattern string) string {
	tmp, err := os.MkdirTemp(tempdir, pattern)
	if err != nil {
		logs.Warn(err)
	}
	return tmp
}

func CreateTemp(pattern string) *os.File {
	tempf, err := os.CreateTemp(tempdir, pattern)
	if err != nil {
		logs.Warn(err)
	}
	return tempf
}
