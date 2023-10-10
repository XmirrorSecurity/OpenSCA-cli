package model

import (
	"fmt"
	"strings"
)

type Language string

const (
	Lan_None       Language = ""
	Lan_Java       Language = "Java"
	Lan_JavaScript Language = "JavaScript"
	Lan_Php        Language = "Php"
	Lan_Ruby       Language = "Ruby"
	Lan_Golang     Language = "Golang"
	Lan_Rust       Language = "Rust"
	Lan_Erlang     Language = "Erlang"
	Lan_Python     Language = "Python"
)

var purlMap = map[Language]string{
	Lan_Rust:       "cargo",
	Lan_Php:        "composer",
	Lan_Ruby:       "gem",
	Lan_Golang:     "golang",
	Lan_Java:       "maven",
	Lan_JavaScript: "npm",
	Lan_Python:     "pypi",
}

var purlRmap = map[string]Language{
	"cargo":    Lan_Rust,
	"composer": Lan_Php,
	"gem":      Lan_Ruby,
	"golang":   Lan_Golang,
	"maven":    Lan_Java,
	"npm":      Lan_JavaScript,
	"pypi":     Lan_Python,
}

func Purl(vendor, name, version string, language Language) string {
	pkg := ""
	if g, ok := purlMap[language]; ok {
		pkg = g
	}
	if vendor == "" {
		return fmt.Sprintf("pkg:%s/%s@%s", pkg, name, version)
	}
	return fmt.Sprintf("pkg:%s/%s/%s@%s", pkg, vendor, name, version)
}

func ParsePurl(purl string) (vendor, name, version string, language Language) {

	// purl示例 pkg:maven/org.apache.xmlgraphics/batik-anim@1.9.1?packaging=sources

	if i := strings.LastIndex(purl, "?"); i != -1 {
		purl = purl[:i]
	}

	if i := strings.Index(purl, "/"); i == -1 {
		return
	} else {
		if pkg := strings.Split(purl[:i], `:`); len(pkg) != 2 {
			return
		} else {
			if l, ok := purlRmap[pkg[1]]; ok {
				language = l
			}
		}
		purl = purl[i+1:]
	}

	if i := strings.LastIndex(purl, "@"); i == -1 {
		return
	} else {
		version = purl[i+1:]
		name = purl[:i]
	}

	if language == Lan_Java {
		if i := strings.LastIndex(name, "/"); i != -1 {
			vendor = name[:i]
			name = name[i+1:]
		}
	}

	return
}
