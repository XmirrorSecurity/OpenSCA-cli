package golang

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

// ParseGomod 解析go.mod文件
func ParseGomod(file *model.File) *model.DepGraph {

	root := &model.DepGraph{Path: file.Relpath}

	var require bool

	file.ReadLineNoComment(&model.FileCommentType{
		Simple: "//",
	}, func(line string) {

		if strings.HasPrefix(line, "module") {
			root.Name = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return
		}

		if strings.HasPrefix(line, "require") {
			require = true
			return
		}

		if strings.HasPrefix(line, ")") {
			require = false
			return
		}

		// 不处理require之外的模块
		if !require {
			return
		}

		line = strings.TrimSpace(line)
		words := strings.Fields(line)
		if len(words) >= 2 {
			root.AppendChild(&model.DepGraph{
				Name:    strings.Trim(words[0], `'"`),
				Version: strings.TrimSuffix(words[1], "+incompatible"),
			})
		}

	})

	return root
}

// ParseGosum 解析go.sum文件
func ParseGosum(file *model.File) *model.DepGraph {

	depMap := map[string]string{}
	file.ReadLine(func(line string) {
		line = strings.TrimSpace(line)
		words := strings.Fields(line)
		if len(words) >= 2 {
			name := strings.Trim(words[0], `'"`)
			version := strings.TrimSuffix(words[1], "+incompatible")
			depMap[name] = version
		}
	})

	root := &model.DepGraph{Path: file.Relpath}
	for name, version := range depMap {
		root.AppendChild(&model.DepGraph{
			Name:    name,
			Version: version,
		})
	}

	return root
}

// GoModGraph 调用 go mod graph 解析依赖
func GoModGraph(ctx context.Context, dir *model.File) []*model.DepGraph {

	_, err := exec.LookPath("go")
	if err != nil {
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	defer os.Chdir(pwd)
	os.Chdir(dir.Abspath)

	cmd := exec.CommandContext(ctx, "go", "mod", "graph")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Name:    s[0],
			Version: s[1],
		}
	})

	parse := func(s string) *model.DepGraph {
		words := strings.Split(s, "@")
		if len(words) == 2 {
			return _dep.LoadOrStore(words...)
		}
		return _dep.LoadOrStore(s, "")
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		words := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(words) == 2 {
			parent := parse(words[0])
			child := parse(words[1])
			parent.AppendChild(child)
		}
	}

	var roots []*model.DepGraph
	_dep.Range(func(k string, v *model.DepGraph) bool {
		if len(v.Parents) == 0 {
			v.Path = dir.Relpath
			roots = append(roots, v)
		}
		return true
	})

	return roots
}
