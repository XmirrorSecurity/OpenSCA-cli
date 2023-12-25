package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "embed"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/format"
	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/ui"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/java"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/javascript"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/php"
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

	// 初始化 HttpClient
	common.InitHttpClient(config.Conf().Optional.Insecure)

	// 运行检测任务
	result := opensca.RunTask(context.Background(), arg)

	// 日志中记录检测结果
	if result.Error != nil {
		logs.Error(result.Error)
	}
	for _, dep := range result.Deps {
		if dep.Name != "" || len(dep.Children) > 0 {
			logs.Debugf("dependency tree:\n%s", dep.Tree(false, false))
		}
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
	logs.Info("\nComplete!\n" + format.Statis(report))

	// 发送检测报告
	if err := format.Saas(report); err != nil {
		logs.Warnf("saas report error: %s", err)
	}

	// 开启ui
	if config.Conf().Optional.UI {
		ui.OpenUI(report)
	}

}

func args() {

	v := false
	login := false
	cfgf := ""
	proj := "x"
	cfg := config.Conf()
	flag.BoolVar(&v, "version", false, "-version")
	flag.BoolVar(&login, "login", false, "login to cloud server. example: -login")
	flag.StringVar(&cfgf, "config", "", "config path. example: -config config.json")
	flag.StringVar(&cfg.Path, "path", cfg.Path, "project path. example: -path project_path")
	flag.StringVar(&cfg.Output, "out", cfg.Output, "report path, support html/json/xml/csv/sqlite/cdx/spdx/swid/dsdx. example: -out out.json,out.html")
	flag.StringVar(&cfg.LogFile, "log", cfg.LogFile, "-log ./my_opensca_log.txt")
	flag.StringVar(&cfg.Origin.Token, "token", "", "web token, example: -token xxxx")
	flag.StringVar(&proj, "proj", proj, "saas project id, example: -proj xxxx")
	flag.Parse()

	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	cfgf = config.LoadConfig(cfgf)
	flag.Parse()

	cfg.Origin.Url = strings.TrimRight(cfg.Origin.Url, "/")
	if proj != "x" {
		cfg.Origin.Proj = &proj
	}

	logs.CreateLog(config.Conf().LogFile)

	logs.Infof("opensca-cli version: %s", version)
	logs.Infof("use config: %s", cfgf)

	if login {
		if err := detail.Login(); err != nil {
			fmt.Printf("login failure: %s\n", err)
		} else {
			fmt.Println("login success")
		}
	}

	java.RegisterMavenRepo(config.Conf().Repo.Maven...)
	javascript.RegisterNpmRepo(config.Conf().Repo.Npm...)
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

	report := format.Report{}
	report.TaskInfo.ToolVersion = version
	// make app name as file or dir name
	report.TaskInfo.AppName = filepath.Base(filepath.Clean(path))
	report.TaskInfo.Size = r.Size

	if r.Error != nil {
		report.TaskInfo.ErrorString = r.Error.Error()
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
		logs.Warnf("database origin error: %s", err.Error())
		if report.TaskInfo.ErrorString != "" {
			report.TaskInfo.ErrorString += "\n"
		}
		report.TaskInfo.ErrorString += err.Error()
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
		for _, d := range deps {
			d.Children = nil
		}
		report.DepDetailGraph = &detail.DepDetailGraph{Children: deps}
	}

	end := time.Now()
	report.TaskInfo.StartTime = r.Start.Format("2006-01-02 15:04:05")
	report.TaskInfo.EndTime = end.Format("2006-01-02 15:04:05")
	report.TaskInfo.CostTime = end.Sub(r.Start).Seconds()

	return report
}

//go:embed config.json
var defaultConfig []byte

func init() {
	config.RegisterDefaultConfig(defaultConfig)
}
