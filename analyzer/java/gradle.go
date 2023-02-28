package java

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"util/enum/language"
	"util/logs"
	"util/model"
	"util/temp"
)

//go:embed oss.gradle
var ossGradle []byte

// gradle 脚本输出的依赖结构
type gradleDep struct {
	GroupId    string       `json:"groupId"`
	ArtifactId string       `json:"artifactId"`
	Version    string       `json:"version"`
	Children   []*gradleDep `json:"children"`
	// 对应的DepTree
	MapDep *model.DepTree `json:"-"`
}

// GradleDepTree 尝试获取 gradle 依赖树
func GradleDepTree(dirpath string, root *model.DepTree) {
	Len := len(root.Children)
	pwd := temp.GetPwd()
	os.Chdir(dirpath)
	// 复制 oss.gradle
	if err := os.WriteFile("oss.gradle", ossGradle, 0444); err != nil {
		logs.Warn(err)
		return
	}
	cmd := exec.Command("gradle", "--I", "oss.gradle", "oss")
	out, _ := cmd.CombinedOutput()
	// 删除 oss.gradle
	os.Remove("oss.gradle")
	os.Chdir(pwd)
	// 获取 gradle 解析内容
	startTag := `ossDepStart`
	endTag := `ossDepEnd`
	root.Direct = true
	for {
		startIndex, endIndex := bytes.Index(out, []byte(startTag)), bytes.Index(out, []byte(endTag))
		if startIndex > -1 && endIndex > -1 {
			data := out[startIndex+len(startTag) : endIndex]
			out = out[endIndex+1:]
			gdep := &gradleDep{MapDep: model.NewDepTree(root)}
			err := json.Unmarshal(data, &gdep.Children)
			if err != nil {
				logs.Warn(err)
			}
			q := []*gradleDep{gdep}
			for len(q) > 0 {
				n := q[0]
				d := n.MapDep
				d.Vendor = n.GroupId
				d.Name = n.ArtifactId
				d.Version = model.NewVersion(n.Version)
				d.Language = language.Java
				for _, c := range n.Children {
					c.MapDep = model.NewDepTree(d)
				}
				q = append(q[1:], n.Children...)
			}
			for _, c := range gdep.MapDep.Children {
				c.Direct = true
			}
		} else {
			break
		}
	}

	// 以root.children数量变化来确定gradle是否有检出结果
	if len(root.Children) != Len {
		gradleSuccess = true
	}
}

// parseGradle parse *.gradle or *.gradle.kts
func parseGradle(root *model.DepTree, file *model.FileInfo) {
	regexs := []*regexp.Regexp{
		regexp.MustCompile(`group: ?['"]([^\s"']+)['"], ?name: ?['"]([^\s"']+)['"], ?version: ?['"]([^\s"']+)['"]`),
		regexp.MustCompile(`group: ?['"]([^\s"']+)['"], ?module: ?['"]([^\s"']+)['"], ?version: ?['"]([^\s"']+)['"]`),
		regexp.MustCompile(`['"]([^\s:'"]+):([^\s:'"]+):([^\s:'"]+)['"]`),
	}
	for _, line := range strings.Split(string(file.Data), "\n") {
		for _, re := range regexs {
			match := re.FindStringSubmatch(line)
			// 有捕获内容
			if len(match) == 4 &&
				// 不以注释开头
				!strings.HasPrefix(strings.TrimSpace(line), "/") &&
				// 不是测试组件
				!strings.Contains(strings.ToLower(line), "testimplementation") &&
				// 去掉非组件内容
				!strings.Contains(line, "//") {
				ver := model.NewVersion(match[3])
				// 版本号正常
				if ver.Ok() {
					dep := model.NewDepTree(root)
					dep.Vendor = match[1]
					dep.Name = match[2]
					dep.Version = ver
					break
				}
			}
		}
	}
}
