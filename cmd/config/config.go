package config

import (
	"encoding/json"
	"flag"
	"os"
	"os/user"
	"path/filepath"

	"github.com/titanous/json5"
	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

type Config struct {
	Path        string `json:"path"`
	Output      string `json:"out"`
	LogFile     string `json:"log"`
	Dedup       bool   `json:"dedup"`
	DirOnly     bool   `json:"dir"`
	VulnOnly    bool   `json:"vuln"`
	ProgressBar bool   `json:"progress"`
	// remote db
	Url   string `json:"url"`
	Token string `json:"token"`
	// local db
	LocalDB  string              `json:"db"`
	Maven    []common.RepoConfig `json:"maven"`
	Composer []common.RepoConfig `json:"composer"`
	// data origin
	Origin map[string]OriginConfig `json:"origin"`

	Version bool `json:"-"`
}

type OriginConfig struct {
	Dsn   string `json:"dsn"`
	Table string `json:"table"`
}

var _config = &Config{
	ProgressBar: true,
	Maven: []common.RepoConfig{
		{Url: `https://repo.maven.apache.org/maven2/`},
		{Url: `https://maven.aliyun.com/repository/public`},
	},
	Composer: []common.RepoConfig{
		{Url: `https://repo.packagist.org/p2`},
	},
}

func Conf() *Config {
	return _config
}

// WriteConfig 写入配置
func WriteConfig(write func(config *Config)) {
	if write != nil {
		write(_config)
	}
}

// loadDefaultConfig 加载默认配置
func loadDefaultConfig() bool {

	defaultConfigPaths := []string{}

	// 读取工作目录的 config.json
	if p, err := os.Getwd(); err == nil {
		defaultConfigPaths = append(defaultConfigPaths, filepath.Join(p, "config.json"))
	}

	// 读取用户目录下的 opensca_config.json
	if user, err := user.Current(); err == nil {
		defaultConfigPaths = append(defaultConfigPaths, filepath.Join(user.HomeDir, "opensca_config.json"))
	}

	// 读取 opensca-cli 所在目录下的 config.json
	if p, err := os.Executable(); err == nil {
		defaultConfigPaths = append(defaultConfigPaths, filepath.Join(filepath.Dir(p), "config.json"))
	}

	for _, config := range defaultConfigPaths {
		if data, err := os.ReadFile(config); err == nil {
			err := json5.Unmarshal(data, &_config)
			if err == nil {
				logs.Debugf("load config %s", config)
				return true
			}
		}
	}

	return false
}

// LoadConfig 加载配置文件
func LoadConfig(filepath string) {

	if filepath == "" {
		logs.Debug("use default config")
		loadDefaultConfig()
		return
	}

	if _, err := os.Stat(filepath); err != nil {
		logs.Debugf("%s not exist, create default config file", filepath)
		CreateConfigFile(filepath)
	}

	f, err := os.Open(filepath)
	if err != nil {
		logs.Warnf("read file %s error: %v", filepath, err)
	}

	err = json5.NewDecoder(f).Decode(_config)
	if err != nil {
		logs.Warnf("unmarshal file %s error: %v", filepath, err)
	}
}

// CreateConfigFile 生成配置文件
func CreateConfigFile(filepath string) {
	data, err := json.MarshalIndent(_config, "", "    ")
	if err != nil {
		logs.Warnf("marshal config error: %v", err)
	}
	err = os.WriteFile(filepath, data, 0666)
	if err != nil {
		logs.Warnf("write file %s error: %v", filepath, err)
	}
}

func ParseArgs() {
	var config string
	flag.StringVar(&config, "config", "", "(可选) 指定配置文件路径,指定后启动程序时将默认使用配置参数，配置参数与命令行输入参数冲突时优先使用输入参数")
	flag.BoolVar(&_config.Version, "version", false, "-version 打印版本信息")
	flag.StringVar(&_config.Path, "path", _config.Path, "(必须) 指定要检测的文件或目录路径,例: -path ./foo 或 -path ./foo.zip")
	flag.StringVar(&_config.Url, "url", _config.Url, "(可选,与token需一起使用) 从云漏洞库查询漏洞,指定要连接云服务的地址,例:-url https://opensca.xmirror.cn")
	flag.StringVar(&_config.Token, "token", _config.Url, "(可选,与url需一起使用) 云服务验证token,需要在云服务平台申请")
	flag.BoolVar(&_config.VulnOnly, "vuln", _config.VulnOnly, "(可选) 结果仅保留有漏洞信息的组件,使用该参数不会保留组件层级结构")
	flag.StringVar(&_config.Output, "out", _config.Output, "(可选) 根据后缀保存为不同格式的文件,支持html/json/xml/csv/sqlite/cdx/spdx/swid/dsdx, 生成多个报告用,分割,例: -out out.json,out.html")
	flag.StringVar(&_config.LocalDB, "db", _config.LocalDB, `(未来将会弃用,可以在配置文件中配置"origin":{"json":{"dsn":"db.json"}}来指定) 指定本地漏洞库文件,例: -db db.json`)
	flag.BoolVar(&_config.ProgressBar, "progress", _config.ProgressBar, "(可选) 显示进度条")
	flag.BoolVar(&_config.Dedup, "dedup", _config.Dedup, "(可选) 相同组件去重")
	flag.BoolVar(&_config.DirOnly, "dironly", _config.DirOnly, "(可选) 仅检测目录，忽略压缩包，加速基于源码的检测")
	flag.StringVar(&_config.LogFile, "log", _config.LogFile, "(可选) 指定日志文件路径")
	flag.Parse()
	if _config.Version {
		return
	}
	LoadConfig(config)
	flag.Parse()
}
