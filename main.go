package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/cmd/format"
	"github.com/xmirrorsecurity/opensca-cli/cmd/ui"
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

	// 检测参数
	arg := &opensca.TaskArg{DataOrigin: config.Conf().Path}

	// 是否跳过压缩包检测
	if config.Conf().Optional.DirOnly {
		arg.ExtractFileFilter = func(relpath string) bool { return false }
	}

	// 开启进度条
	var stopProgress func()
	if config.Conf().Optional.ProgressBar {
		stopProgress = startProgressBar(arg)
	}

	// 运行检测任务
	result := opensca.RunTask(context.Background(), arg)

	// 日志中记录检测结果
	if result.Error != nil {
		logs.Error(result.Error)
	}
	for _, dep := range result.Deps {
		logs.Debugf("dependency tree:\n%s", dep.Tree(false, false))
	}

	// 生成报告
	report := taskReport(result)

	// 导出报告
	format.Save(report, config.Conf().Output)

	// 等待进度条完成
	if config.Conf().Optional.ProgressBar {
		<-time.After(time.Millisecond * 200)
		if stopProgress != nil {
			stopProgress()
		}
	}

	// 打印概览信息
	fmt.Println("\n\nComplete!\n" + format.Statis(report))

	// 开启ui
	if config.Conf().Optional.UI {
		ui.OpenUI(report)
	}

}

func args() {

	v := false
	var cfgf string
	cfg := config.Conf()
	flag.BoolVar(&v, "version", false, "-version")
	flag.StringVar(&cfgf, "config", "", "config path. example: -config config.json")
	flag.StringVar(&cfg.Path, "path", cfg.Path, "project path. example: -path project_path")
	flag.StringVar(&cfg.Output, "out", cfg.Output, "report path, support html/json/xml/csv/sqlite/cdx/spdx/swid/dsdx. example: -out out.json,out.html")
	flag.StringVar(&cfg.LogFile, "log", cfg.LogFile, "-log ./my_opensca_log.txt")
	flag.StringVar(&cfg.Origin.Token, "token", "", "web token, example: -token xxxx")
	flag.Parse()

	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	config.LoadConfig(cfgf)
	flag.Parse()

	logs.CreateLog(config.Conf().LogFile)

	java.RegisterMavenRepo(config.Conf().Repo.Maven...)
	php.RegisterComposerRepo(config.Conf().Repo.Composer...)
}

func startProgressBar(arg *opensca.TaskArg) (stop func()) {

	progress := true

	var find, deps, bar int

	go func() {
		logos := []string{`[   ]`, `[=  ]`, `[== ]`, `[===]`, `[ ==]`, `[  =]`, `[   ]`, `[  =]`, `[ ==]`, `[===]`, `[== ]`, `[=  ]`}
		for progress {
			fmt.Printf("\r%s file:%d dependencies:%d", logos[bar], find, deps)
			bar = (bar + 1) % len(logos)
			<-time.After(time.Millisecond * 100)
		}
	}()

	// 记录解析过的文件及依赖
	arg.ResCallFunc = func(file *model.File, root ...*model.DepGraph) {
		find++
		for _, dep := range root {
			dep.ForEachNode(func(p, n *model.DepGraph) bool {
				if n.Name != "" {
					deps++
				}
				return true
			})
		}
	}

	return func() {
		progress = false
	}
}

func taskReport(r opensca.TaskResult) format.Report {

	path := config.Conf().Path
	optional := config.Conf().Optional

	report := format.Report{
		ToolVersion: version,
		AppName:     path,
		Size:        r.Size,
	}

	if r.Error != nil {
		report.ErrorString = r.Error.Error()
	}

	// 合并检测结果
	root := &model.DepGraph{}
	if len(r.Deps) > 1 {
		for _, dep := range r.Deps {
			root.AppendChild(dep)
		}
	} else if len(r.Deps) == 1 {
		root = r.Deps[0]
	}
	report.DepDetailGraph = detail.NewDepDetailGraph(root)

	// 组件去重
	if optional.Dedup {
		report.RemoveDedup()
	}

	// 去掉dev组件
	if !optional.SaveDev {
		report.RemoveDev()
	}

	// 查询组件详情(漏洞/许可证)
	err := detail.SearchDetail(report.DepDetailGraph)
	if err != nil {
		if report.ErrorString != "" {
			report.ErrorString += "\n"
		}
		report.ErrorString += err.Error()
	}

	// 仅保留漏洞组件
	if optional.VulnOnly {
		var deps []*detail.DepDetailGraph
		report.ForEach(func(n *detail.DepDetailGraph) bool {
			if len(n.Vulnerabilities) > 0 {
				deps = append(deps, n)
			}
			return true
		})
		report.DepDetailGraph = &detail.DepDetailGraph{Children: deps}
	}

	end := time.Now()
	report.StartTime = r.Start.Format("2006-01-02 15:04:05")
	report.EndTime = end.Format("2006-01-02 15:04:05")
	report.CostTime = end.Sub(r.Start).Seconds()

	return report
}

//go:embed config.json
var defaultConfig []byte

func init() {
	config.RegisterDefaultConfig(defaultConfig)
}
