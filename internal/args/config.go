package args

import (
	"encoding/json"
	"opensca/internal/logs"
	"os"
)

// loadConfigFile 加载配置文件
func loadConfigFile() bool {
	configFilePath := Config
	if configFilePath == "" {
		return false
	}
	if _, err := os.Stat(configFilePath); err != nil {
		logs.Error(err)
		return false
	}
	if data, err := os.ReadFile(configFilePath); err != nil {
		logs.Error(err)
		return false
	} else {
		config := struct {
			Path        string `json:"path"`
			DB          string `json:"db"`
			Url         string `json:"url"`
			Token       string `json:"token"`
			Out         string `json:"out"`
			Cache       *bool  `json:"cache"`
			OnlyVuln    *bool  `json:"vuln"`
			ProgressBar *bool  `json:"progress"`
		}{}
		if err = json.Unmarshal(data, &config); err != nil {
			logs.Error(err)
			return false
		}
		if Filepath == "" && config.Path != "" {
			Filepath = config.Path
		}
		if VulnDB == "" && config.DB != "" {
			VulnDB = config.DB
		}
		if Url == "" && config.Url != "" {
			Url = config.Url
		}
		if Token == "" && config.Token != "" {
			Token = config.Token
		}
		if Out == "" && config.Out != "" {
			Out = config.Out
		}
		if !Cache && config.Cache != nil {
			Cache = *config.Cache
		}
		if !OnlyVuln && config.OnlyVuln != nil {
			OnlyVuln = *config.OnlyVuln
		}
		if !ProgressBar && config.ProgressBar != nil {
			ProgressBar = *config.ProgressBar
		}
		return true
	}
}
