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
		ReadLine(reader, do)
	})
}

func (file File) ReadLineNoComment(t *CommentType, do func(line string)) {
	file.OpenReader(func(reader io.Reader) {
		ReadLineNoComment(reader, t, do)
	})
}

type CommentType struct {
	Simple string
	Begin  string
	End    string
}

var (
	CTypeComment = &CommentType{
		Simple: "//",
		Begin:  "/*",
		End:    "*/",
	}
	PythonTypeComment = &CommentType{
		Simple: "#",
		Begin:  "'''",
		End:    "'''",
	}
)

func ReadLine(reader io.Reader, do func(line string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		do(strings.TrimRight(scanner.Text(), "\n\r"))
	}
}

func ReadLineNoComment(reader io.Reader, t *CommentType, do func(line string)) {

	if t == nil {
		t = CTypeComment
	}

	comment := false

	ReadLine(reader, func(line string) {

		// 单行注释
		if t.Simple != "" {
			i := strings.Index(line, t.Simple)
			if i != -1 {
				line = line[:i]
			}
			if strings.TrimSpace(line) == "" {
				return
			}
		}

		// 多行注释
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
