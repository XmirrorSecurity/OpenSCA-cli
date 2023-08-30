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
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java"
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

	// 查询组件详情(漏洞/许可证)
	report.DepDetailGraph = detail.NewDepDetailGraph(root)
	err = detail.SearchDetail(report.DepDetailGraph)
	if err != nil {
		if report.ErrorString != "" {
			report.ErrorString += "\n"
		}
		report.ErrorString += err.Error()
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

	var v bool
	flag.BoolVar(&v, "version", false, "-version 打印版本信息")

	flag.Parse()

	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	config.ParseArgs()

	path := config.Conf().Path
	if len(path) == 0 {
		flag.PrintDefaults()
		return
	}

	var repos []java.MvnRepo
	for _, repo := range config.Conf().Maven {
		repos = append(repos, java.MvnRepo{
			Url:      repo.Repo,
			Username: repo.User,
			Password: repo.Password,
		})
	}
	if len(repos) == 0 {
		repos = []java.MvnRepo{
			{Url: "https://maven.aliyun.com/repository/public"},
			{Url: "https://repo1.maven.org/maven2"},
		}
	}
	java.RegisterRepo(repos...)

}
