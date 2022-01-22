/*
 * @Descripation: mvn解析依赖树
 * @Date: 2021-12-16 10:10:13
 */

package java

import (
	"bytes"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

/**
 * @description: 调用mvn解析项目获取依赖树
 * @param {string} path 项目目录
 * @return {*srt.DepTree} 项目根节点
 */
func MvnDepTree(path string) (root *srt.DepTree) {
	root = srt.NewDepTree(nil)
	pwd, err := os.Getwd()
	if err != nil {
		logs.Error(err)
		return
	}
	os.Chdir(path)
	cmd := exec.Command("mvn", "dependency:tree", "--fail-never")
	out, _ := cmd.CombinedOutput()
	os.Chdir(pwd)
	// 统一替换换行符为\n
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\n\r"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\r"), []byte("\n"))
	// 获取mvn解析内容
	lines := strings.Split(string(out), "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "[INFO] ")
	}
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)
	// 记录依赖树起始位置行号
	start := 0
	// 标记是否在依赖范围内树
	tree := false
	// 获取mvn依赖树
	for i, line := range lines {
		if title.MatchString(line) {
			tree = true
			start = i
			continue
		}
		if tree && strings.Trim(line, "-") == "" {
			tree = false
			buildMvnDepTree(root, lines[start+1:i])
			continue
		}
	}
	return
}

/**
 * @description: 构建mvn树
 * @param {*srt.DepTree} root 依赖树根节点
 * @param {[]string} lines 依赖树信息
 */
func buildMvnDepTree(root *srt.DepTree, lines []string) {
	// 记录当前的顶点节点列表
	tops := []*srt.DepTree{root}
	// 上一层级
	lastLevel := -1
	for _, line := range lines {
		// 计算层级
		level := 0
		for line[level*3+2] == ' ' {
			level++
		}
		tops = tops[:len(tops)-lastLevel+level-1]
		root = tops[len(tops)-1]
		tags := strings.Split(line[level*3:], ":")
		if len(tags) < 4 {
			logs.Error(errors.New("mvn parse error"))
			break
		}
		dep := srt.NewDepTree(root)
		dep.Vendor = tags[0]
		dep.Name = tags[1]
		dep.Version = srt.NewVersion(tags[3])
		dep.Language = language.Java
		tops = append(tops, dep)
		lastLevel = level
	}
}
