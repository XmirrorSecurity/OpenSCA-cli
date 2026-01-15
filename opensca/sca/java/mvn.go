package java

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/sca/cache"
)

// ParsePoms 解析一个项目中的pom文件
// poms: 项目中全部的pom文件列表
// exclusion: 不需要解析的pom文件
// call: 每个pom文件会解析成一个依赖图 返回对应的依赖图
func ParsePoms(ctx context.Context, poms []*Pom, exclusion []*Pom, call func(pom *Pom, root *model.DepGraph)) {

	// modules继承属性
	inheritModules(poms)

	// 记录当前项目的pom文件信息
	gavMap := map[string]*model.File{}
	PathMap := map[string]*model.File{}
	for _, pom := range poms {
		gavMap[pom.GAV()] = pom.File
		pom.Update(&pom.PomDependency)
		gavMap[pom.GAV()] = pom.File
		if pom.File.Relpath() != "" {
			PathMap[pom.File.Relpath()] = pom.File
		}
	}

	// 获取dependency对应的pom
	getpom := func(dep PomDependency, repos ...[]string) *Pom {
		// 通过gav查找pom
		f, ok := gavMap[dep.GAV()]
		// 通过relativaPath查找pom
		if !ok && dep.RelativePath != "" && dep.Define != nil && dep.Define.File.Relpath() != "" {
			pompath := filepath.Join(filepath.Dir(dep.Define.File.Relpath()), dep.RelativePath)
			f, ok = PathMap[pompath]
		}
		var p *Pom
		if ok {
			f.OpenReader(func(reader io.Reader) {
				p = ReadPom(reader)
				p.File = f
			})
		}
		if p != nil {
			return p
		}
		// 从组件仓库下载pom
		var rs []common.RepoConfig
		for _, urls := range repos {
			for _, url := range urls {
				rs = append(rs, common.RepoConfig{Url: url})
			}
		}
		p = mavenOrigin(dep.GroupId, dep.ArtifactId, dep.Version, rs...)

		if p == nil {
			logs.Warnf("not found pom %s", dep.Index3())
		}

		return p
	}

	exclusionMap := map[*Pom]bool{}
	for _, pom := range exclusion {
		exclusionMap[pom] = true
	}

	wg := sync.WaitGroup{}
	for _, pom := range poms {

		// 跳过不需要解析的pom
		if exclusionMap[pom] {
			continue
		}

		wg.Add(1)
		go func(pom *Pom) {
			defer wg.Done()
			call(pom, parsePom(ctx, pom, getpom))
		}(pom)
	}
	wg.Wait()
}

// inheritModules 继承modules属性
func inheritModules(poms []*Pom) {

	gavMap := map[string]bool{}
	for _, pom := range poms {
		gavMap[pom.GAV()] = true
	}

	// 记录pom继承关系
	_mod := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph { return &model.DepGraph{Name: s[0]} })
	for _, pom := range poms {
		n := _mod.LoadOrStore(pom.ArtifactId)
		n.Expand = pom
		// 记录modules继承关系
		for _, subMod := range pom.Modules {
			n.AppendChild(_mod.LoadOrStore(subMod))
		}
		// 记录parent继承关系
		if gavMap[pom.Parent.GAV()] {
			_mod.LoadOrStore(pom.Parent.ArtifactId).AppendChild(n)
		}
	}

	// 传递属性
	_mod.Range(func(k string, v *model.DepGraph) bool {

		// 跳过非根pom
		if len(v.Parents) > 0 {
			return true
		}

		// 从每个根pom开始遍历
		v.ForEachPath(func(p, n *model.DepGraph) bool {

			// Expand为空代表当前节点已传递属性
			if n.Expand == nil {
				return true
			}

			// 判断parent.Expand是否为空来判断parent是否已经传递过属性
			for _, p := range n.Parents {
				if p.Expand != nil {
					// 至少一个parent尚未传递属性则不处理当前节点
					return true
				}
			}

			// 获取当前pom
			pom, ok := n.Expand.(*Pom)
			if !ok {
				return true
			}

			// 置空Expand标记该节点已传递属性
			n.Expand = nil

			// 将属性传递给需要继承的pom
			for _, c := range n.Children {
				mod := _mod.LoadOrStore(c.Name)
				if mod.Expand == nil {
					continue
				}
				modpom, ok := mod.Expand.(*Pom)
				if !ok {
					continue
				}
				if modpom.Properties == nil {
					modpom.Properties = PomProperties{}
				}
				for k, v := range pom.Properties {
					if _, ok := modpom.Properties[k]; !ok {
						modpom.Properties[k] = v
					}
				}
			}

			return true
		})

		return true
	})
}

type getPomFunc func(dep PomDependency, repos ...[]string) *Pom

// inheritPom 继承pom所需内容
func inheritPom(pom *Pom, getpom getPomFunc) {

	// 记录统计过的parent 避免pom循环引用
	parentSet := map[string]bool{}

	// 继承parent
	parent := pom.Parent
	for parent.ArtifactId != "" {

		pom.Update(&parent)

		if parentSet[parent.Index3()] {
			break
		} else {
			parentSet[parent.Index3()] = true
		}

		parentPom := getpom(parent, pom.Repositories, pom.Mirrors)
		if parentPom == nil {
			break
		}
		parentPom.PomDependency = parent
		parent = parentPom.Parent

		// 继承properties
		for k, v := range parentPom.Properties {
			if _, ok := pom.Properties[k]; !ok {
				pom.Properties[k] = v
			}
		}

		// 继承dependencyManagement
		pom.DependencyManagement = append(pom.DependencyManagement, parentPom.DependencyManagement...)

		// 继承dependencies
		pom.Dependencies = append(pom.Dependencies, parentPom.Dependencies...)

		// 继承repo&mirror
		pom.Repositories = append(pom.Repositories, parentPom.Repositories...)
		pom.Mirrors = append(pom.Mirrors, parentPom.Mirrors...)

	}

	// 更新pom坐标
	pom.Update(&pom.PomDependency)
	pom.Update(&pom.Parent)

	// 删除重复依赖项
	depIndex2Set := map[string]bool{}
	for i := len(pom.Dependencies) - 1; i >= 0; i-- {
		dep := pom.Dependencies[i]
		if depIndex2Set[dep.Index2()] {
			pom.Dependencies = append(pom.Dependencies[:i], pom.Dependencies[i+1:]...)
		} else {
			depIndex2Set[dep.Index2()] = true
		}
	}

	// 处理dependencyManagement
	depIndex2Set = map[string]bool{}
	for i := 0; i < len(pom.DependencyManagement); {

		dep := pom.DependencyManagement[i]

		pom.Update(dep)

		// 去重 保留第一个声明
		if depIndex2Set[dep.Index2()] {
			pom.DependencyManagement = append(pom.DependencyManagement[:i], pom.DependencyManagement[i+1:]...)
			continue
		} else {
			i++
			depIndex2Set[dep.Index2()] = true
		}

		if dep.Scope != "import" {
			continue
		}

		// 引入scope为import的pom
		ipom := getpom(*dep, pom.Repositories, pom.Mirrors)
		if ipom == nil {
			continue
		}
		ipom.PomDependency = *dep

		// import引入的pom需要继承parent
		inheritPom(ipom, getpom)

		// 复制dependencyManagement内容
		for _, idep := range ipom.DependencyManagement {
			if depIndex2Set[idep.Index2()] {
				continue
			}
			// import的dependencyManagement使用自身pom属性而非根pom属性
			ipom.Update(idep)
			pom.DependencyManagement = append(pom.DependencyManagement, idep)
		}
	}
}

func replacePomDependency(old, new *PomDependency, indirect bool) (replaced *PomDependency) {
	originVersion := old.Version
	originScope := old.Scope
	dep := *new
	replaced = &dep
	// 间接依赖优先使用新的version
	if indirect && replaced.Version == "" {
		replaced.Version = originVersion
	}
	// 直接依赖优先保留原始scope
	if !indirect && originScope != "" {
		replaced.Scope = originScope
	}
	return
}

// parsePom 解析单个pom 返回该pom的依赖图
func parsePom(ctx context.Context, pom *Pom, getpom getPomFunc) *model.DepGraph {

	// 补全nil值
	if pom.Properties == nil {
		pom.Properties = PomProperties{}
	}

	pom.Update(&pom.PomDependency)

	// 继承pom
	inheritPom(pom, getpom)

	// 记录在根pom的dependencyManagement中非import组件信息
	rootPomManagement := map[string]*PomDependency{}
	for _, dep := range pom.DependencyManagement {
		if dep.Scope != "import" {
			rootPomManagement[dep.Index2()] = dep
		}
	}

	root := &model.DepGraph{Vendor: pom.GroupId, Name: pom.ArtifactId, Version: pom.Version, Path: pom.File.Relpath()}
	root.Expand = pom

	// 解析子依赖构建依赖关系
	depIndex2Set := map[string]bool{}
	root.ForEachNode(func(p, n *model.DepGraph) bool {

		select {
		case <-ctx.Done():
			return false
		default:
		}

		if n.Expand == nil {
			return true
		}

		np := n.Expand.(*Pom)

		for _, lic := range np.Licenses {
			n.AppendLicense(lic)
		}

		// 记录在当前pom的dependencyManagement中非import组件信息
		depManagement := map[string]*PomDependency{}
		for _, dep := range np.DependencyManagement {
			if dep.Scope != "import" {
				depManagement[dep.Index2()] = dep
			}
		}

		for _, dep := range np.Dependencies {

			// 丢弃provided或optional=true的组件
			if dep.Scope == "provided" || dep.Optional {
				continue
			}
			// 丢弃scope为test的间接依赖
			if np != pom && dep.Scope == "test" {
				continue
			}

			// 非根pom直接引入的依赖先用自身pom的dependencyManament检查是否需要排除
			if np != pom {
				if d, ok := depManagement[dep.Index2()]; ok {
					if d.Optional ||
						d.Scope == "provided" ||
						d.Scope == "test" {
						continue
					}
				}
			}

			np.Update(dep)

			// 使用当前pom的dependencyManagement补全
			if d, ok := depManagement[dep.Index2()]; ok {
				exclusion := append(dep.Exclusions, d.Exclusions...)
				if dep.Version == "" {
					dep = replacePomDependency(dep, d, false)
				}
				dep.Exclusions = exclusion
				np.Update(dep)
			}

			// 非根pom直接引入的依赖 或者组件版本号为空 需要再次使用根pom的dependencyManagement补全
			if np != pom || dep.Version == "" {
				d, ok := rootPomManagement[dep.Index2()]
				if ok {
					exclusion := append(dep.Exclusions, d.Exclusions...)
					dep = replacePomDependency(dep, d, true)
					dep.Exclusions = exclusion
					pom.Update(dep)
				}
			}

			// 查看是否在Exclusion列表中
			if np.NeedExclusion(*dep) {
				continue
			}

			// 保留先声明的组件
			if depIndex2Set[dep.Index2()] {
				continue
			}
			depIndex2Set[dep.Index2()] = true

			if dep.Check() {
				logs.Debugf("find %s", dep.ImportPathStack())
			} else {
				logs.Warnf("find invalid %s", dep.ImportPathStack())
				continue
			}

			sub := &model.DepGraph{Vendor: dep.GroupId, Name: dep.ArtifactId, Version: dep.Version}
			sub.Develop = dep.Scope == "test"

			if subpom := getpom(*dep, np.Repositories, np.Mirrors); subpom != nil {
				subpom.PomDependency = *dep
				// 继承根pom的exclusion
				subpom.Exclusions = append(subpom.Exclusions, np.Exclusions...)
				// 依赖继承parent
				inheritPom(subpom, getpom)
				sub.Expand = subpom
			}

			n.AppendChild(sub)
		}

		return true
	})

	return root
}

var mavenOrigin = func(groupId, artifactId, version string, repos ...common.RepoConfig) *Pom {

	var p *Pom

	path := cache.Path(groupId, artifactId, version, model.Lan_Java)
	cache.Load(path, func(reader io.Reader) {
		p = ReadPom(reader)
	})

	if p != nil {
		return p
	}

	DownloadPomFromRepo(PomDependency{GroupId: groupId, ArtifactId: artifactId, Version: version}, func(r io.Reader) {
		data, err := io.ReadAll(r)
		if err != nil {
			logs.Warn(err)
			return
		}
		reader := bytes.NewReader(data)
		p = ReadPom(reader)
		reader.Seek(0, io.SeekStart)
		cache.Save(path, reader)
	}, repos...)

	return p
}

// RegisterMavenOrigin 注册maven数据源
// origin: 获取数据源 gav=>pom
func RegisterMavenOrigin(origin func(groupId, artifactId, version string) *Pom) {
	if origin != nil {
		mavenOrigin = func(groupId, artifactId, version string, repos ...common.RepoConfig) *Pom {
			return origin(groupId, artifactId, version)
		}
	}
}

// DownloadPomFromRepo 从maven仓库下载pom
// dep: pom的dependency内容
// do: 对http.Response.Body的操作
// repos: 额外使用的maven仓库
func DownloadPomFromRepo(dep PomDependency, do func(r io.Reader), repos ...common.RepoConfig) {

	if !dep.Check() {
		return
	}

	// 正式版本
	pom := fmt.Sprintf("%s/%s/%s/%s-%s.pom", strings.ReplaceAll(dep.GroupId, ".", "/"), dep.ArtifactId, dep.Version, dep.ArtifactId, dep.Version)
	common.DownloadUrlFromRepos(pom, func(repo common.RepoConfig, r io.Reader) { do(r) }, append(defaultMavenRepo, repos...)...)

	// 快照版本
	if !strings.HasSuffix(strings.ToLower(dep.Version), "-snapshot") {
		return
	}
	snap := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", strings.ReplaceAll(dep.GroupId, ".", "/"), dep.ArtifactId, dep.Version)
	common.DownloadUrlFromRepos(snap, func(repo common.RepoConfig, r io.Reader) {

		metadata := struct {
			LastTime     string `xml:"versioning>lastUpdated"`
			SnapVersions []struct {
				Version string `xml:"value"`
				Time    string `xml:"updated"`
			} `xml:"versioning>snapshotVersions>snapshotVersion"`
		}{}

		err := xml.NewDecoder(r).Decode(&metadata)
		if err != nil {
			logs.Warn(err)
		}

		if metadata.LastTime == "" {
			return
		}

		for _, snap := range metadata.SnapVersions {
			if snap.Time == metadata.LastTime {
				snapom := fmt.Sprintf("%s/%s/%s/%s-%s.pom", strings.ReplaceAll(dep.GroupId, ".", "/"), dep.ArtifactId, snap.Version, dep.ArtifactId, snap.Version)
				common.DownloadUrlFromRepos(snapom, func(repo common.RepoConfig, r io.Reader) { do(r) }, repo)
				break
			}
		}

	}, append(defaultMavenRepo, repos...)...)

}

// MvnTree 调用mvn dependency:tree解析依赖
// pom: pom文件信息
func MvnTree(ctx context.Context, pom *Pom) *model.DepGraph {

	if !config.Conf().Optional.Dynamic {
		return nil
	}

	if pom == nil {
		return nil
	}

	if _, err := exec.LookPath("mvn"); err != nil {
		return nil
	}

	cmd := exec.CommandContext(ctx, "mvn", "dependency:tree")
	cmd.Dir = filepath.Dir(pom.File.Abspath())
	output, err := cmd.CombinedOutput()
	if err != nil {
		// logs.Warn(err)
		return nil
	}

	// 记录当前处理的依赖树数据
	var lines []string
	// 标记是否在依赖范围内树
	tree := false
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)

	scan := bufio.NewScanner(bytes.NewBuffer(output))
	for scan.Scan() {
		line := strings.TrimPrefix(scan.Text(), "[INFO] ")
		if title.MatchString(line) {
			tree = true
			continue
		}
		if tree && strings.Trim(line, "-") == "" {
			tree = false
			root := parseMvnTree(lines)
			if root != nil && root.Name == pom.ArtifactId {
				root.Path = pom.File.Relpath()
				return root
			}
			lines = nil
			continue
		}
		if tree {
			lines = append(lines, line)
			continue
		}
	}

	return nil
}

// parseMvnTree 解析 mvn dependency:tree 的输出
func parseMvnTree(lines []string) *model.DepGraph {

	// 记录当前的顶点节点列表
	var tops = []*model.DepGraph{}
	// 上一层级
	lastLevel := -1

	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	for _, line := range lines {

		// 计算层级
		level := 0
		for level*3+2 < len(line) && line[level*3+2] == ' ' {
			level++
		}
		if level*3+2 >= len(line) {
			continue
		}

		if level-lastLevel > 1 {
			// 在某个依赖解析失败的时候 子依赖会出现这种情况
			continue
		}

		tags := strings.Split(line[level*3:], ":")
		if len(tags) < 4 {
			continue
		}

		dep := _dep(tags[0], tags[1], tags[3])

		if dep == nil {
			continue
		}

		scope := tags[len(tags)-1]
		if scope == "test" || scope == "provided" {
			dep.Develop = true
		}

		if level > 0 {
			tops[level-1].AppendChild(dep)
		}

		tops = append(tops[:level], dep)

		lastLevel = level
	}

	if len(tops) > 0 {
		return tops[0]
	} else {
		return nil
	}
}
