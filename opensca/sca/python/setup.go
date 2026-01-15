package python

import (
	_ "embed"
	"io"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

// ParseSetup 解析setup.py
func ParseSetup(file *model.File) *model.DepGraph {

	root := &model.DepGraph{Path: file.Relpath()}

	// 静态解析
	file.OpenReader(func(reader io.Reader) {
		data, err := io.ReadAll(reader)
		if err != nil {
			return
		}
		reg := regexp.MustCompile(`install_requires\s*=\s*\[([^\]]+)\]`)
		requires := reg.FindStringSubmatch(string(data))
		if len(requires) < 2 {
			return
		}
		model.ReadLineNoComment(strings.NewReader(requires[1]), model.PythonTypeComment, func(line string) {
			line = strings.Trim(strings.TrimSpace(line), `'",`)
			words := strings.Fields(line)
			if len(words) == 0 {
				return
			}
			name := words[0]
			var version string
			if len(words) > 1 {
				version = strings.Join(words[1:], "")
			}
			root.AppendChild(&model.DepGraph{
				Name:    name,
				Version: version,
			})
		})
	})

	return root
}
