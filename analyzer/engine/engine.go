/*
 * @Descripation:
 * @Date: 2021-11-03 17:17:16
 */

package engine

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"util/args"
	"util/filter"
	"util/logs"
	"util/model"
	"util/vuln"

	"analyzer/analyzer"
	"analyzer/erlang"
	"analyzer/golang"
	"analyzer/java"
	"analyzer/javascript"
	"analyzer/php"
	"analyzer/ruby"
	"analyzer/rust"
)

type Engine struct {
	Analyzers []analyzer.Analyzer
}

// NewEngine 创建新引擎
func NewEngine() Engine {
	return Engine{
		Analyzers: []analyzer.Analyzer{
			java.New(),
			javascript.New(),
			php.New(),
			ruby.New(),
			rust.New(),
			golang.New(),
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
