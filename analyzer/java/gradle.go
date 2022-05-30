package java

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
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
	return
}
