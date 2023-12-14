package python

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

// ParseSetup 解析setup.py
func ParseSetup(file *model.File) *model.DepGraph {

	// 尝试调用python解析
	root := ParseSetupPyWithPython(file)
	if root != nil && len(root.Children) > 0 {
		return root
	}

	root = &model.DepGraph{Path: file.Relpath()}

	// 静态解析
	file.OpenReader(func(reader io.Reader) {
		data, err := io.ReadAll(reader)
		if err != nil {
			return
		}
		reg := regexp.MustCompile(`install_requires\s*=\s*\[([^\]]+)\]`)
		requires := reg.FindStringSubmatch(string(data))
		if len(requires) < 2 {
			return
		}
		model.ReadLineNoComment(strings.NewReader(requires[1]), model.PythonTypeComment, func(line string) {
			line = strings.Trim(strings.TrimSpace(line), `'",`)
			words := strings.Fields(line)
			name := words[0]
			version := strings.Join(words[1:], "")
			root.AppendChild(&model.DepGraph{
				Name:    name,
				Version: version,
			})
		})
	})

	return root
}

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

func ParseSetupPyWithPython(file *model.File) *model.DepGraph {

	if _, err := exec.LookPath("python"); err != nil {
		return nil
	}

	dir := filepath.Dir(file.Abspath())
	ossfile := filepath.Join(dir, "oss.py")

	// 创建 oss.py
	if err := os.WriteFile(ossfile, ossPy, 0777); err != nil {
		logs.Warn(err)
		return nil
	}

	// 解析 setup.py
	cmd := exec.Command("python", ossfile, file.Abspath())
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

	root := &model.DepGraph{Name: dep.Name, Version: dep.Version, Path: file.Relpath()}
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
