package model

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// File 文件相关信息
type File struct {
	abspath string
	relpath string
}

// NewFile 创建文件对象
// abs: 文件绝对路径
// rel: 文件相对路径(相对于项目根目录)
func NewFile(abs, rel string) *File {
	return &File{
		abspath: abs,
		relpath: rel,
	}
}

// Abspath 文件绝对路径
func (file *File) Abspath() string {
	if file != nil {
		return file.abspath
	}
	return ""
}

// Relpath 文件相对路径
func (file *File) Relpath() string {
	if file != nil {
		return file.relpath
	}
	return ""
}

func (file *File) String() string {
	return file.Relpath()
}

// OpenReader 打开文件reader
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

// ReadLine 按行读取文件内容 去除行尾换行符
func (file File) ReadLine(do func(line string)) {
	file.OpenReader(func(reader io.Reader) {
		ReadLine(reader, do)
	})
}

// ReadLineNoComment 按行读取文件内容 忽略注释
func (file File) ReadLineNoComment(t *CommentType, do func(line string)) {
	file.OpenReader(func(reader io.Reader) {
		ReadLineNoComment(reader, t, do)
	})
}

// 注释类型
type CommentType struct {
	// 单行注释标记
	Simple string
	// 多行注释起始标记
	Begin string
	// 多行注释终止标记
	End string
}

var (
	// C语言注释类型
	CTypeComment = &CommentType{
		Simple: "//",
		Begin:  "/*",
		End:    "*/",
	}
	// Python语言注释类型
	PythonTypeComment = &CommentType{
		Simple: "#",
		Begin:  "'''",
		End:    "'''",
	}
)

// ReadLine 按行读取内容 去除行尾换行符
func ReadLine(reader io.Reader, do func(line string)) {

	if do == nil {
		return
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		do(strings.TrimRight(scanner.Text(), "\n\r"))
	}
}

// ReadLineNoComment 按行读取内容 忽略注释
func ReadLineNoComment(reader io.Reader, t *CommentType, do func(line string)) {

	if do == nil {
		return
	}

	if t == nil {
		t = CTypeComment
	}

	// 标记当前是非位于多行注释段
	comment := false

	ReadLine(reader, func(line string) {

		// 单行注释
		if t.Simple != "" {
			i := strings.Index(line, t.Simple)
			if i != -1 {
				line = line[:i]
			}
		}

		// 多行注释
		if t.Begin != "" && t.End != "" {
			for {
				// 当前非注释段且存在注释起始标记
				if start_i := strings.Index(line, t.Begin); !comment && start_i != -1 {
					comment = true
					do(line[:start_i])
					line = line[start_i+len(t.Begin):]
					continue
				}
				// 当前为注释段且存在注释终止标记
				if end_i := strings.Index(line, t.End); comment && end_i != -1 {
					comment = false
					line = line[end_i+len(t.End):]
					continue
				}
				break
			}
			if comment {
				return
			}
		}

		do(line)
	})

}

// ResCallback 检测结果回调函数
// file: 检出组件的文件信息
// root: 组件依赖图根节点列表
type ResCallback func(file *File, root ...*DepGraph)
