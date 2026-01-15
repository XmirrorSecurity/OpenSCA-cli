package golang

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

// ParseGomod 解析go.mod文件
func ParseGomod(file *model.File) *model.DepGraph {

	root := &model.DepGraph{Path: file.Relpath()}

	var require bool
	var replace bool

	parseDepLine := func(line string) (name, version string) {
		line = strings.TrimSpace(line)
		words := strings.Fields(line)
		if len(words) >= 2 {
			name = strings.Trim(words[0], `'"`)
			version = strings.TrimSuffix(words[1], "+incompatible")
		}
		return
	}

	deps := map[string]string{}

	file.ReadLineNoComment(&model.CommentType{
		Simple: "//",
	}, func(line string) {

		if strings.HasPrefix(line, "module") {
			root.Name = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return
		}

		if strings.HasPrefix(line, "require") {
			require = true
			line = strings.TrimPrefix(line, "require")
			if strings.TrimSpace(line) == "" {
				return
			}
		}

		if strings.HasPrefix(line, "replace") {
			replace = true
			line = strings.TrimPrefix(line, "replace")
			if strings.TrimSpace(line) == "" {
				return
			}
		}

		if strings.HasPrefix(line, ")") {
			require = false
			replace = false
			return
		}

		if require {
			name, version := parseDepLine(line)
			if name != "" {
				deps[name] = version
			}
			return
		}

		if replace {
			line = strings.TrimSpace(line)
			i := strings.Index(line, "=>")
			if i == -1 {
				return
			}
			old := strings.TrimSpace(line[:i])
			delete(deps, old)
			name, version := parseDepLine(line[i+2:])
			if name != "" {
				deps[name] = version
			}
			return
		}

	})

	for name, version := range deps {
		root.AppendChild(&model.DepGraph{
			Name:    name,
			Version: version,
		})
	}

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
			version := strings.TrimSuffix(words[1], "/go.mod")
			version = strings.TrimSuffix(version, "+incompatible")
			depMap[name] = version
		}
	})

	root := &model.DepGraph{Path: file.Relpath()}
	for name, version := range depMap {
		root.AppendChild(&model.DepGraph{
			Name:    name,
			Version: version,
		})
	}

	sort.Slice(root.Children, func(i, j int) bool {
		return root.Children[i].Name < root.Children[j].Name
	})

	return root
}

// GoModGraph 调用 go mod graph 解析依赖
func GoModGraph(ctx context.Context, modfile *model.File) *model.DepGraph {

	if !config.Conf().Optional.Dynamic {
		return nil
	}

	_, err := exec.LookPath("go")
	if err != nil {
		return nil
	}

	cmd := exec.CommandContext(ctx, "go", "mod", "graph")
	cmd.Dir = filepath.Dir(modfile.Abspath())
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

	root := &model.DepGraph{Path: modfile.Relpath()}
	_dep.Range(func(k string, v *model.DepGraph) bool {
		if len(v.Parents) == 0 {
			root.AppendChild(v)
		}
		return true
	})

	return root
}
