package model

import (
	"fmt"

	"github.com/xmirrorsecurity/opensca-cli/util/enum/language"
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

var purlStrMap = map[string]string{
	language.Rust.String():       "cargo",
	language.Php.String():        "composer",
	language.Ruby.String():       "gem",
	language.Golang.String():     "golang",
	language.Java.String():       "maven",
	language.JavaScript.String(): "npm",
	language.Python.String():     "pypi",
}

func (dep Dependency) Purl() string {
	group := ""
	if dep.Language == language.None {
		if g, ok := purlStrMap[dep.LanguageStr]; ok {
			group = g
		}
	} else {
		if g, ok := purlMap[dep.Language]; ok {
			group = g
		}
	}
	version := dep.GetVersion()
	if dep.Vendor == "" {
		return fmt.Sprintf("pkg:%s/%s@%s", group, dep.Name, version)
	}
	return fmt.Sprintf("pkg:%s/%s/%s@%s", group, dep.Vendor, dep.Name, version)
}
