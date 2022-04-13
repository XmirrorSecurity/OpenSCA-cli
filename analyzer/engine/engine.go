/*
 * @Descripation:
 * @Date: 2021-11-03 17:17:16
 */

package engine

import (
	"analyzer/analyzer"
	"analyzer/erlang"
	"analyzer/golang"
	"analyzer/java"
	"analyzer/javascript"
	"analyzer/php"
	"analyzer/ruby"
	"analyzer/rust"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"util/args"
	"util/enum/language"
	"util/filter"
	"util/logs"
	"util/model"
	"util/vuln"
)

type Engine struct {
	javaAnalyzer java.Analyzer
	Analyzers    []analyzer.Analyzer
}

// NewEngine 创建新引擎
func NewEngine() Engine {
	j := java.New()
	return Engine{
		javaAnalyzer: j,
		Analyzers: []analyzer.Analyzer{
			j,
			javascript.New(),
			php.New(),
			ruby.New(),
			golang.New(),
			rust.New(),
			erlang.New(),
		},
	}
}

// ParseFile 解析一个目录或文件
func (e Engine) ParseFile(filepath string) (*model.DepTree, error) {
	// 目录树
	dirRoot := model.NewDirTree()
	depRoot := model.NewDepTree(nil)
	if f, err := os.Stat(filepath); err != nil {
		logs.Error(err)
		return depRoot, err
	} else {
		if f.IsDir() {
			// 目录
			dirRoot = e.opendir(filepath)
			// 尝试解析mvn依赖
			depRoot = java.MvnDepTree(filepath)
		} else if filter.AllPkg(filepath) {
			// 压缩包
			dirRoot = e.unArchiveFile(filepath)
		} else if e.checkFile(filepath) {
			// 单个文件
			if data, err := ioutil.ReadFile(filepath); err != nil {
				logs.Error(err)
			} else {
				dirRoot.Files = append(dirRoot.Files, model.NewFileData(filepath, data))
			}
		}
		dirRoot.Path = path.Base(strings.ReplaceAll(filepath, `\`, `/`))
	}
	dirRoot.BuildDirPath()
	// 解析目录树获取依赖树
	e.parseDependency(dirRoot, depRoot)
	// 同组件去重
	q := model.NewQueue()
	q.Push(depRoot)
	// 用于记录相同组件信息
	depMap := map[language.Type]map[string]*model.DepTree{}
	for !q.Empty() {
		node := q.Pop().(*model.DepTree)
		for _, child := range node.Children {
			q.Push(child)
		}
		// 保留第一个同语言同厂商同名组件的组件
		if _, ok := depMap[node.Language]; !ok {
			depMap[node.Language] = map[string]*model.DepTree{}
		}
		key := strings.ToLower(fmt.Sprintf("%s:%s", node.Vendor, node.Name))
		if dep, ok := depMap[node.Language][key]; ok {
			node.Move(dep)
		} else {
			depMap[node.Language][key] = node
		}
	}
	// 再次排除exclusion组件
	depRoot.Exclusion()
	// 获取漏洞
	err := vuln.SearchVuln(depRoot)
	// 是否仅保留漏洞组件
	if args.OnlyVuln {
		root := model.NewDepTree(nil)
		q := model.NewQueue()
		q.Push(depRoot)
		for !q.Empty() {
			dep := q.Pop().(*model.DepTree)
			for _, child := range dep.Children {
				q.Push(child)
			}
			dep.Children = nil
			dep.Parent = nil
			if len(dep.Vulnerabilities) > 0 {
				dep.Parent = root
				root.Children = append(root.Children, dep)
			}
		}
		depRoot = root
	} else {
		// save indirect vulninfo
		q := model.NewQueue()
		q.Push(depRoot)
		for !q.Empty() {
			dep := q.Pop().(*model.DepTree)
			vulnExist := map[string]struct{}{}
			for _, child := range dep.Children {
				for _, vuln := range child.Vulnerabilities {
					if _, exist := vulnExist[vuln.Id]; !exist {
						vulnExist[vuln.Id] = struct{}{}
					}
				}
				q.Push(child)
			}
			dep.IndirectVulnerabilities = len(vulnExist)
		}
	}
	return depRoot, err
}
