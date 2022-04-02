/*
 * @Descripation: 语言类型
 * @Date: 2021-11-04 20:53:03
 */

package language

import (
	"strings"
)

// 语言类型
type Type int

const (
	None Type = iota
	Java
	JavaScript
	Php
	Ruby
	Golang
	Rust
)

/**
 * @description: 语言类型
 * @return {string} 语言类型
 */
func (l Type) String() string {
	switch l {
	case None:
		return "None"
	case Java:
		return "Java"
	case JavaScript:
		return "JavaScript"
	case Php:
		return "Php"
	case Ruby:
		return "Ruby"
	case Golang:
		return "Golang"
	case Rust:
		return "Rust"
	default:
		return "None"
	}
}

/**
 * @description: 漏洞语言类型
 * @return {string} 漏洞语言类型
 */
func (l Type) Vuln() string {
	switch l {
	case None:
		return ""
	case Java:
		return "java"
	case JavaScript:
		return "js"
	case Php:
		return "php"
	case Ruby:
		return "ruby"
	case Golang:
		return "golang"
	case Rust:
		return "rust"
	default:
		return ""
	}
}

var (
	lanMap = map[string]Type{}
)

func init() {
	lm := map[Type][]string{}
	lm[Java] = []string{"java", "maven"}
	lm[JavaScript] = []string{"js", "node", "nodejs", "javascript", "npm", "vue", "react"}
	lm[Php] = []string{"php", "composer"}
	lm[Ruby] = []string{"ruby"}
	lm[Golang] = []string{"golang", "go", "gomod"}
	lm[Rust] = []string{"rust", "cargo"}
	for t, ls := range lm {
		for _, l := range ls {
			lanMap[l] = t
		}
	}
}

/**
 * @description: 获取最相似的语言
 * @param {string} language 语言
 * @return {language.Type} 语言类型
 */
func NewLanguage(language string) Type {
	if language == "" {
		return None
	}
	if l, ok := lanMap[strings.ToLower(language)]; ok {
		return l
	} else {
		return None
	}
}
