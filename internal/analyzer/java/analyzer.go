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
	"sort"
	"strings"
)

type Analyzer struct {
	// 属性
	properties map[string]map[string]string
	// 记录获取过的文件
	poms map[string]map[string]struct{}
	// maven仓库地址
	repos map[int64][]string
}

/**
 * @description: 创建java解析器
 * @return {java.Analyzer} java解析器
 */
func New() Analyzer {
	return Analyzer{
		properties: map[string]map[string]string{},
		poms:       map[string]map[string]struct{}{},
		repos:      map[int64][]string{},
	}
}

/**
 * @description: 获取当前Analyzer的语言
 * @return {language.Type} 语言
 */
func (Analyzer) GetLanguage() language.Type {
	return language.Java
}

/**
 * @description: 检测是否是可解析的文件
 * @param {string} filename 文件名
 * @return {bool} 是可解析的文件返回true
 */
func (Analyzer) CheckFile(filename string) bool {
	return filter.JavaPom(filename) ||
		filter.JavaPomProperties(filename)
}

/**
 * @descriptsrt筛选当前解析器需要解析的文件
 * @param {*modsrtrTree} dirRoot 目录树节点
 * @param {*srt.DepTree} depRoot 依赖树节点
 * @return {[]*srt.FileData} 需要解析的文件列表
 */
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
	}
	// 筛选需要解析的文件
	files = []*srt.FileData{}
	for _, f := range dirRoot.Files {
		if a.CheckFile(f.Name) {
			files = append(files, f)
		}
	}
	// 文件排序
	sort.Slice(files, func(i, j int) bool {
		// 优先解析pom.properties文件
		return filter.JavaPomProperties(files[i].Name) && !filter.JavaPomProperties(files[j].Name)
	})
	return files
}

/**
 * @descriptsrt解析文件
 * @param {*srt.DirTree} dirRoot 目录树节点
 * @param {*modsrtpTree} depRoot 依赖树节点
 * @param {*srt.FileData} file 文件信息
 * @return {[]*srt.DepTree} 解析出的依赖列表
 */
func (a Analyzer) ParseFile(dirRoot *srt.DirTree, depRoot *srt.DepTree, file *srt.FileData) []*srt.DepTree {
	if filter.JavaPom(file.Name) {
		return a.parsePom(dirRoot, depRoot, file)
	} else if filter.JavaPomProperties(file.Name) {
		a.parsePomProperties(dirRoot.Path, file.Data)
	}
	return make([]*srt.DepTree, 0)
}
