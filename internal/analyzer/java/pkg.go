/*
 * @Descripation: 供外部使用的方法
 * @Date: 2021-11-17 10:09:30
 */

package java

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"opensca/internal/bar"
	"opensca/internal/cache"
	"opensca/internal/enum/language"
	"opensca/internal/logs"
	"opensca/internal/srt"
	"path"
	"strings"
)

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
func (a Analyzer) ParseSubDependencies(root *srt.DepTree) {
	exist := map[string]struct{}{}
	queue := srt.NewQueue()
	queue.Push(root)
	for !queue.Empty() {
		node := queue.Pop().(*srt.DepTree)
		// java组件尝试获取子依赖
		if node.Language == language.Java {
			// 记录当前节点的子依赖
			childs := map[string]struct{}{}
			for _, child := range node.Children {
				childs[strings.ToLower(child.Name)] = struct{}{}
			}
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
					// 已在当前子依赖中则不添加
					if _, ok := childs[strings.ToLower(dep.ArtifactId)]; ok {
						continue
					}
					// 检查组件信息是否完整
					ver := srt.NewVersion(dep.Version)
					if dep.GroupId == "" || dep.ArtifactId == "" || !ver.Ok() {
						continue
					}
					bar.Maven.Add(1)
					sub := srt.NewDepTree(node)
					sub.Vendor = dep.GroupId
					sub.Name = dep.ArtifactId
					sub.Version = ver
					sub.Language = language.Java
					for _, exclusion := range dep.Exclusions {
						key := strings.ToLower(fmt.Sprintf("%s+%s", exclusion.GroupId, exclusion.ArtifactId))
						sub.Exclusions[key] = struct{}{}
					}
					sub.Path = path.Join(node.Path, sub.Dependency.String())
				}
			}
		}
		// 子依赖中未搜索过的组件入队
		for _, child := range node.Children {
			key := strings.ToLower(fmt.Sprintf("%s|%s", child.Name, child.Version.Org))
			if _, ok := exist[key]; !ok {
				exist[key] = struct{}{}
				queue.Push(child)
			}
		}
	}
	root.Exclusion()
}
