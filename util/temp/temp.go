package temp

import (
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"util/logs"
)

const tempdir = ".temp"

func init() {
	os.RemoveAll(path.Join(GetPwd(), tempdir))
}

// GetPwd 获取当前目录
func GetPwd() string {
	filepath, err := os.Executable()
	if err != nil {
		logs.Error(err)
		return ""
	}
	return path.Dir(strings.ReplaceAll(filepath, `\`, `/`))
}

// DoInTempDir 在临时目录中执行
func DoInTempDir(do func(tempdir string)) {
	dir := path.Join(GetPwd(), tempdir, strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logs.Warn(err)
	} else {
		defer os.RemoveAll(dir)
		do(dir)
	}
}
