package ruby

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

// ParseGemfileLock 解析Gemfile.lock文件
func ParseGemfileLock(file *model.File) *model.DepGraph {

	// map[name]dep
	depMap := map[string]*model.DepGraph{}

	space4 := "    "
	space6 := "      "
	parseLine := func(line string) (name, version string) {
		line = strings.TrimSpace(line)
		i := strings.Index(line, " ")
		if i == -1 {
			name = line
			return
		}
		name = line[:i]
		version = strings.Trim(line[i+1:], "()")
		return
	}

	// 第一次记录依赖信息
	file.ReadLine(func(line string) {
		if !strings.HasPrefix(line, space4) || strings.HasPrefix(line, space6) {
			return
		}
		name, version := parseLine(line)
		if version == "" {
			return
		}
		depMap[name] = &model.DepGraph{
			Name:     name,
			Version:  version,
			Language: model.Lan_Ruby,
		}
	})

	// 第二次记录依赖关系
	var last string
	file.ReadLine(func(line string) {
		if strings.HasPrefix(line, space6) {
			name, _ := parseLine(line)
			parent := depMap[last]
			child := depMap[name]
			if parent != nil && child != nil {
				parent.AppendChild(child)
			}
			return
		}
		if strings.HasPrefix(line, space4) {
			last, _ = parseLine(line)
			return
		}
	})

	root := &model.DepGraph{Path: file.Relpath}
	for _, d := range depMap {
		if len(d.Parents) == 0 {
			root.AppendChild(d)
		}
	}

	return root
}
