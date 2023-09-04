package model

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type File struct {
	Abspath string
	Relpath string
}

func (file *File) Path() string {
	if file != nil {
		return file.Relpath
	}
	return ""
}

func (file *File) OpenReader(do func(reader io.Reader)) error {
	if file == nil || file.Abspath == "" {
		return nil
	}
	f, err := os.Open(file.Abspath)
	if err != nil {
		return err
	}
	defer f.Close()
	do(f)
	return nil
}

func (file File) ReadLine(do func(line string)) {
	file.OpenReader(func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			do(strings.TrimRight(scanner.Text(), "\n\r"))
		}
	})
}
