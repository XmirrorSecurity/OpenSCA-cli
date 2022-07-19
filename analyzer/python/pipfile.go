package python

import (
	"encoding/json"
	"strings"
	"util/logs"
	"util/model"

	"github.com/BurntSushi/toml"
)

// parsePipfile parse Pipfile file
func parsePipfile(root *model.DepTree, file *model.FileInfo) {
	pip := struct {
		DevPackages map[string]string `toml:"dev-packages"`
		Packages    map[string]string `toml:"packages"`
	}{}
	if err := toml.Unmarshal(file.Data, &pip); err != nil {
		logs.Warn(err)
	}
	for name, version := range pip.Packages {
		dep := model.NewDepTree(root)
		dep.Name = name
		dep.Version = model.NewVersion(formatVer(version))
	}
	for name, version := range pip.DevPackages {
		dep := model.NewDepTree(root)
		dep.Name = name
		dep.Version = model.NewVersion(formatVer(version))
	}
}

// parsePipfileLock parse pipfile.lock file
func parsePipfileLock(root *model.DepTree, file *model.FileInfo) {
	lock := struct {
		Default map[string]struct {
			Version string `json:"version"`
		} `json:"default"`
	}{}
	err := json.Unmarshal(file.Data, &lock)
	if err != nil {
		logs.Warn(err)
	}
	names := []string{}
	for n := range lock.Default {
		names = append(names, n)
	}
	for _, n := range names {
		v := lock.Default[n].Version
		if v != "" {
			dep := model.NewDepTree(root)
			dep.Name = n
			dep.Version = model.NewVersion(formatVer(v))
		}
	}
	return
}

// 后续使用其他办法确定版本号
func formatVer(v string) string {
	res := strings.ReplaceAll(v, "==", "")
	res = strings.ReplaceAll(res, "~=", "")
	res = strings.ReplaceAll(res, ">=", "")
	res = strings.ReplaceAll(res, "<=", "")
	return res
}
