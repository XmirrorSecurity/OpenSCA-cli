package groovy

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

// ParseGradle 解析gradle脚本
func ParseGradle(files []*model.File) []*model.DepGraph {

	v := Variable{}
	gradle := []*model.File{}
	for _, f := range files {
		if filter.GroovyGradle(f.Relpath()) {
			v.Scan(f)
			gradle = append(gradle, f)
		}
	}

	var roots []*model.DepGraph

	for _, f := range gradle {

		_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
			return &model.DepGraph{
				Vendor:  s[0],
				Name:    s[1],
				Version: s[2],
				Develop: s[3] == "dev",
			}
		}).LoadOrStore

		root := &model.DepGraph{Path: f.Relpath()}

		f.ReadLineNoComment(model.CTypeComment, func(line string) {

			line = v.Replace(line)

			for _, re := range regexs {
				match := re.FindStringSubmatch(line)
				if len(match) < 4 {
					continue
				}

				vendor := match[1]
				name := match[2]
				version := match[3]
				index := strings.Index(version, "@")
				if index != -1 {
					version = version[:index]
				}

				if version == "" || strings.Contains(version, "$") {
					continue
				}

				dev := ""
				if strings.Contains(strings.ToLower(line), "testimplementation") {
					dev = "dev"
				}

				root.AppendChild(_dep(vendor, name, version, dev))
			}

		})

		roots = append(roots, root)
	}

	return roots
}

// TODO: 优化gradle解析
// 依赖冲突 https://docs.gradle.org/current/userguide/dependency_management.html
// 依赖定义 https://docs.gradle.org/current/userguide/dependency_downgrade_and_exclude.html#sec:enforcing_dependency_version
var regexs = []*regexp.Regexp{
	regexp.MustCompile(`group: ?['"]([a-zA-Z]+[^\s"']+)['"], ?name: ?['"]([a-zA-Z]+[^\s"']+)['"], ?version: ?['"]([^\s"']+)['"]`),
	regexp.MustCompile(`group: ?['"]([a-zA-Z]+[^\s"']+)['"], ?module: ?['"]([a-zA-Z]+[^\s"']+)['"], ?version: ?['"]([^\s"']+)['"]`),
	regexp.MustCompile(`['"]([a-zA-Z]+[^\s:'"]+):([a-zA-Z]+[^\s:'"]+):([^\s:'"]+)['"]`),
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
