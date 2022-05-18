/*
 * @Descripation: java Analyzer
 * @Date: 2021-11-03 11:17:09
 */
package java

import (
	"path"
	"sort"
	"strings"
	"util/bar"
	"util/enum/language"
	"util/filter"
	"util/model"
)

type Analyzer struct {
	mvn *Mvn
	// maven仓库地址
	repos map[int64][]string
}

// New 创建java解析器
func New() Analyzer {
	return Analyzer{
		mvn:   NewMvn(),
		repos: map[int64][]string{},
	}
}

// GetLanguage Get language of Analyzer
func (Analyzer) GetLanguage() language.Type {
	return language.Java
}

// CheckFile Check if it is a parsable file
func (Analyzer) CheckFile(filename string) bool {
	return filter.JavaPom(filename)
}

// pomTree pom文件树
type pomTree struct {
	// 当前包内的文件列表
	poms []*Pom
	// 子pom树
	subTree map[string]*pomTree
}

// buildJarTree 构建jar树
func buildJarTree(jarMap map[string]*model.DepTree) []*model.DepTree {
	deps := []*model.DepTree{}
	paths := []string{}
	for path := range jarMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for i, path := range paths {
		jar, ok := jarMap[path]
		for ok {
			// 找到上一个jar包
			index := strings.LastIndex(path, ".jar/")
			if index != -1 {
				parentPath := path[:index+4]
				// 检测上一个jar包是否有依赖信息
				if parent, exist := jarMap[parentPath]; exist {
					// 有依赖信息则将当前jar包添加为父jar包的子依赖
					parent.Children = append(parent.Children, jar)
					jar.Parent = parent
					// 从paths中删除当前jar路径用来标识该jar包已有父节点
					paths[i] = ""
					// 跳出
					break
				} else {
					// 无依赖信息则尝试找再上一层jar包
					path = parentPath
					continue
				}
			} else {
				// 无父jar包跳出
				break
			}
		}
	}
	// 将无父节点的jar包作为依赖树根返回
	for _, path := range paths {
		if path != "" {
			if jar, exist := jarMap[path]; exist {
				deps = append(deps, jar)
			}
		}
	}
	return deps
}

// buildPomTree 构建pom树
func buildPomTree(poms []*Pom) *pomTree {
	tree := &pomTree{}
	for _, f := range poms {
		cur := tree
		// 提取路径中的压缩包
		// 例如: "lib/a.zip/b.jar/d2/pom.xml"
		// 结果: ["lib/a.zip","b.jar"]
		dirs := strings.Split(f.Path, "/")
		pkgs := []string{}
		last := 0
		for i := 0; i < len(dirs); i++ {
			if dirs[i] == "" || filter.AllPkg(dirs[i]) {
				pkgs = append(pkgs, path.Join(dirs[last:i+1]...))
				last = i + 1
			}
		}
		// 按包层级构建文件树
		for _, pkg := range pkgs {
			if cur.subTree == nil {
				cur.subTree = map[string]*pomTree{}
			}
			if _, exist := cur.subTree[pkg]; !exist {
				cur.subTree[pkg] = &pomTree{}
			}
			cur = cur.subTree[pkg]
		}
		// 向当前文件树中添加文件
		cur.poms = append(cur.poms, f)
	}
	return tree
}

// parsePomTree 解析pom树
func (pt *pomTree) parsePomTree(jarMap map[string]*model.DepTree) []*model.DepTree {
	deps := []*model.DepTree{}
	mvn := NewMvn()
	for _, p := range pt.poms {
		mvn.AppendPom(p)
	}
	for _, p := range mvn.MvnSimulation() {
		d := model.NewDepTree(nil)
		d.Path = p.Path
		buildTree(p, d)
		deps = append(deps, d)
	}
	for pkg, subTree := range pt.subTree {
		subPoms := subTree.parsePomTree(jarMap)
		if jar, exist := jarMap[pkg]; exist {
			for _, pom := range subPoms {
				if pom.Name == jar.Name && pom.Version == jar.Version {
					// 如果组件名和版本相同则移除jar
					jar.Dependency = pom.Dependency
					for _, c := range pom.Children {
						c.Parent = jar
					}
					jar.Children = append(jar.Children, pom.Children...)
					jar.Path = pom.Path
				} else {
					pom.Parent = jar
					jar.Children = append(jar.Children, pom)
				}
			}
		} else {
			jar = model.NewDepTree(nil)
			jar.Path = pkg
			jar.Children = subPoms
			for _, pom := range subPoms {
				pom.Parent = jar
			}
			deps = append(deps, jar)
		}
	}
	return deps
}

// ParseFiles 解析一组文件，结果只返回依赖树的根，如果解析出多个依赖树则返回多个根
func (a Analyzer) ParseFiles(files []*model.FileInfo) (deps []*model.DepTree) {
	deps = []*model.DepTree{}
	// 通过jar包解析出的依赖 map[jarpath]depTree
	jarMap := map[string]*model.DepTree{}
	// pom 文件列表
	poms := []*Pom{}
	for _, f := range files {
		// 读取pom文件
		if filter.JavaPom(f.Name) {
			p := ReadPom(f.Data)
			p.Path = f.Name
			poms = append(poms, p)
		}
	}
	// 构建jar树
	deps = buildJarTree(jarMap)
	// 构建pom树
	deps = append(deps, buildPomTree(poms).parsePomTree(jarMap)...)
	return
}

// buildTree 构建依赖树
func buildTree(p *Pom, root *model.DepTree) {
	dep := model.NewDepTree(root)
	dep.Vendor = p.GroupId
	dep.Name = p.ArtifactId
	dep.Version = model.NewVersion(p.Version)
	bar.Dependency.Add(1)
	for _, dp := range p.DependenciesPom {
		bar.Maven.Add(1)
		buildTree(dp, dep)
	}
}
