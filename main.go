package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
	"github.com/xmirrorsecurity/opensca-cli/opensca"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/php"
)

var version string

func main() {

	// 处理参数
	args()

	path := config.Conf().Path

	// 记录开始时间
	start := time.Now()
	report := format.Report{
		ToolVersion: version,
		AppName:     path,
		StartTime:   start.Format("2006-01-02 15:04:05"),
	}

	// 记录检测目标文件大小
	if f, err := os.Stat(path); err == nil {
		report.Size = f.Size()
	}

	// 检测参数
	taskArg := &opensca.TaskArg{DataOrigin: path}

	// 是否跳过压缩包检测
	if config.Conf().DirOnly {
		taskArg.ExtractFileFilter = func(relpath string) bool { return false }
	}

	// 进度条
	if config.Conf().ProgressBar {
		var find, done, bar int
		go func() {
			logos := []string{`-`, `\`, `|`, `/`}
			for {
				fmt.Printf("\r%s find:%d done:%d", logos[bar], find, done)
				bar = (bar + 1) % len(logos)
				<-time.After(time.Millisecond * 100)
			}
		}()
		taskArg.WalkFileFunc = func(parent *model.File, files []*model.File) {
			for range files {
				find++
			}
		}
		taskArg.DeferWalkFileFunc = func(parent *model.File, files []*model.File) {
			for range files {
				done++
			}
		}
	}

	// 运行检测任务
	deps, err := opensca.RunTask(context.Background(), taskArg)
	if err != nil {
		report.ErrorString = err.Error()
	}

	// 合并检测结果
	root := &model.DepGraph{}
	if len(deps) > 1 {
		for _, dep := range deps {
			root.AppendChild(dep)
		}
	} else if len(deps) == 1 {
		root = deps[0]
	}
	logs.Debugf("dependencies tree:\n%s", root.Tree(false, false))

	report.DepDetailGraph = detail.NewDepDetailGraph(root)

	// 组件去重
	if config.Conf().Dedup {
		report.Dedup()
	}

	// 查询组件详情(漏洞/许可证)
	err = detail.SearchDetail(report.DepDetailGraph)
	if err != nil {
		if report.ErrorString != "" {
			report.ErrorString += "\n"
		}
		report.ErrorString += err.Error()
	}

	// 仅保留漏洞组件
	if config.Conf().VulnOnly {
		var deps []*detail.DepDetailGraph
		report.ForEach(func(n *detail.DepDetailGraph) bool {
			if len(n.Vulnerabilities) > 0 {
				deps = append(deps, n)
			}
			return true
		})
		report.DepDetailGraph = &detail.DepDetailGraph{Children: deps}
	}

	// 记录检测时长
	report.CostTime = time.Since(start).Seconds()
	report.EndTime = time.Now().Format("2006-01-02 15:04:05")

	// 打印概览信息
	fmt.Println(format.Statis(report))

	// 导出报告
	format.Save(report, config.Conf().Output)

}

func args() {

	config.ParseArgs()

	if config.Conf().Version {
		fmt.Println(version)
		os.Exit(0)
	}

	logs.CreateLog(config.Conf().LogFile)

	path := config.Conf().Path
	if len(path) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	java.RegisterMavenRepo(config.Conf().Maven...)
	php.RegisterComposerRepo(config.Conf().Composer...)

}
