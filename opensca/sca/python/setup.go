package python

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
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

func ParseSetup(file *model.File) *model.DepGraph {

	if _, err := exec.LookPath("python"); err != nil {
		return nil
	}

	ossfile := path.Join(file.Abspath, "oss.py")
	setupfile := path.Join(file.Abspath, "setup.py")

	// 创建 oss.py
	if err := os.WriteFile(ossfile, ossPy, 0444); err != nil {
		logs.Warn(err)
		return nil
	}

	// 创建 setup.py
	file.OpenReader(func(reader io.Reader) {
		f, err := os.Create(setupfile)
		if err != nil {
			logs.Warn(err)
			return
		}
		defer f.Close()
		io.Copy(f, reader)
	})

	// 解析 setup.py
	cmd := exec.Command("python", ossfile, setupfile)
	out, _ := cmd.CombinedOutput()
	startTag, endTag := `opensca_start<<`, `>>opensca_end`
	startIndex, endIndex := strings.Index(string(out), startTag), strings.Index(string(out), endTag)
	if startIndex == -1 || endIndex == -1 {
		return nil
	} else {
		out = out[startIndex+len(startTag) : endIndex]
	}

	// 获取解析结果
	var dep setupDep
	if err := json.Unmarshal(out, &dep); err != nil {
		logs.Warn(err)
		return nil
	}

	root := &model.DepGraph{Name: dep.Name, Version: dep.Version, Path: file.Path()}
	root.AppendLicense(dep.License)

	for _, pkg := range [][]string{dep.Packages, dep.InstallRequires, dep.Requires} {
		for _, p := range pkg {
			index := strings.IndexAny(p, "=<>")
			var name, version string
			if index > -1 {
				name = p[:index]
				version = p[index:]
			} else {
				name = p
			}
			root.AppendChild(&model.DepGraph{Name: name, Version: version})
		}
	}

	return root
}
