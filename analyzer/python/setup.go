package python

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"strings"
	"util/logs"
	"util/model"
	"util/temp"
)

//go:embed oss.py
var ossPy []byte

// oss.py 脚本输出的依赖结构
type setupDep struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	License         string   `json:"license"`
	Packages        []string `json:"packages"`
	InstallRequires []string `json:"install_requires"`
	Requires        []string `json:"requires"`
}

// parseSetup 解析 setup.py 文件
func parseSetup(root *model.DepTree, file *model.FileInfo) {
	temp.DoInTempDir(func(tempdir string) {
		ossfile := path.Join(tempdir, "oss.py")
		setupfile := path.Join(tempdir, "setup.py")
		// 创建 oss.py
		if err := os.WriteFile(ossfile, ossPy, 0444); err != nil {
			logs.Warn(err)
			return
		}
		// 创建 setup.py
		if err := os.WriteFile(setupfile, file.Data, 0444); err != nil {
			logs.Warn(err)
			return
		}
		// 解析 setup.py
		cmd := exec.Command("python", ossfile, setupfile)
		out, _ := cmd.CombinedOutput()
		startTag, endTag := `oss_start<<`, `>>oss_end`
		startIndex, endIndex := strings.Index(string(out), startTag), strings.Index(string(out), endTag)
		if startIndex == -1 || endIndex == -1 {
			return
		} else {
			out = out[startIndex+len(startTag) : endIndex]
		}
		// 获取解析结果
		var dep setupDep
		if err := json.Unmarshal(out, &dep); err != nil {
			logs.Warn(err)
		}
		root.Name = dep.Name
		root.Version = model.NewVersion(dep.Version)
		root.Licenses = append(root.Licenses, dep.License)
		for _, pkg := range [][]string{dep.Packages, dep.InstallRequires, dep.Requires} {
			for _, p := range pkg {
				index := strings.IndexAny(p, "=<>")
				sub := model.NewDepTree(root)
				if index > -1 {
					sub.Name = p[:index]
					sub.Version = model.NewVersion(p[index:])
				} else {
					sub.Name = p
				}
			}
		}
	})
	return
}
