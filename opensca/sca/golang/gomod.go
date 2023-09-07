package golang

import (
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

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
