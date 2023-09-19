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

	start := time.Now()
	deps, err := opensca.RunTask(context.Background(), taskArg(path))
	end := time.Now()

	// 日志中记录检测结果
	for _, dep := range deps {
		logs.Debugf("dependency tree:\n%s", dep.Tree(false, false))
	}

	// 生成报告
	report := taskReport(start, end, deps)

	if err != nil {
		report.ErrorString = err.Error()
	}

	// 打印概览信息
	fmt.Println(format.Statis(report))

	// 导出报告
	format.Save(report, config.Conf().Output)

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
	flag.Parse()

	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	config.LoadConfig(cfgf)
	flag.Parse()

	logs.CreateLog(config.Conf().LogFile)

	if len(config.Conf().Path) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	java.RegisterMavenRepo(config.Conf().Repo.Maven...)
	php.RegisterComposerRepo(config.Conf().Repo.Composer...)
}

func taskArg(origin string) *opensca.TaskArg {

	// 检测参数
	arg := &opensca.TaskArg{DataOrigin: origin}

	// 是否跳过压缩包检测
	if config.Conf().Optional.DirOnly {
		arg.ExtractFileFilter = func(relpath string) bool { return false }
	}

	// 进度条
	if config.Conf().Optional.ProgressBar {
		var find, done, deps, bar int
		go func() {
			logos := []string{`[   ]`, `[=  ]`, `[== ]`, `[===]`, `[ ==]`, `[  =]`, `[   ]`, `[  =]`, `[ ==]`, `[===]`, `[== ]`, `[=  ]`}
			for {
				fmt.Printf("\r%s find:%d done:%d dependencies:%d", logos[bar], find, done, deps)
				bar = (bar + 1) % len(logos)
				<-time.After(time.Millisecond * 100)
			}
		}()
		// 记录需要解析的文件
		arg.WalkFileFunc = func(parent *model.File, files []*model.File) {
			for range files {
				find++
			}
		}
		// 记录处理完的文件
		arg.DeferWalkFileFunc = func(parent *model.File, files []*model.File) {
			for range files {
				done++
			}
		}
		// 记录解析到的依赖个数
		arg.WalkDepFunc = func(dep *model.DepGraph) {
			dep.ForEachNode(func(p, n *model.DepGraph) bool {
				deps++
				return true
			})
		}
	}

	return arg
}

func taskReport(start, end time.Time, deps []*model.DepGraph) format.Report {

	path := config.Conf().Path
	optional := config.Conf().Optional

	report := format.Report{
		ToolVersion: version,
		AppName:     path,
		StartTime:   start.Format("2006-01-02 15:04:05"),
		EndTime:     end.Format("2006-01-02 15:04:05"),
		CostTime:    end.Sub(start).Seconds(),
	}

	// 记录检测目标文件大小
	if f, err := os.Stat(path); err == nil {
		report.Size = f.Size()
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
	report.DepDetailGraph = detail.NewDepDetailGraph(root)

	// 组件去重
	if optional.Dedup {
		report.Dedup()
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

	return report
}
