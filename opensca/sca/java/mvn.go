package java

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func MvnTree(dirpath string) []*model.DepGraph {

	if _, err := exec.LookPath("mvn"); err != nil {
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		logs.Warn(err)
		return nil
	}
	defer os.Chdir(pwd)

	os.Chdir(dirpath)
	cmd := exec.Command("mvn", "dependency:tree", "--fail-never")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logs.Warn(err)
		return nil
	}

	// 记录当前处理的依赖树数据
	var lines []string
	// 标记是否在依赖范围内树
	tree := false
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)

	var roots []*model.DepGraph

	scan := bufio.NewScanner(bytes.NewBuffer(output))
	for scan.Scan() {
		line := strings.TrimPrefix(scan.Text(), "[INFO] ")
		if title.MatchString(line) {
			tree = true
			continue
		}
		if tree && strings.Trim(line, "-") == "" {
			tree = false
			root := parseMvnTree(lines)
			if root != nil {
				roots = append(roots, root)
			}
			lines = nil
			continue
		}
		if tree {
			lines = append(lines, line)
			continue
		}
	}

	return roots
}

func parseMvnTree(lines []string) *model.DepGraph {

	// 记录当前的顶点节点列表
	var tops []*model.DepGraph
	// 上一层级
	lastLevel := -1

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	for _, line := range lines {

		// 计算层级
		level := 0
		for level*3+2 < len(line) && line[level*3+2] == ' ' {
			level++
		}
		if level*3+2 >= len(line) {
			continue
		}

		if level-lastLevel > 1 {
			// 在某个依赖解析失败的时候 子依赖会出现这种情况
			continue
		}

		tags := strings.Split(line[level*3:], ":")
		if len(tags) < 4 {
			continue
		}

		dep := _dep(tags[0], tags[1], tags[3])

		if dep == nil {
			continue
		}

		scope := tags[len(tags)-1]
		if scope == "test" || scope == "provided" {
			dep.Develop = true
		}

		tops[len(tops)-1].AppendChild(dep)

		tops = append(tops[:len(tops)-lastLevel+level-1], dep)

		lastLevel = level
	}

	if len(tops) > 1 {
		return tops[0]
	} else {
		return nil
	}
}
