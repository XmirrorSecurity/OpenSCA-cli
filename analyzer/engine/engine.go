/*
 * @Descripation:
 * @Date: 2021-11-03 17:17:16
 */

package engine

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/util/args"
	"github.com/xmirrorsecurity/opensca-cli/util/filter"
	"github.com/xmirrorsecurity/opensca-cli/util/logs"
	"github.com/xmirrorsecurity/opensca-cli/util/model"
	"github.com/xmirrorsecurity/opensca-cli/util/report"
	"github.com/xmirrorsecurity/opensca-cli/util/vuln"

	"github.com/xmirrorsecurity/opensca-cli/analyzer/analyzer"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/erlang"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/golang"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/java"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/javascript"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/php"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/python"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/ruby"
	"github.com/xmirrorsecurity/opensca-cli/analyzer/rust"
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
			// 暂不解析groovy文件
			// groovy.New(),
			python.New(),
		},
	}
}

// ParseFile 解析一个目录或文件
func (e Engine) ParseFile(filepath string) (depRoot *model.DepTree, taskInfo report.TaskInfo) {
	// 目录树
	dirRoot := model.NewDirTree()
	depRoot = model.NewDepTree(nil)
	filepath = strings.ReplaceAll(filepath, `\`, `/`)
	taskInfo = report.TaskInfo{
		AppName:   strings.TrimSuffix(path.Base(filepath), path.Ext(path.Base(filepath))),
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	s := time.Now()
	defer func() {
		taskInfo.CostTime = time.Since(s).Seconds()
		taskInfo.EndTime = time.Now().Format("2006-01-02 15:04:05")
	}()
	if f, err := os.Stat(filepath); err != nil {
		taskInfo.Error = err
		logs.Error(err)
		fmt.Println(err)
		return depRoot, taskInfo
	} else {
		if f.IsDir() {
			// 目录
			dirRoot = e.opendir(filepath)
			// 尝试解析mvn依赖
			java.MvnDepTree(filepath, depRoot)
			// 尝试解析gradle依赖
			java.GradleDepTree(filepath, depRoot)
		} else if filter.AllPkg(filepath) {
			if f, err := os.Stat(filepath); err != nil {
				logs.Warn(err)
			} else {
				taskInfo.Size = f.Size()
			}
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
	taskInfo.Error = vuln.SearchDetail(depRoot)
	// 是否仅保留漏洞组件
	if args.Config.OnlyVuln {
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
	return depRoot, taskInfo
}
