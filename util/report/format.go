package report

import (
	"strings"
	"util/enum/language"
	"util/model"
)

// format 按照输出内容格式化(不可逆)
func format(dep *model.DepTree) {
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		node := q[0]
		q = append(q[1:], node.Children...)
		if node.Language != language.None {
			node.LanguageStr = node.Language.String()
		}
		if node.Version != nil {
			node.VersionStr = node.Version.Org
		}
		node.Path = node.Path[strings.Index(node.Path, "/")+1:]
		node.Language = language.None
		node.Version = nil
	}
}
