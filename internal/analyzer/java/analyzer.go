/*
 * @Descripation: java Analyzer
 * @Date: 2021-11-03 11:17:09
 */
package java

import (
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/srt"
	"path"
	"regexp"
	"strings"
)

type Analyzer struct {
	mvn *Mvn
	// maven仓库地址
	repos map[int64][]string
}

// New 创建java解析器
func New() Analyzer {
	return Analyzer{
		mvn:   NewMvn(),
		repos: map[int64][]string{},
	}
}

// GetLanguage Get language of Analyzer
func (Analyzer) GetLanguage() language.Type {
	return language.Java
}

// CheckFile Check if it is a parsable file
func (Analyzer) CheckFile(filename string) bool {
	return filter.JavaPom(filename)
}

// FilterFile filters the files that the current parser needs to parse
func (a Analyzer) FilterFile(dirRoot *srt.DirTree, depRoot *srt.DepTree) (files []*srt.FileData) {
	// 通过jar包预解析组件名
	if filter.Jar(dirRoot.Path) {
		fileName := strings.TrimSuffix(path.Base(dirRoot.Path), path.Ext(dirRoot.Path))
		depRoot.Language = language.Java
		re := regexp.MustCompile(`-(\d+(\.[\d\w]+)*)`)
		// 未获取到组件信息或获取到多个同名组件时解析jarname
		if re.MatchString(fileName) {
			index := re.FindStringIndex(fileName)[0]
			depRoot.Version = srt.NewVersion(fileName[index+1:])
			depRoot.Name = fileName[:index]
		}
		depRoot.Path = dirRoot.Path
	}
	// 筛选需要解析的文件
	files = []*srt.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	return files
}

// ParseFile Parse the file
func (a Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	if filter.JavaPom(file.Name) {
		p := ReadPom(file.Data)
		p.Path = path.Join(dirRoot.Path, file.Name)
		a.mvn.AppendPom(p)
	}
	return make([]*srt.DepTree, 0)
}
