package common

import (
	"os"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

var tempdir = ".temp"

func init() {
	excpath, _ := os.Executable()
	tempdir = filepath.Join(filepath.Dir(excpath), tempdir)
	// os.RemoveAll(tempdir)
	os.MkdirAll(tempdir, 0755)
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
