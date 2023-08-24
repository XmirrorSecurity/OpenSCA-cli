package golang

import (
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseGomod(file *model.File) *model.DepGraph {
	root := &model.DepGraph{Path: file.Relpath}
	reg := regexp.MustCompile(`(\S*)\s+v([\d\w\-+.]*)[\s\n]`)
	file.ReadLine(func(line string) {
		if !reg.MatchString(line) {
			return
		}
		match := reg.FindStringSubmatch(line)
		root.AppendChild(&model.DepGraph{
			Name:    strings.Trim(match[1], `'"`),
			Version: match[2],
		})
	})
	return root
}

func ParseGosum(file *model.File) *model.DepGraph {

	root := ParseGomod(file)

	exist := map[string]bool{}
	for dep := range root.Children {
		exist[dep.Name] = true
	}

	reg := regexp.MustCompile(`(\S*)\s+v([\d\w\-+.]*)/go.mod[\s\n]`)
	file.ReadLine(func(line string) {
		match := reg.FindStringSubmatch(line)
		if len(match) != 3 || exist[match[1]] {
			return
		}
		exist[match[1]] = true
		root.AppendChild(&model.DepGraph{
			Name:    match[1],
			Version: match[2],
		})
	})

	return root
}
