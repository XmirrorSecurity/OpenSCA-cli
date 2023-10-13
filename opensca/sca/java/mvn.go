package java

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
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

	for _, pom := range poms {

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 提过不需要解析的pom
		if exclusionMap[pom] {
			continue
		}

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

				// 间接依赖先用自身pom的dependencyManament检查是否需要排除
				if np != pom {
					if d, ok := depManagement[dep.Index2()]; ok {
						if d.Optional ||
							d.Scope == "provided" ||
							d.Scope == "test" {
							continue
						}
					}
				}

				// 间接依赖优先通过dependencyManagement补全
				if np != pom || dep.Version == "" {
					d, ok := rootPomManagement[dep.Index2()]
					if !ok {
						d, ok = depManagement[dep.Index2()]
					}
					if ok {
						// exclusion 需要保留
						exclusion := append(dep.Exclusions, d.Exclusions...)
						dep = d
						dep.Exclusions = exclusion
						pom.Update(dep)
					}
				}

				np.Update(dep)

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
					// 子依赖继承自身pom
					inheritPom(subpom, getpom)
					sub.Expand = subpom
				}

				n.AppendChild(sub)
			}

			return true
		})

		if root.Name != "" {
			call(pom, root)
		}
	}
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

			// 判断parent是否有expand来判断是否已经继承过属性
			expand := false
			for _, p := range n.Parents {
				if p.Expand != nil {
					expand = true
					return true
				}
			}

			// 至少一个parent尚未继承属性则暂不处理当前节点
			if expand {
				return true
			}

			if n.Expand == nil {
				return true
			}

			// 获取当前pom
			pom, ok := n.Expand.(*Pom)
			if !ok {
				return true
			}

			// 删除expand标识已继承属性
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

	// 继承parent
	parent := pom.Parent
	for parent.ArtifactId != "" {

		pom.Update(&parent)

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
		if p == nil {
			return
		}

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

	repoSet := map[string]bool{}

	for _, repo := range append(defaultMavenRepo, repos...) {

		if repo.Url == "" {
			continue
		}
		if repoSet[repo.Url] {
			continue
		}
		repoSet[repo.Url] = true

		url := fmt.Sprintf("%s/%s/%s/%s/%s-%s.pom", strings.TrimRight(repo.Url, "/"),
			strings.ReplaceAll(dep.GroupId, ".", "/"), dep.ArtifactId, dep.Version,
			dep.ArtifactId, dep.Version)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logs.Warn(err)
			continue
		}
		if repo.Username+repo.Password != "" {
			req.SetBasicAuth(repo.Username, repo.Password)
		}

		resp, err := common.HttpClient.Do(req)
		if err != nil {
			logs.Warn(err)
			continue
		}

		if resp.StatusCode != 200 {
			logs.Warnf("%d %s", resp.StatusCode, url)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			continue
		} else {
			logs.Debugf("%d %s", resp.StatusCode, url)
			do(resp.Body)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			break
		}
	}
}

// MvnTree 调用mvn dependency:tree解析依赖
// pom: pom文件信息
func MvnTree(ctx context.Context, pom *Pom) *model.DepGraph {

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
