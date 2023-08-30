package groovy

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseGradle(file []*model.File) []*model.DepGraph {
	// TODO
	return nil
}

//go:embed opensca.gradle
var openscaGradle []byte

// gradle 脚本输出的依赖结构
type gradleDep struct {
	GroupId    string       `json:"groupId"`
	ArtifactId string       `json:"artifactId"`
	Version    string       `json:"version"`
	Children   []*gradleDep `json:"children"`
}

func GradleTree(dirpath string) []*model.DepGraph {

	pwd, err := os.Getwd()
	if err != nil {
		logs.Warn(err)
		return nil
	}
	defer os.Chdir(pwd)
	os.Chdir(dirpath)

	// 复制 opensca.gradle
	if err := os.WriteFile("opensca.gradle", openscaGradle, 0444); err != nil {
		logs.Warn(err)
		return nil
	}
	defer os.Remove("opensca.gradle")

	cmd := exec.Command("gradle", "--I", "opensca.gradle", "opensca")
	out, _ := cmd.CombinedOutput()

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	var roots []*model.DepGraph
	// 获取 gradle 解析内容
	startTag := `openscaDepStart`
	endTag := `openscaDepEnd`
	for {

		startIndex, endIndex := bytes.Index(out, []byte(startTag)), bytes.Index(out, []byte(endTag))
		if startIndex < 0 || endIndex < 0 {
			break
		}

		data := out[startIndex+len(startTag) : endIndex]
		out = out[endIndex+1:]

		gdep := &gradleDep{}
		err := json.Unmarshal(data, &gdep.Children)
		if err != nil {
			logs.Warn(err)
		}

		root := &model.DepGraph{Vendor: gdep.GroupId, Name: gdep.ArtifactId, Version: gdep.Version, Expand: gdep}
		root.ForEachNode(func(p, n *model.DepGraph) bool {
			g := n.Expand.(*gradleDep)
			for _, c := range g.Children {
				dep := _dep(c.GroupId, c.ArtifactId, c.Version)
				if dep == nil {
					continue
				}
				dep.Expand = c
				n.AppendChild(dep)
			}
			return true
		})

		roots = append(roots, root)
	}

	return roots
}
