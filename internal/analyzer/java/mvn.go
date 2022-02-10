/*
 * @Descripation: mvn解析依赖树
 * @Date: 2021-12-16 10:10:13
 */

package java

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"opensca/internal/cache"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"os"
	"os/exec"
	"path"
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

/**
 * @description: 下载pom文件
 * @param {srt.Dependency} dep 依赖信息
 * @param {...string} repos 仓库地址
 * @return {[]byte} pom文件数据
 * @return {error} 错误
 */
func downloadPom(dep srt.Dependency, repos ...string) (data []byte, err error) {
	if repos == nil {
		repos = []string{}
	}
	tags := strings.Split(dep.Vendor, ".")
	tags = append(tags, dep.Name)
	tags = append(tags, dep.Version.Org)
	tags = append(tags, fmt.Sprintf("%s-%s.pom", dep.Name, dep.Version.Org))
	// 遍历仓库地址, 默认maven仓库
	for i, repo := range append(repos, `https://repo.maven.apache.org/maven2/`) {
		// 是否是最后一个仓库(默认的maven仓库)
		last := i == len(repos)
		// 拼接完整的pom下载地址
		url := repo + strings.Join(tags, "/")
		if rep, err := http.Get(url); err != nil {
			if last {
				return nil, err
			} else {
				continue
			}
		} else {
			defer rep.Body.Close()
			if rep.StatusCode != 200 {
				if last {
					return ioutil.ReadAll(rep.Body)
				} else {
					continue
				}
			} else {
				return ioutil.ReadAll(rep.Body)
			}
		}
	}
	// 应该走不到这里
	return nil, fmt.Errorf("download failure")
}

/**
 * @description: 获取pom文件
 * @param {srt.Dependency} dep 依赖组件信息
 * @param {string} dirpath 当前依赖所在目录路径
 * @param {[]string} repos mvn仓库地址
 * @param {bool} isimport 是否是import组件
 * @return {*PomXml} Pom结构
 */
func (a Analyzer) getpom(dep srt.Dependency, dirpath string, repos []string, isimport bool) *PomXml {
	if dep.Vendor == "" || dep.Name == "" || !dep.Version.Ok() {
		return nil
	}
	dep.Language = language.Java
	data := cache.LoadCache(dep)
	if len(data) != 0 {
		return a.parsePomXml(dirpath, data, isimport)
	} else {
		// 无本地缓存下载pom文件
		if data, err := downloadPom(dep, repos...); err == nil {
			// 保存pom文件
			cache.SaveCache(dep, data)
			return a.parsePomXml(dirpath, data, isimport)
		} else {
			logs.Warn(err)
		}
	}
	return nil
}

/**
 * @description: 获取子依赖
 * @param {*srt.DepTree} root 依赖组件信息
 */
func (a Analyzer) mavenSimulation(node *srt.DepTree) []*srt.DepTree {
	deps := []*srt.DepTree{}
	// 获取mvn仓库地址
	repos := []string{}
	if node.Parent != nil {
		if reps, ok := a.repos[node.Parent.ID]; ok {
			repos = reps
		}
	}
	// 获取组件pom文件并解析子依赖
	pom := a.getpom(node.Dependency, path.Dir(node.Path), repos, false)
	if pom != nil {
		for _, dep := range pom.Dependencies {
			if dep.Scope == "test" || dep.Scope == "provided" || dep.Optional {
				continue
			}
			// 检查组件信息是否完整
			ver := srt.NewVersion(dep.Version)
			if dep.GroupId == "" || dep.ArtifactId == "" || !ver.Ok() {
				continue
			}
			sub := srt.NewDepTree(node)
			sub.Vendor = dep.GroupId
			sub.Name = dep.ArtifactId
			sub.Version = ver
			deps = append(deps, sub)
			for _, exclusion := range dep.Exclusions {
				key := strings.ToLower(fmt.Sprintf("%s+%s", exclusion.GroupId, exclusion.ArtifactId))
				sub.Exclusions[key] = struct{}{}
			}
		}
	}
	return deps
}
