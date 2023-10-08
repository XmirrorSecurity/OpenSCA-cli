package model

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type File struct {
	abspath string
	relpath string
}

func NewFile(abs, rel string) *File {
	return &File{
		abspath: abs,
		relpath: rel,
	}
}

func (file *File) Abspath() string {
	if file != nil {
		return file.abspath
	}
	return ""
}

func (file *File) Relpath() string {
	if file != nil {
		return file.relpath
	}
	return ""
}

func (file *File) OpenReader(do func(reader io.Reader)) error {
	if file == nil || file.abspath == "" {
		return nil
	}
	f, err := os.Open(file.abspath)
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

type FileCommentType struct {
	Simple string
	Begin  string
	End    string
}

var (
	CTypeComment = &FileCommentType{
		Simple: "//",
		Begin:  "/*",
		End:    "*/",
	}
	PythonTypeComment = &FileCommentType{
		Simple: "#",
		Begin:  "'''",
		End:    "'''",
	}
)

func (file File) ReadLineNoComment(t *FileCommentType, do func(line string)) {

	if t == nil {
		t = CTypeComment
	}

	comment := false

	file.ReadLine(func(line string) {

		if t.Simple != "" && strings.HasPrefix(strings.TrimSpace(line), t.Simple) {
			return
		}

		if t.Begin != "" && t.End != "" {
			i := strings.Index(line, t.Begin)
			if i != -1 {
				comment = true
				do(line[:i])
				return
			}
			i = strings.Index(line, t.End)
			if comment && i != -1 {
				comment = false
				do(line[i+len(t.End):])
				return
			}
			if comment {
				return
			}
		}

		do(line)
	})
}

type ResCallback func(file *File, root ...*DepGraph)
