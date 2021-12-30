/*
 * @Descripation: 文件相关数据结构
 * @Date: 2021-11-03 11:24:58
 */
package srt

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

/**
 * @description: 文件数据
 */
type FileData struct {
	Name string `json:"name"`
	Data []byte `json:"-"`
}

/**
 * @description: 创建FileData
 * @param {string} name 文件名
 * @param {[]byte} data 文件内容
 * @return {*FileData} FileData结构
 */
func NewFileData(name string, data []byte) *FileData {
	// 统一替换换行符为\n
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\n\r"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	return &FileData{
		Name: strings.ReplaceAll(name, `\`, `/`),
		Data: data,
	}
}

/**
 * @description: 目录树
 */
type DirTree struct {
	// 子目录
	SubDir map[string]*DirTree `json:"subdir"`
	// 目录列表
	DirList []string `json:"-"`
	// 当前目录下文件
	Files []*FileData `json:"files"`
	// 路径
	Path string `json:"path"`
}

/**
 * @description: 创建空目录树
 * @return {*DirTree} 空目录树
 */
func NewDirTree() *DirTree {
	return &DirTree{SubDir: map[string]*DirTree{}, DirList: []string{}, Files: make([]*FileData, 0)}
}

/**
 * @description: 向目录树添加一个文件
 * @param {*FileData} file 文件内容
 */
func (root *DirTree) AddFile(file *FileData) {
	now := root.GetDir(file.Name)
	now.Files = append(now.Files, file)
}

/**
 * @description: 获取文件所在目录
 * @param {string} filepath 文件目录
 * @return {*DirTree} 文件所在目录
 */
func (root *DirTree) GetDir(filepath string) *DirTree {
	// 格式化路径
	filepath = strings.ReplaceAll(filepath, `\`, `/`)
	paths := strings.Split(filepath, "/")
	now := root
	for _, dirName := range paths[:len(paths)-1] {
		if _, ok := now.SubDir[dirName]; !ok {
			now.SubDir[dirName] = &DirTree{SubDir: map[string]*DirTree{}, DirList: []string{}, Files: make([]*FileData, 0)}
			now.DirList = append(now.DirList, dirName)
		}
		now = now.SubDir[dirName]
	}
	return now
}

/**
 * @description: 构建目录路径
 */
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

/**
 * @description: 目录树结构
 * @return {string} 目录树字符串
 */
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
