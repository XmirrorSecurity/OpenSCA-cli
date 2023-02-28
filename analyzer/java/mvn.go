package java

import (
	"bytes"
	"path"
	"strings"
	"sync"
	"util/args"
	"util/model"
)

type Mvn struct {
	repos map[string]args.RepoConfig
	poms  map[string]*model.FileInfo
	lock  *sync.RWMutex
}

func NewMvn() Mvn {
	repos := args.GetRepoConfig()
	mvn := `https://repo.maven.apache.org/maven2/`
	repos[mvn] = args.RepoConfig{
		Repo: mvn,
	}
	return Mvn{
		repos: repos,
		poms:  map[string]*model.FileInfo{},
		lock:  &sync.RWMutex{},
	}
}

// parseProperties 获取 Properties
func (m Mvn) parseProperties(p *Pom) PomEnv {
	env := PomEnv{}
	for p != nil {
		env = unionEnv(env, PomEnv{Properties: p.Properties})
		p = m.GetPom(p.Parent)
	}
	return env
}

// checkExclusion 判断是否被排除
func checkExclusion(pd PomDependency, exc PomExclusions) bool {
	for _, d := range []PomDependency{
		pd,
		{GroupId: "*", ArtifactId: "*"},
		{GroupId: "*", ArtifactId: pd.ArtifactId},
		{GroupId: pd.GroupId, ArtifactId: "*"},
		{GroupId: pd.GroupId},
		{ArtifactId: pd.ArtifactId},
	} {
		if _, ok := exc[d.Index2()]; ok {
			return true
		}
	}
	return false
}

func (m Mvn) importPom(n *Pom, sp PomDependency, other func(n, s *Pom)) {
	n.Update(n.Properties, &sp)
	s := m.GetPom(sp)
	if s != nil {
		s.PomDependency = sp
		s.define = n
		for k, d := range s.DependencyManagement {
			d.define = s
			s.DependencyManagement[k] = d
		}
		other(n, s)
	}
}

// parseEnv 解析环境变量
func (m Mvn) parseEnv(p *Pom) PomEnv {
	env := m.parseProperties(p)
	// 获取 DependencyManagement
	q := []*Pom{p}
	set := map[string]struct{}{p.Index3(): {}}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		e := m.parseProperties(n)
		// 获取 management
		for key, d := range n.DependencyManagement {
			// 非当前 pom 引入的 DependencyManagement
			if d.define != nil && d.define != n {
				continue
			}
			// 更新坐标
			n.Update(n.Properties, &d)
			// 排除 exclusion
			if checkExclusion(d, n.DependencyManagementExclusion) {
				delete(n.DependencyManagement, key)
				continue
			}
			n.DependencyManagement[d.Index2()] = d
		}
		env = unionEnv(env, PomEnv{DependencyManagement: n.DependencyManagement})
		e = unionEnv(n.PomEnv, e)
		// parent添加默认路径
		if n.Parent.RelativePath == "" && strings.Contains(n.Parent.Index3(), "$") {
			if n.Fileinfo != nil {
				n.Parent.RelativePath = path.Join(path.Dir(n.Fileinfo.Name), "../pom.xml")
			}
		}
		// 添加 parent
		m.importPom(n, n.Parent, func(n, s *Pom) {
			if _, exist := set[s.Index3()]; exist {
				return
			}
			set[s.Index3()] = struct{}{}
			s.PomEnv = unionEnv(unionEnv(PomEnv{Properties: n.Properties}, s.PomEnv), n.PomEnv)
			q = append([]*Pom{s}, q...)
		})
		// 获取 management 的 import
		for _, d := range n.DependencyManagement {
			if d.define != nil && d.define != n {
				continue
			}
			p.Update(e.Properties, &d)
			// 排除 exclusion
			if checkExclusion(d, n.DependencyManagementExclusion) {
				continue
			}
			if d.Scope == "import" {
				m.importPom(n, d, func(n, s *Pom) {
					if _, exist := set[s.Index3()]; exist {
						return
					}
					set[s.Index3()] = struct{}{}
					// 记录累计 exclusion
					s.DependencyManagementExclusion = PomExclusions{}
					for k, v := range d.Exclusions {
						s.DependencyManagementExclusion[k] = v
					}
					for key, v := range n.DependencyManagementExclusion {
						s.DependencyManagementExclusion[key] = v
					}
					q = append(q, s)
				})
			}
		}
	}
	return env
}

// ParsePoms 解析一组 Pom
func (m Mvn) ParsePoms(ps []*Pom, deep bool) []*Pom {
	env := PomEnv{Properties: PomProperties{}}
	for _, p := range ps {
		if v, exist := p.Properties["revision"]; exist {
			env.Properties["revision"] = v
		}
	}
	for _, p := range ps {
		p.PomEnv = unionEnv(p.PomEnv, env)
		p.Update(p.Properties, &p.Parent)
		p.Update(p.Properties, &p.PomDependency)
		// cache local pom
		if p.Fileinfo != nil {
			setCachePom(p.GroupId, p.ArtifactId, p.Version, p.Fileinfo.Data)
		}
	}
	// 存在build标签的pom及其module(包含pom本身)
	buildMap := map[string]struct{}{}
	for _, p := range ps {
		// 找出实际构建的pom
		if p.Fileinfo != nil {
			data := p.Fileinfo.Data
			if bytes.Contains(data, []byte("<build>")) && bytes.Contains(data, []byte("</build>")) {
				buildMap[p.ArtifactId] = struct{}{}
				for _, module := range p.Modules {
					buildMap[module] = struct{}{}
				}
			}
		}
	}
	poms := []*Pom{}
	// buildMap的pom中的dependency(不包含pom本身)
	depMap := map[string]struct{}{}
	for _, p := range ps {
		if _, ok := buildMap[p.ArtifactId]; ok {
			for _, d := range p.Dependencies {
				depMap[d.ArtifactId] = struct{}{}
			}
		}
	}
	// 将会被其他pom依赖的pom从buildMap中移除
	for name := range depMap {
		delete(buildMap, name)
	}
	// buildMap中的pom判断为需要真实解析的pom
	if len(buildMap) > 0 {
		for _, p := range ps {
			if _, ok := buildMap[p.ArtifactId]; ok {
				m.ParsePom(p, deep)
				poms = append(poms, p)
			}
		}
	} else {
		for _, p := range ps {
			m.ParsePom(p, deep)
			poms = append(poms, p)
		}
	}
	return poms
}

// ParsePom 解析 Pom
func (m Mvn) ParsePom(p *Pom, deep bool) {
	env := m.parseEnv(p)
	q := []*Pom{p}
	set := map[string]struct{}{p.Index2(): {}}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		var e PomEnv
		if n == p {
			e = env
		} else {
			e = m.parseEnv(n)
			// 先仅用当前环境解析
			for i, d := range n.Dependencies {
				n.UpdateWeak(e, &d)
				n.Dependencies[i] = d
			}
			// 合并根环境
			e = unionEnv(env, e)
		}
		// 合并parent的dependencies
		if n.Parent.ArtifactId != "" {
			depSet := map[string]bool{}
			for _, d := range n.Dependencies {
				depSet[d.Index2()] = true
			}
			p := m.GetPom(n.Parent)
			for p != nil {
				for _, d := range p.Dependencies {
					if !depSet[d.Index2()] {
						depSet[d.Index2()] = true
						d.define = p
						n.Dependencies = append(n.Dependencies, d)
					}
				}
				p = m.GetPom(p.Parent)
			}
		}
		// 根据maven规则校准当前dependencies
		for _, d := range n.Dependencies {
			// 更新坐标
			if n == p {
				n.UpdateWeak(e, &d)
			} else {
				n.UpdateForce(e, &d)
			}
			// 排除 exclusion
			if checkExclusion(d, n.DependencyExclusion) {
				continue
			}
			// 排除未引入的依赖
			if n != p && (d.Scope == "test" || d.Optional == "true") {
				continue
			}
			if d.Scope == "provided" {
				continue
			}
			// 已处理过
			if _, exist := set[d.Index2()]; exist {
				continue
			}
			m.importPom(n, d, func(n, s *Pom) {
				set[d.Index2()] = struct{}{}
				s.ParentPom = n
				n.DependenciesPom = append(n.DependenciesPom, s)
				if deep {
					// 记录累计 exclusion
					s.DependencyExclusion = PomExclusions{}
					for k, v := range d.Exclusions {
						s.DependencyExclusion[k] = v
					}
					for k, v := range n.DependencyExclusion {
						s.DependencyExclusion[k] = v
					}
					q = append(q, s)
				}
			})
		}
	}
}
