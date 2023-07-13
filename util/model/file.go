/*
 * @Description: 文件相关数据结构
 * @Date: 2021-11-03 11:24:58
 */
package model

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

// FileInfo 文件数据
type FileInfo struct {
	Name string `json:"name"`
	Data []byte `json:"-"`
}

// NewFileData 创建FileData
func NewFileData(name string, data []byte) *FileInfo {
	// 统一替换换行符为\n
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\n\r"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	return &FileInfo{
		Name: strings.ReplaceAll(name, `\`, `/`),
		Data: data,
	}
}

// DirTree 目录树
type DirTree struct {
	// 子目录
	SubDir map[string]*DirTree `json:"subdir"`
	// 目录列表
	DirList []string `json:"-"`
	// 当前目录下文件
	Files []*FileInfo `json:"files"`
	// 路径
	Path string `json:"path"`
}

// NewDirTree 创建空目录树
func NewDirTree() *DirTree {
	return &DirTree{SubDir: map[string]*DirTree{}, DirList: []string{}, Files: make([]*FileInfo, 0)}
}

// AddFile 向目录树添加一个文件
func (root *DirTree) AddFile(file *FileInfo) {
	now := root.GetDir(file.Name)
	now.Files = append(now.Files, file)
}

// GetDir 获取文件所在目录
func (root *DirTree) GetDir(filepath string) *DirTree {
	// 格式化路径
	filepath = strings.ReplaceAll(filepath, `\`, `/`)
	paths := strings.Split(filepath, "/")
	now := root
	for _, dirName := range paths[:len(paths)-1] {
		if _, ok := now.SubDir[dirName]; !ok {
			now.SubDir[dirName] = &DirTree{SubDir: map[string]*DirTree{}, DirList: []string{}, Files: make([]*FileInfo, 0)}
			now.DirList = append(now.DirList, dirName)
		}
		now = now.SubDir[dirName]
	}
	return now
}

// BuildDirPath 构建目录路径
func (root *DirTree) BuildDirPath() {
	queue := NewQueue()
	queue.Push(root)
	for !queue.Empty() {
		node := queue.Pop().(*DirTree)
		for _, dirName := range node.DirList {
			sub := node.SubDir[dirName]
			sub.Path = path.Join(node.Path, dirName)
			if len(sub.DirList) > 0 {
				queue.Push(sub)
			}
		}
	}
}

// String 目录树结构
func (root *DirTree) String() string {
	type node struct {
		Dir  *DirTree
		Deep int
	}
	newNode := func(dir *DirTree, deep int) *node {
		return &node{
			Dir:  dir,
			Deep: deep,
		}
	}
	res := ""
	stack := NewStack()
	stack.Push(newNode(root, 0))
	for !stack.Empty() {
		node := stack.Pop().(*node)
		res += fmt.Sprintf("%s%s\n", strings.Repeat("\t", node.Deep), node.Dir.Path)
		for i := len(node.Dir.DirList) - 1; i >= 0; i-- {
			dir := node.Dir.DirList[i]
			stack.Push(newNode(node.Dir.SubDir[dir], node.Deep+1))
		}
	}
	return res
}
