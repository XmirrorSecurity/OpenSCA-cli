package groovy

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

// groovy 变量表
type Variable map[string]string

var (
	refReg = regexp.MustCompile(`((\w+)\[['"](\w+)['"]\])|(\$\{?[^{}"']*\}?)`)
)

// Replace 使用当前变量表中的变量替换文本中的变量值
func (v Variable) Replace(text string) string {

	if text == "" {
		return text
	}

	for exist := map[string]bool{}; !exist[text]; {
		exist[text] = true
		check := func(k, v string) bool {
			return len(v) > 0 && !strings.Contains(v, k)
		}
		text = refReg.ReplaceAllStringFunc(text, func(s string) string {
			if strings.HasPrefix(s, "$") {
				k := strings.Trim(s[1:], "{}")
				if value, ok := v[k]; ok {
					if check(s, value) {
						s = value
					}
				}
			} else {
				l := strings.Index(s, "[")
				if l > 0 {
					if value, ok := v[fmt.Sprintf("%s.%s", s[:l], s[l+2:len(s)-2])]; ok {
						if check(s, value) {
							s = value
						}
					}
				}
			}
			return s
		})
	}

	return text
}

var startReg = regexp.MustCompile(`\s*(\w+)\s*=?\s*[\[\{]`)
var varReg = regexp.MustCompile(`([\w]+)\s*[=:][\s\n]*['"]?([^\s()/'"]+)['"]?`)

// Scan 提取文件中的变量
func (v Variable) Scan(file *model.File) {

	if file == nil {
		return
	}

	var text string
	file.OpenReader(func(reader io.Reader) {
		data, _ := io.ReadAll(reader)
		text = string(data)
	})

	blockIndex := startReg.FindAllStringIndex(text, -1)
	for i, bi := range blockIndex {
		end := len(text)
		if i+1 < len(blockIndex) {
			end = blockIndex[i+1][0]
		}
		m := startReg.FindStringSubmatch(text[bi[0]:bi[1]])
		var object string
		if len(m) == 2 {
			object = m[1]
		}
		if object == "" {
			continue
		}
		for _, line := range strings.Split(text[bi[1]:end], "\n") {
			match := varReg.FindStringSubmatch(line)
			if len(match) == 3 {
				if object == "ext" {
					v[match[1]] = match[2]
				}
				v[fmt.Sprintf("%s.%s", object, match[1])] = match[2]
			}
		}
	}

}
