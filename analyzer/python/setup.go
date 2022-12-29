package python

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"util/client"
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

// parseSetupTmp 在临时目录中解析setup.py文件 有的项目不止需要setup.py
func parseSetupTmp(root *model.DepTree, file *model.FileInfo) {
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
		pwd := temp.GetPwd()
		os.Chdir(tempdir)
		cmd := exec.Command("python", ossfile, setupfile)
		out, _ := cmd.CombinedOutput()
		os.Chdir(pwd)
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
}

// parseSetup 解析setup.py文件
func parseSetup(root *model.DepTree, file *model.FileInfo) {
	setupfile, err := filepath.Abs(file.Name)
	if err != nil {
		logs.Warn(err)
		return
	}
	setupfile = strings.ReplaceAll(setupfile, `\`, `/`)
	workdir := path.Dir(setupfile)
	//名称尽量特殊点，避免原项目存在同名文件
	ossfile := path.Join(workdir, "oss-xm1rr0r.py")
	// 创建 oss.py
	if err := os.WriteFile(ossfile, ossPy, 0444); err != nil {
		logs.Warn(err)
		return
	}
	defer os.Remove(ossfile)
	// 解析 setup.py
	pwd := temp.GetPwd()
	os.Chdir(workdir)
	cmd := exec.Command("python", ossfile, setupfile)
	out, _ := cmd.CombinedOutput()
	os.Chdir(pwd)
	startTag, endTag := `oss_start<<`, `>>oss_end`
	startIndex, endIndex := strings.Index(string(out), startTag), strings.Index(string(out), endTag)
	if startIndex == -1 || endIndex == -1 {
		logs.Debug(string(out))
		return
	} else {
		out = out[startIndex+len(startTag) : endIndex]
	}
	// 获取解析结果
	var dep setupDep
	if err := json.Unmarshal(out, &dep); err != nil {
		logs.Warn(err)
	}
	// 运行时提取的包名和版本号，只提取根目录的
	if strings.HasSuffix(file.Name, client.PackageBasePath+"/"+path.Base(file.Name)) {
		if dep.Name != "" {
			client.PackageName = dep.Name
		}
		if dep.Version != "" {
			client.PackageVersion = dep.Version
		}
	}
	root.Name = dep.Name
	root.Version = model.NewVersion(dep.Version)
	root.Licenses = append(root.Licenses, dep.License)
	for _, pkg := range [][]string{dep.Packages, dep.InstallRequires, dep.Requires} {
		for _, p := range pkg {
			index := strings.IndexAny(p, "~=<>")
			sub := model.NewDepTree(root)
			if index > -1 {
				sub.Name = p[:index]
				sub.Version = model.NewVersion(p[index:])
			} else {
				sub.Name = p
			}
		}
	}
}
