/*
 * @Descripation: mvn解析依赖树
 * @Date: 2021-12-16 10:10:13
 */

package java

import (
	"bytes"
	"encoding/xml"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"util/enum/language"
	"util/filter"
	"util/logs"
	"util/model"
	"util/temp"
)

// MvnDepTree 调用mvn解析项目获取依赖树
func MvnDepTree(dirpath string, root *model.DepTree) {
	root.Direct = true
	poms := map[string]Pom{}
	// 记录目录下的pom文件
	filepath.WalkDir(dirpath, func(fullpath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !filter.JavaPom(d.Name()) {
			return nil
		}
		fullpath = strings.ReplaceAll(fullpath, `\`, `/`)
		data, err := os.ReadFile(fullpath)
		if err != nil {
			logs.Warn(err)
			return nil
		}
		pom := Pom{}
		xml.Unmarshal(data, &pom)
		if pom.GroupId == "" && pom.Parent.GroupId != "" {
			pom.GroupId = pom.Parent.GroupId
		}
		poms[fullpath] = pom
		return nil
	})
	// 关联mvn结果和pom文件
	deps := []*model.DepTree{}
	for pomfile, pom := range poms {
		for _, d := range parseMvnOutput(path.Dir(pomfile)) {
			if d.Vendor == pom.GroupId && d.Name == pom.ArtifactId {
				d.Direct = true
				d.Path = pomfile
				q := []*model.DepTree{d}
				for len(q) > 0 {
					n := q[0]
					for _, nc := range n.Children {
						nc.Path = path.Join(n.Path, nc.Dependency.String())
					}
					q = append(q[1:], n.Children...)
				}
				deps = append(deps, d)
			}
		}
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Path < deps[j].Path
	})
	for _, d := range deps {
		d.Parent = root
		root.Children = append(root.Children, d)
	}
	// 判断mvn是否调用成功
	mvnSuccess = len(deps) > 0
}

func parseMvnOutput(dirpath string) []*model.DepTree {
	pwd := temp.GetPwd()
	os.Chdir(dirpath)
	cmd := exec.Command("mvn", "dependency:tree", "--fail-never")
	out, _ := cmd.CombinedOutput()
	os.Chdir(pwd)
	// 统一替换换行符为\n
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\n\r"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\r"), []byte("\n"))
	// 获取mvn解析内容
	lines := strings.Split(string(out), "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "[INFO] ")
	}
	// 记录依赖树起始位置行号
	start := 0
	// 标记是否在依赖范围内树
	tree := false
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)
	tops := []*model.DepTree{}
	// 获取mvn依赖树
	for i, line := range lines {
		if title.MatchString(line) {
			tree = true
			start = i
			continue
		}
		if tree && strings.Trim(line, "-") == "" {
			tree = false
			top := buildMvnDepTree(lines[start+1 : i])
			if top != nil {
				tops = append(tops, top)
			}
			continue
		}
	}
	return tops
}

// buildMvnDepTree 构建mvn树
func buildMvnDepTree(lines []string) *model.DepTree {
	// 记录当前的顶点节点列表
	root := model.NewDepTree(nil)
	tops := []*model.DepTree{root}
	// 上一层级
	lastLevel := -1
	for _, line := range lines {
		// 计算层级
		level := 0
		// 防止数组越界
		if level*3+2 >= len(line) {
			continue
		}
		for line[level*3+2] == ' ' {
			level++
		}
		root = tops[len(tops)-1]
		tags := strings.Split(line[level*3:], ":")
		if len(tags) < 4 {
			continue
		}
		scope := tags[len(tags)-1]
		if scope == "test" || scope == "provided" {
			continue
		}
		if level-lastLevel > 1 {
			// 在某个依赖解析失败的时候 子依赖会出现这种情况
			continue
		}
		dep := model.NewDepTree(root)
		dep.Vendor = tags[0]
		dep.Name = tags[1]
		dep.Version = model.NewVersion(tags[3])
		dep.Language = language.Java
		tops = tops[:len(tops)-lastLevel+level-1]
		tops = append(tops, dep)
		lastLevel = level
	}
	if len(tops) > 1 {
		return tops[1]
	} else {
		return nil
	}
}
