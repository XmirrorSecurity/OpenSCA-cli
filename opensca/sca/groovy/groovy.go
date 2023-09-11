package groovy

import (
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

// ParseGroovy 解析groovy文件
func ParseGroovy(file *model.File) *model.DepGraph {

	// @Grab(group='org.springframework', module='spring-orm', version='3.2.5.RELEASE')
	depLongReg := regexp.MustCompile(`@Grab\(group='([^'\s]+)', module='([^'\s]+)', version='([^'\s]+)'\)`)

	// @Grab('org.springframework:spring-orm:3.2.5.RELEASE;transitive=false')
	depShortReg := regexp.MustCompile(`@Grab\('([^:\s]+):([^:\s]+):([^:;'\)\s]+)`)

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	root := &model.DepGraph{Path: file.Relpath()}

	file.ReadLineNoComment(model.CTypeComment, func(line string) {
		for _, match := range append(depLongReg.FindAllStringSubmatch(line, -1), depShortReg.FindAllStringSubmatch(line, -1)...) {

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

			root.AppendChild(_dep(vendor, name, version))
		}
	})

	return root
}
