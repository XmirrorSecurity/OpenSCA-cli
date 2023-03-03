package model

import (
	"fmt"
	"util/enum/language"
)

var purlMap = map[language.Type]string{
	language.Rust:       "cargo",
	language.Php:        "composer",
	language.Ruby:       "gem",
	language.Golang:     "golang",
	language.Java:       "maven",
	language.JavaScript: "npm",
	language.Python:     "pypi",
}

func (dep DepTree) Purl() string {
	group := ""
	if g, ok := purlMap[dep.Language]; ok {
		group = g
	}
	version := dep.VersionStr
	if dep.Version != nil && dep.Version.Org != "" {
		version = dep.Version.Org
	}
	if dep.Vendor == "" {
		return fmt.Sprintf("pkg:%s/%s@%s", group, dep.Name, version)
	}
	return fmt.Sprintf("pkg:%s/%s/%s@%s", group, dep.Vendor, dep.Name, version)
}
