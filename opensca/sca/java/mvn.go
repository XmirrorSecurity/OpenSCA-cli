package java

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/cache"
)

func ParsePoms(poms []*Pom) []*model.DepGraph {

	// 记录module信息
	modules := map[string]*Pom{}
	for _, pom := range poms {
		modules[filepath.Base(filepath.Dir(pom.File.Path()))] = pom
	}

	// 将revision主动推送到所有modules
	for _, pom := range poms {
		if revision, ok := pom.Properties["revision"]; ok {
			for _, name := range pom.Modules {
				if p, ok := modules[name]; ok {
					if _, ok := p.Properties["revision"]; !ok {
						p.Properties["revision"] = revision
					}
				}
			}
		}
	}

	// 记录当前项目的pom文件信息
	gavMap := map[string]*model.File{}
	for _, pom := range poms {
		pom.Update(&pom.PomDependency)
		gavMap[pom.GAV()] = pom.File
	}

	// 获取对应的pom信息
	getpom := func(dep PomDependency, repos ...[]MvnRepo) *Pom {
		var p *Pom
		if f, ok := gavMap[dep.GAV()]; ok {
			f.OpenReader(func(reader io.Reader) {
				p = ReadPom(reader)
			})
		}
		if p != nil {
			return p
		}
		var rs []MvnRepo
		for _, repo := range repos {
			rs = append(rs, repo...)
		}
		return mavenOrigin(dep.GroupId, dep.ArtifactId, dep.Version, rs...)
	}

	var roots []*model.DepGraph

	for _, pom := range poms {

		// 补全nil值
		if pom.Properties == nil {
			pom.Properties = PomProperties{}
		}

		pom.Update(&pom.PomDependency)

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
		depManagement := map[string]*PomDependency{}
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
				depManagement[dep.Index2()] = dep
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

		_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
			return &model.DepGraph{
				Vendor:  s[0],
				Name:    s[1],
				Version: s[2],
			}
		}).LoadOrStore

		root := &model.DepGraph{Vendor: pom.GroupId, Name: pom.ArtifactId, Version: pom.Version, Path: pom.File.Path()}
		root.Expand = pom

		// 解析子依赖构建依赖关系
		depIndex2Set = map[string]bool{}
		root.ForEachNode(func(p, n *model.DepGraph) bool {

			if n.Expand == nil {
				return true
			}

			np := n.Expand.(*Pom)

			for _, lic := range np.Licenses {
				n.AppendLicense(lic)
			}

			for _, dep := range np.Dependencies {

				if dep.Scope == "provided" || dep.Optional {
					continue
				}

				// 丢弃子依赖的test依赖
				if np != pom && dep.Scope == "test" {
					continue
				}

				// 间接依赖优先通过dependencyManagement补全
				if np != pom || dep.Version == "" {
					if d, ok := depManagement[dep.Index2()]; ok {
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

				logs.Debugf("find %s", dep.ImportPathStack())

				sub := _dep(dep.GroupId, dep.ArtifactId, dep.Version)

				if sub.Expand != nil {
					if dep.Scope != "test" {
						sub.Develop = false
					}
					continue
				}

				subpom := getpom(*dep, np.Repositories, np.Mirrors)
				if subpom == nil {
					continue
				}

				subpom.PomDependency = *dep
				// 继承根pom的exclusion
				subpom.Exclusions = append(subpom.Exclusions, np.Exclusions...)

				// 子依赖继承自身parent属性
				subParent := subpom.Parent
				for subParent.ArtifactId != "" {
					subParentPom := getpom(subParent, np.Repositories, np.Mirrors)
					if subParentPom == nil {
						break
					}
					for k, v := range subParentPom.Properties {
						if _, ok := subpom.Properties[k]; !ok {
							subpom.Properties[k] = v
						}
					}
					subParent = subParentPom.Parent
				}

				sub.Expand = subpom
				sub.Develop = dep.Scope == "test"
				n.AppendChild(sub)
			}

			return true
		})

		roots = append(roots, root)
	}

	return roots
}

var mavenOrigin = func(groupId, artifactId, version string, repos ...MvnRepo) *Pom {

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

func RegisterMavenOrigin(origin func(groupId, artifactId, version string) *Pom) {
	if origin != nil {
		mavenOrigin = func(groupId, artifactId, version string, repos ...MvnRepo) *Pom {
			return origin(groupId, artifactId, version)
		}
	}
}

var httpClient = http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        500,
		MaxConnsPerHost:     500,
		MaxIdleConnsPerHost: 500,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	Timeout: 10 * time.Second,
}

func DownloadPomFromRepo(dep PomDependency, do func(r io.Reader), repos ...MvnRepo) {

	if !dep.Check() {
		return
	}

	repoSet := map[string]bool{}

	for _, repo := range append(defaultRepo, repos...) {

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

		resp, err := httpClient.Do(req)
		if err != nil {
			logs.Warn(err)
			continue
		}

		defer resp.Body.Close()
		defer io.Copy(io.Discard, resp.Body)

		if resp.StatusCode != 200 {
			logs.Warnf("%d %s", resp.StatusCode, url)
			continue
		} else {
			logs.Debugf("%d %s", resp.StatusCode, url)
			do(resp.Body)
			break
		}
	}
}

func MvnTree(dir *model.File) []*model.DepGraph {

	if dir == nil {
		return nil
	}

	if _, err := exec.LookPath("mvn"); err != nil {
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		logs.Warn(err)
		return nil
	}
	defer os.Chdir(pwd)

	os.Chdir(dir.Abspath)
	cmd := exec.Command("mvn", "dependency:tree", "--fail-never")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logs.Warn(err)
		return nil
	}

	// 记录当前处理的依赖树数据
	var lines []string
	// 标记是否在依赖范围内树
	tree := false
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)

	var roots []*model.DepGraph

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
			if root != nil {
				root.Path = dir.Relpath
				roots = append(roots, root)
			}
			lines = nil
			continue
		}
		if tree {
			lines = append(lines, line)
			continue
		}
	}

	return roots
}

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
