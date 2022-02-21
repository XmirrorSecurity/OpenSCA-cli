/*
 * @Descripation:
 * @Date: 2021-11-03 17:17:16
 */

package engine

import (
	"fmt"
	"io/ioutil"
	"opensca/internal/analyzer"
	"opensca/internal/analyzer/golang"
	"opensca/internal/analyzer/java"
	"opensca/internal/analyzer/javascript"
	"opensca/internal/analyzer/php"
	"opensca/internal/analyzer/ruby"
	"opensca/internal/args"
	"opensca/internal/enum/language"
	"opensca/internal/filter"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"opensca/internal/vuln"
	"os"
	"path"
	"strings"
)

type Engine struct {
	Analyzers []analyzer.Analyzer
}

/**
 * @description: 创建新引擎
 * @return {engine.Engine} 解析引擎
 */
func NewEngine() Engine {
	return Engine{
		Analyzers: []analyzer.Analyzer{
			java.New(),
			javascript.New(),
			php.New(),
			ruby.New(),
			golang.New(),
		},
	}
}

/**
 * @description: 解析一个目录或文件
 * @param {string} path 目录或文件路径
 */
func (e Engine) ParseFile(filepath string) {
	// 目录树
	dirRoot := srt.NewDirTree()
	depRoot := srt.NewDepTree(nil)
	if f, err := os.Stat(filepath); err != nil {
		logs.Error(err)
		return
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
				dirRoot.Files = append(dirRoot.Files, srt.NewFileData(filepath, data))
			}
		}
		dirRoot.Path = path.Base(strings.ReplaceAll(filepath, `\`, `/`))
	}
	dirRoot.BuildDirPath()
	// 解析目录树获取依赖树
	e.parseDependency(dirRoot, depRoot)
	// 同组件去重
	q := srt.NewQueue()
	q.Push(depRoot)
	// 用于记录相同组件信息
	depMap := map[language.Type]map[string]*srt.DepTree{}
	for !q.Empty() {
		node := q.Pop().(*srt.DepTree)
		for _, child := range node.Children {
			q.Push(child)
		}
		// 保留第一个同语言同厂商同名组件的组件
		if _, ok := depMap[node.Language]; !ok {
			depMap[node.Language] = map[string]*srt.DepTree{}
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
	vulnError := vuln.SearchVuln(depRoot)
	// 是否仅保留漏洞组件
	if args.OnlyVuln {
		root := srt.NewDepTree(nil)
		q := srt.NewQueue()
		q.Push(depRoot)
		for !q.Empty() {
			dep := q.Pop().(*srt.DepTree)
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
		q := srt.NewQueue()
		q.Push(depRoot)
		for !q.Empty() {
			dep := q.Pop().(*srt.DepTree)
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
	// 整理错误信息
	errInfo := ""
	if vulnError != nil {
		errInfo = vulnError.Error()
	}
	// 记录依赖
	logs.Debug("\n" + depRoot.String())
	// 输出结果
	if args.Out != "" {
		// 保存到json
		if f, err := os.Create(args.Out); err != nil {
			logs.Error(err)
		} else {
			defer f.Close()
			if size, err := f.Write(depRoot.Json(errInfo)); err != nil {
				logs.Error(err)
			} else {
				logs.Info(fmt.Sprintf("size: %d, output: %s", size, args.Out))
			}
		}
	} else {
		fmt.Println(string(depRoot.Json(errInfo)))
	}
}
