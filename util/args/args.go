/*
 * @Descripation: 程序启动参数
 * @Date: 2021-11-06 16:50:53
 */

package args

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"util/temp"
)

var (
	ConfigPath string
	Config     = struct {
		// detect option
		Path     string `json:"path"`
		Out      string `json:"out"`
		Cache    bool   `json:"cache"`
		Bar      bool   `json:"progress"`
		OnlyVuln bool   `json:"vuln"`
		Dedup    bool   `json:"dedup"`
		// remote vuldb
		Url   string `json:"url"`
		Token string `json:"token"`
		// local vuldb
		VulnDB string `json:"db"`
		// prvate repository
		Maven []RepoConfig `json:"maven"`
	}{
		Cache: true,
	}
)

func init() {
	flag.StringVar(&ConfigPath, "config", "", "(可选) 指定配置文件路径,指定后启动程序时将默认使用配置参数，配置参数与命令行输入参数冲突时优先使用输入参数")
	flag.StringVar(&Config.Path, "path", Config.Path, "(必须) 指定要检测的文件或目录路径,例: -path ./foo 或 -path ./foo.zip")
	flag.StringVar(&Config.Url, "url", Config.Url, "(可选,与token需一起使用) 从云漏洞库查询漏洞,指定要连接云服务的地址,例:-url https://opensca.xmirror.cn")
	flag.StringVar(&Config.Token, "token", Config.Token, "(可选,与url需一起使用) 云服务验证token,需要在云服务平台申请")
	flag.BoolVar(&Config.Cache, "cache", Config.Cache, "(可选,建议开启) 缓存下载的文件(例如pom文件),重复检测相同组件时会节省时间,下载的文件会保存到工具所在目录的.cache目录下")
	flag.BoolVar(&Config.OnlyVuln, "vuln", Config.OnlyVuln, "(可选) 结果仅保留有漏洞信息的组件,使用该参数不会保留组件层级结构")
	flag.StringVar(&Config.Out, "out", Config.Out, "(可选) 将检测结果保存到指定文件,根据后缀生成不同格式的文件,默认为json格式,例: -out output.json")
	flag.StringVar(&Config.VulnDB, "db", Config.VulnDB, "(可选) 指定本地漏洞库文件,希望使用自己漏洞库时可用,漏洞库文件为json格式,具体格式会在开源项目文档中给出;若同时使用云端漏洞库与本地漏洞库,漏洞查询结果取并集,例: -db db.json")
	flag.BoolVar(&Config.Bar, "progress", Config.Bar, "(可选) 显示进度条")
	flag.BoolVar(&Config.Dedup, "dedup", Config.Dedup, "(可选) 相同组件去重")
}

func Parse() {
	flag.Parse()
	if ConfigPath != "" {
		if data, err := ioutil.ReadFile(ConfigPath); err != nil {
			fmt.Printf("load config file error: %s\n", err)
		} else {
			if err = json.Unmarshal(data, &Config); err != nil {
				fmt.Printf("parse config file error: %s\n", err)
			}
		}
	} else {
		// 默认读取目录下的config.json文件
		if data, err := ioutil.ReadFile(path.Join(temp.GetPwd(), "config.json")); err == nil {
			// 不处理错误
			json.Unmarshal(data, &Config)
		}
	}
	// 再次调用Parse, 优先使用cli参数
	flag.Parse()
	Config.Url = strings.TrimSuffix(Config.Url, "/")
}
