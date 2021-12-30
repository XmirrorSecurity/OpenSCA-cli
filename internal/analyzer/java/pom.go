/*
 * @Descripation: 解析Pom文件相关
 * @Date: 2021-11-04 09:40:33
 */

package java

import (
	"encoding/xml"
	"fmt"
	"io"
	"opensca/internal/srt"
	"regexp"
	"strings"
)

var (
	// 存储property信息 map[dirpath]map[property-key]property-value
	properties = map[string]map[string]string{}
)

// Pom文件
type PomXml struct {
	PomDependency
	Parent               PomDependency   `xml:"parent"`
	Dependencies         []PomDependency `xml:"dependencies>dependency"`
	DependencyManagement []PomDependency `xml:"dependencyManagement>dependencies>dependency"`
	Repositories         []string        `xml:"repositories>repository>url"`
	Licenses             []string        `xml:"licenses>license>name"`
}

// Pom文件Dependency标签
type PomDependency struct {
	ArtifactId string          `xml:"artifactId"`
	GroupId    string          `xml:"groupId"`
	Version    string          `xml:"version"`
	Scope      string          `xml:"scope"`
	Optional   bool            `xml:"optional"`
	Exclusions []PomDependency `xml:"exclusions>exclusion"`
}

/**
 * @description: 解析pom.xml文件
 * @param {string} dirpath 当前所在目录树路径
 * @param {[]byte} data 文件数据
 * @param {bool} isimport 是否是import解析的组件
 * @return {*PomXml} *PomXml结构
 */
func (a Analyzer) parsePomXml(dirpath string, data []byte, isimport bool) *PomXml {

	// 检测是否是无效数据
	invalid := func(v string) bool {
		return strings.Contains(v, "$") || len(v) == 0
	}

	properties = a.properties

	// 解析pom.xml
	pom := &PomXml{}
	data = regexp.MustCompile(`xml version="1.0" encoding="\S+"`).ReplaceAll(data, []byte(`xml version="1.0" encoding="UTF-8"`))
	err := xml.Unmarshal(data, pom)
	if err != nil && err != io.EOF {
		return pom
	}
	if invalid(pom.GroupId) && pom.Parent.GroupId != "" {
		pom.GroupId = pom.Parent.GroupId
	}
	if invalid(pom.Version) && pom.Parent.Version != "" {
		pom.Version = pom.Parent.Version
	}

	// 解析property
	if _, ok := properties[dirpath]; !ok {
		properties[dirpath] = map[string]string{}
	}
	// 当前目录的property
	dirProperty := properties[dirpath]
	propertiesTag := regexp.MustCompile(`<properties>([\s\S\n]*)</properties>`).FindSubmatch(data)
	if len(propertiesTag) == 2 {
		for _, tags := range regexp.MustCompile(`<(\S+)>(\S+)<\S+>`).FindAllSubmatch(propertiesTag[1], -1) {
			dirProperty[string(tags[1])] = string(tags[2])
		}
	}

	// 当前property集合
	property := map[string]string{}
	// 获取parent的property
	parent := pom.Parent
	if _, ok := a.poms[dirpath]; !ok {
		a.poms[dirpath] = map[string]struct{}{}
	}
	for parent.ArtifactId != "" && parent.GroupId != "" && parent.Version != "" {
		key := strings.ToLower(fmt.Sprintf("%s:%s:%s", parent.GroupId, parent.ArtifactId, parent.Version))
		if _, ok := a.poms[dirpath][key]; !ok {
			a.poms[dirpath][key] = struct{}{}
			p := a.getpom(srt.Dependency{Vendor: parent.GroupId, Name: parent.ArtifactId, Version: srt.NewVersion(parent.Version)}, dirpath, pom.Repositories, false)
			if p != nil {
				parent = p.Parent
			}
		} else {
			break
		}
	}
	// 计算根目录到当前目录的所有property并集
	paths := strings.Split(dirpath, "/")
	for i := range paths {
		// 必须保证新属性覆盖旧属性，因此从根目录向当前目录遍历
		if m, ok := properties[strings.Join(paths[:i+1], "/")]; ok {
			for k, v := range m {
				property[k] = v
			}
		}
	}

	// 存储属性值
	getValue := func(key string) string {
		return regexp.MustCompile(`\$\{[^{}]*\}`).ReplaceAllStringFunc(key,
			func(s string) string {
				exist := map[string]struct{}{}
				// 递归搜索版本号
				for strings.HasPrefix(s, "$") {
					if _, ok := exist[s]; ok {
						break
					} else {
						exist[s] = struct{}{}
					}
					k := s[2 : len(s)-1]
					if v, ok := property[k]; ok {
						if len(v) > 0 {
							s = v
						} else {
							break
						}
					} else {
						break
					}
				}
				return s
			})
	}

	// 检查当前pom的无效数据
	if groupId, ok := property["groupId"]; ok && len(pom.GroupId) == 0 {
		pom.GroupId = groupId
	}
	if version, ok := property["version"]; ok && len(pom.Version) == 0 {
		pom.Version = version
	}
	if invalid(pom.GroupId) {
		pom.GroupId = getValue(pom.GroupId)
	}
	if invalid(pom.ArtifactId) {
		pom.ArtifactId = getValue(pom.ArtifactId)
	}
	if invalid(pom.Version) {
		pom.Version = getValue(pom.Version)
	}

	// 设置属性
	setProperty := func(key, value string) {
		if value != "" {
			property[key] = value
		}
	}

	// 模拟预定义property
	setProperty("pom.groupId", pom.Parent.GroupId)
	setProperty("pom.artifactId", pom.Parent.ArtifactId)
	setProperty("pom.version", pom.Parent.Version)
	setProperty("parent.groupId", pom.Parent.GroupId)
	setProperty("parent.artifactId", pom.Parent.ArtifactId)
	setProperty("parent.version", pom.Parent.Version)
	setProperty("project.parent.groupId", pom.Parent.GroupId)
	setProperty("project.parent.artifactId", pom.Parent.ArtifactId)
	setProperty("project.parent.version", pom.Parent.Version)
	setProperty("project.groupId", pom.GroupId)
	setProperty("project.artifactId", pom.ArtifactId)
	setProperty("project.version", pom.Version)

	// 存储DependencyManagement的值
	for _, dep := range pom.DependencyManagement {
		ver := getValue(dep.Version)
		property[fmt.Sprintf("%s|%s", dep.GroupId, dep.ArtifactId)] = ver
		a.properties[dirpath][fmt.Sprintf("%s|%s", dep.GroupId, dep.ArtifactId)] = ver
		// 检查scope标签是否为import
		if dep.Scope == "import" {
			d := srt.NewDependency()
			d.Vendor = dep.GroupId
			d.Name = dep.ArtifactId
			d.Version = srt.NewVersion(ver)
			a.getpom(d, dirpath, pom.Repositories, true)
		}
	}

	// 修正Dependencies的无效数据
	for i, dep := range pom.Dependencies {
		if invalid(dep.GroupId) {
			pom.Dependencies[i].GroupId = getValue(dep.GroupId)
		}
		if invalid(dep.ArtifactId) {
			pom.Dependencies[i].ArtifactId = getValue(dep.ArtifactId)
		}
		if invalid(dep.Version) {
			pom.Dependencies[i].Version = getValue(dep.Version)
			if invalid(pom.Dependencies[i].Version) {
				if v, ok := property[fmt.Sprintf("%s|%s", pom.Dependencies[i].GroupId, pom.Dependencies[i].ArtifactId)]; ok {
					pom.Dependencies[i].Version = v
				}
			}
		}
	}
	return pom
}

/**
 * @description: 解析pom.properties文件
 * @param {[]byte} data 文件数据
 */
func (a Analyzer) parsePomProperties(dirpath string, data []byte) {
	properties = a.properties
	if _, ok := properties[dirpath]; !ok {
		properties[dirpath] = map[string]string{}
	}
	for _, match := range regexp.MustCompile(`(\S+)=(\S+)`).FindAllSubmatch(data, -1) {
		properties[dirpath][string(match[1])] = string(match[2])
	}
}
