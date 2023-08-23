package model

import (
	"bufio"
	"os"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

type File struct {
	Abspath string
	Relpath string
}

func (file *File) ReadLine(do func(line string)) {
	f, err := os.Open(file.Abspath)
	if err != nil {
		logs.Warnf("open file %s fail: %s", file.Relpath, err)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		do(scanner.Text())
	}
}
