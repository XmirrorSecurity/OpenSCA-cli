/*
 * @Descripation: 程序启动参数
 * @Date: 2021-11-06 16:50:53
 */

package args

import "flag"

var (
	// 配置文件路径
	Config string
	// 解析文件路径
	Filepath string
	// 云服务地址
	Url string
	// 云服务token
	Token string
	// 开启本地缓存
	Cache bool
	// 输出文件
	Out string
	// 仅展示有漏洞的组件
	OnlyVuln bool
	// 本地漏洞库文件路径
	VulnDB string
)

func Parse() {
	// 设置参数信息
	flag.StringVar(&Config, "config", "", "(可选) 指定配置文件路径,指定后启动程序时将默认使用配置参数，配置参数与命令行输入参数冲突时优先使用输入参数")
	flag.StringVar(&Filepath, "path", "", "(必须) 指定要检测的文件或目录路径,例: -path ./foo 或 -path ./foo.zip")
	flag.StringVar(&Url, "url", "", "(可选,与token需一起使用) 从云漏洞库查询漏洞,指定要连接云服务的地址,仅支持http协议,需要写明端口,例:-url http://opensca.xmirror.cn:8003")
	flag.StringVar(&Token, "token", "", "(可选,与url需一起使用) 云服务验证token,需要在云服务平台申请")
	flag.BoolVar(&Cache, "cache", false, "(可选,建议开启) 缓存下载的文件(例如pom文件),重复检测相同组件时会节省时间,下载的文件会保存到工具所在目录的.cache目录下")
	flag.BoolVar(&OnlyVuln, "vuln", false, "(可选) 结果仅保留有漏洞信息的组件,使用该参数不会保留组件层级结构")
	flag.StringVar(&Out, "out", "", "(可选) 将检测结果保存到指定文件,检测结果为json格式,例: -out output.json")
	flag.StringVar(&VulnDB, "db", "", "(可选) 指定本地漏洞库文件,希望使用自己漏洞库时可用,漏洞库文件为json格式,具体格式会在开源项目文档中给出;若同时使用云端漏洞库与本地漏洞库,漏洞查询结果取并集,例: -db db.json")
	flag.Parse()
	loadConfigFile()
}
