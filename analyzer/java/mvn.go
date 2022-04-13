package java

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

// DefaultDownloadPom is download pom form mvn repos
func DefaultDownloadPom(groupId, artifactId, version string) *Pom {
	local_path := `./.cache/`
	repo_url := `https://repo.maven.apache.org/maven2/`
	pom_dir := fmt.Sprintf("%s/%s/%s", strings.ReplaceAll(groupId, ".", "/"), artifactId, version)
	pom_path := fmt.Sprintf("%s/%s-%s.pom", pom_dir, artifactId, version)
	if _, err := os.Stat(local_path + pom_path); err == nil {
		if data, err := os.ReadFile(local_path + pom_path); err == nil {
			return ReadPom(data)
		}
	} else {
		fmt.Println(repo_url + pom_path)
		if rep, err := http.Get(repo_url + pom_path); err == nil {
			defer rep.Body.Close()
			if data, err := ioutil.ReadAll(rep.Body); err == nil {
				os.MkdirAll(local_path+pom_dir, os.ModeDir)
				if f, err := os.Create(local_path + pom_path); err == nil {
					defer f.Close()
					f.Write(data)
				} else {
					fmt.Println(err)
				}
				if rep.StatusCode == 200 {
					return ReadPom(data)
				}
			}
		}
	}
	return nil
}

// copyMap is copy map[string]map[string]struct{}
func copyMap(dst, src map[string]map[string]struct{}) {
	for k1, v1 := range src {
		if _, ok := dst[k1]; !ok {
			dst[k1] = map[string]struct{}{}
		}
		v := dst[k1]
		for k2, v2 := range v1 {
			v[k2] = v2
		}
	}
}

func transferProperties(src, dst *Pom) {
	// transfer properties
	if src.Properties == nil {
		src.Properties = PomProperties{}
	}
	if dst.Properties == nil {
		dst.Properties = PomProperties{}
	}
	for k, v := range src.Properties {
		if _, ok := dst.Properties[k]; !ok {
			dst.Properties[k] = v
		}
	}
}

func transferManagement(src, dst *Pom) {
	// parent dependency map
	pdm := map[string]string{}
	// module dependency map
	mdm := map[string]string{}
	for _, dm := range src.DependencyManagement {
		pdm[dm.Index2()] = dm.Version
	}
	for _, dm := range dst.DependencyManagement {
		mdm[dm.Index2()] = dm.Version
	}
	// parent cover module
	for i, dm := range dst.DependencyManagement {
		if version, exist := pdm[dm.Index2()]; exist {
			dst.DependencyManagement[i].Version = version
		}
	}
	// module append parent
	for _, dm := range src.DependencyManagement {
		if _, exist := mdm[dm.Index2()]; !exist {
			(*dst).DependencyManagement = append((*dst).DependencyManagement, dm)
			mdm[dm.Index2()] = dm.Version
		}
	}
}

// Mvn is maven
type Mvn struct {
	pomMap      map[string]*Pom
	poms        []*Pom
	downloadPom func(groupId, artifactId, version string) *Pom
	transMap    map[string]struct{}
}

// NewMvn is create a *Mvn
func NewMvn() *Mvn {
	return &Mvn{
		pomMap:      map[string]*Pom{},
		poms:        []*Pom{},
		downloadPom: getpom,
		transMap:    map[string]struct{}{},
	}
}

// getPom is get pom from pomMap if exist else download pom
func (m *Mvn) getPom(p PomDependency) *Pom {
	if par, exist := m.pomMap[p.Index2()]; exist {
		if par.Properties == nil {
			par.Properties = PomProperties{}
		}
		return par
	} else {
		par = m.downloadPom(p.GroupId, p.ArtifactId, p.Version)
		if par == nil {
			par = &Pom{Properties: PomProperties{}, PomDependency: p}
		}
		if par.Properties == nil {
			par.Properties = PomProperties{}
		}
		return par
	}
}

// transferAll is transferParent and transferImport
func (m *Mvn) transferAll(p *Pom) {
	if _, exist := m.transMap[p.Index2()]; exist {
		return
	} else {
		m.transMap[p.Index2()] = struct{}{}
	}
	if p.Parent.ArtifactId != "" {
		par := m.getPom(p.Parent)
		if par != nil && par.ArtifactId != "" {
			m.pomMap[par.Index2()] = par
			m.transferAll(par)
			transferProperties(par, p)
			transferManagement(par, p)
		}
	}
	for i := range p.DependencyManagement {
		p.Complete(&p.DependencyManagement[i])
		dm := p.DependencyManagement[i]
		if dm.Scope == "import" {
			// import pom
			imp := m.getPom(dm)
			if imp != nil && imp.ArtifactId != "" {
				m.pomMap[imp.Index2()] = imp
				m.transferAll(imp)
				transferProperties(imp, p)
				transferManagement(imp, p)
			}
		}
	}
}

// AppendPom is Append pom to mvn
func (m *Mvn) AppendPom(p *Pom) {
	m.poms = append(m.poms, p)
	m.pomMap[p.Index2()] = p
}

// ReadPom is parse a pom file retuen Pom pointer
func (m *Mvn) ReadPom(data []byte) {
	m.AppendPom(ReadPom(data))
}

// MvnSimulation is simulation maven project
func (m *Mvn) MvnSimulation() []*Pom {
	poms := m.poms
	// complete dependencies by level each
	queue := make([]*Pom, len(poms))
	copy(queue, poms)
	// save exist dependency
	exist := map[string]struct{}{}
	for len(queue) > 0 {
		p := queue[0]
		p.Complete(&p.PomDependency)
		if _, ok := exist[p.Index2()]; !ok {
			exist[p.Index2()] = struct{}{}
		}
		// transfer pom self
		m.transferAll(p)
		// complete management
		dm := map[string]*PomDependency{}
		p.Complete(&p.PomDependency)
		for i := range p.DependencyManagement {
			pd := &p.DependencyManagement[i]
			p.Complete(pd)
			dm[pd.Index2()] = pd
		}
		for i := range p.Dependencies {
			dp := &p.Dependencies[i]
			if dp.Scope == "provided" {
				continue
			}
			if (dp.Scope == "test" || dp.Optional == "true") && p.Deep() > 0 {
				continue
			}
			p.Complete(dp)
			if _, ok := exist[dp.Index2()]; ok {
				continue
			} else {
				exist[dp.Index2()] = struct{}{}
			}
			if pdm, exist := dm[dp.Index2()]; exist {
				dp.Version = pdm.Version
				p.Complete(dp)
				if pdm.Scope == "provided" || pdm.Scope == "test" {
					dp.Scope = pdm.Scope
				}
				if pdm.Optional == "true" {
					dp.Optional = pdm.Optional
				}
				if dp.Scope == "provided" || dp.Scope == "test" || dp.Optional == "true" {
					continue
				}
			}
			// download indirect
			dpom := m.downloadPom(dp.GroupId, dp.ArtifactId, dp.Version)
			if dpom == nil {
				dpom = &Pom{}
			}
			transferManagement(p, dpom)
			dpom.PomDependency = *dp
			dpom.ParentPom = p
			p.DependenciesPom = append(p.DependenciesPom, dpom)
		}
		queue = append(queue[1:], p.DependenciesPom...)
	}
	// each stack
	stack := make([]*Pom, len(poms))
	copy(stack, poms)
	// exclusion dependencies
	for len(stack) > 0 {
		// current pom
		p := stack[len(stack)-1]
		// exclusion
		if exc, ok := p.Exclusion[p.Index2()]; ok {
			savePom := []*Pom{}
			saveDep := []PomDependency{}
			for _, d := range p.DependenciesPom {
				if _, ok := exc[d.Index2()]; !ok {
					savePom = append(savePom, d)
				}
			}
			for _, d := range p.Dependencies {
				if _, ok := exc[d.Index2()]; !ok {
					saveDep = append(saveDep, d)
				}
			}
			p.DependenciesPom = savePom
			p.Dependencies = saveDep
		}
		// save exclusion
		if p.Exclusion == nil {
			p.Exclusion = map[string]map[string]struct{}{}
		}
		for _, dep := range p.Dependencies {
			for _, exc := range dep.Exclusions {
				if _, ok := p.Exclusion[dep.Index2()]; !ok {
					p.Exclusion[dep.Index2()] = map[string]struct{}{}
				}
				p.Exclusion[dep.Index2()][exc.Index2()] = struct{}{}
			}
		}
		for _, child := range p.DependenciesPom {
			if child.Exclusion == nil {
				child.Exclusion = map[string]map[string]struct{}{}
			}
			copyMap(child.Exclusion, p.Exclusion)
		}
		stack = append(stack[:len(stack)-1], p.DependenciesPom...)
		p = nil
	}
	return poms
}

// ReadPomsFromDir is build maven
func (m *Mvn) ReadPomsFromDir(filepath string) {
	files, err := ioutil.ReadDir(filepath)
	if err != nil {
		fmt.Println(err)
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			fmt.Println(err)
		}
		m.ReadPom(data)
	}
	for _, f := range files {
		fp := path.Join(filepath, f.Name())
		if f.IsDir() {
			m.ReadPomsFromDir(fp)
		} else if strings.HasSuffix(f.Name(), "pom.xml") {
			data, err := ioutil.ReadFile(fp)
			if err != nil {
				fmt.Println(err)
			}
			m.ReadPom(data)
		}
	}
}

// SetDownloadPomFunc is use custom pom download function
func (m *Mvn) SetDownloadPomFunc(f func(groupId, artifactId, version string) *Pom) {
	m.downloadPom = f
}
