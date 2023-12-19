package config

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/titanous/json5"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

type Config struct {
	BaseConfig
	Optional OptionalConfig `json:"optional"`
	Repo     RepoConfig     `json:"repo"`
	Origin   OriginConfig   `json:"origin"`
}

type BaseConfig struct {
	Path    string `json:"path"`
	Output  string `json:"out"`
	LogFile string `json:"log"`
}

type OriginConfig struct {
	Url    string    `json:"url"`
	Token  string    `json:"token"`
	Proj   *string   `json:"proj"`
	Json   string    `json:"json"`
	Mysql  SqlOrigin `json:"mysql"`
	Sqlite SqlOrigin `json:"sqlite"`
}

type OptionalConfig struct {
	UI          bool `json:"ui"`
	Dedup       bool `json:"dedup"`
	DirOnly     bool `json:"dir"`
	VulnOnly    bool `json:"vuln"`
	SaveDev     bool `json:"dev"`
	ProgressBar bool `json:"progress"`
	Insecure    bool `json:"insecure"`
}

type RepoConfig struct {
	Maven    []common.RepoConfig `json:"maven"`
	Npm      []common.RepoConfig `json:"npm"`
	Composer []common.RepoConfig `json:"composer"`
}

type SqlOrigin struct {
	Dsn   string `json:"dsn"`
	Table string `json:"table"`
}

var _config *Config = &Config{}

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

var defalutConfigJson []byte

func RegisterDefaultConfig(data []byte) {
	defalutConfigJson = data
	if _config == nil {
		_config = &Config{}
		json5.Unmarshal(defalutConfigJson, &_config)
	}
}

// CreateConfigFile 生成配置文件
func CreateConfigFile(filepath string) {
	err := os.WriteFile(filepath, defalutConfigJson, 0666)
	if err != nil {
		logs.Warnf("write file %s error: %v", filepath, err)
	}
}
