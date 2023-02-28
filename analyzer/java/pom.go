package java

import (
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"util/args"
	"util/model"
)

type SimplePom struct {
	PomEnv
	PomDependency
	Parent       PomDependency   `xml:"parent" json:"parent,omitempty"`
	Dependencies []PomDependency `xml:"dependencies>dependency" json:"dependencies,omitempty"`
}

// Pom is Project Object Model
type Pom struct {
	SimplePom
	// 依赖当前pom的pom，并非parent标签
	ParentPom       *Pom            `xml:"-" json:"-"`
	DependenciesPom []*Pom          `xml:"-" json:"-"`
	Modules         []string        `xml:"modules>module" json:"-"`
	Repositories    []string        `xml:"repositories>repository>url" json:"-"`
	Mirrors         []string        `xml:"mirrors>mirror>url" json:"-"`
	Licenses        []string        `xml:"licenses>license>name" json:"-"`
	Fileinfo        *model.FileInfo `xml:"-" json:"-"`
}

// PomDependency is dependency tag
type PomDependency struct {
	ArtifactId string        `xml:"artifactId" json:"artifactId"`
	GroupId    string        `xml:"groupId" json:"groupId,omitempty"`
	Version    string        `xml:"version" json:"version,omitempty"`
	Type       string        `xml:"type" json:"type,omitempty"`
	Scope      string        `xml:"scope" json:"scope,omitempty"`
	Optional   string        `xml:"optional" json:"optional,omitempty"`
	Exclusions PomExclusions `xml:"exclusions" json:"exclusions,omitempty"`
	Classifier string        `xml:"classifier" json:"classifier,omitempty"`
	// 这里会被处理为绝对路径
	RelativePath string `xml:"relativePath" json:"-"`
	define       *Pom   `xml:"-" json:"-"`
}

// PomProperties is properties tag
type PomProperties map[string]string

// PomManagement is dependencyManagement tag
type PomManagement map[string]PomDependency

// PomExclusions is exclusion tag
type PomExclusions map[string]struct{}

// PomEnv Pom的环境变量
type PomEnv struct {
	Properties                    PomProperties `xml:"properties" json:"properties,omitempty"`
	DependencyManagement          PomManagement `xml:"dependencyManagement>dependencies" json:"dependencyManagement,omitempty"`
	DependencyExclusion           PomExclusions `xml:"-" json:"-"`
	DependencyManagementExclusion PomExclusions `xml:"-" json:"-"`
}

func (p PomDependency) ImportPath() string {
	paths := []string{p.Index4()}
	for p.define != nil {
		p = p.define.PomDependency
		paths = append(paths, p.Index4())
	}
	return strings.Join(paths, "\n ")
}

// Implement the Unmarshaler interface
func (pp *PomProperties) UnmarshalXML(d *xml.Decoder, s xml.StartElement) error {
	*pp = PomProperties{}
	for {
		e := struct {
			XMLName xml.Name
			Value   string `xml:",chardata"`
		}{}
		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		(*pp)[e.XMLName.Local] = trim(e.Value)
	}
	return nil
}

// Implement the Unmarshaler interface
func (pm *PomManagement) UnmarshalXML(d *xml.Decoder, s xml.StartElement) error {
	*pm = PomManagement{}
	for {
		e := PomDependency{}
		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		pm.Add(e)
	}
	return nil
}

// Implement the Unmarshaler interface
func (pm *PomExclusions) UnmarshalXML(d *xml.Decoder, s xml.StartElement) error {
	*pm = PomExclusions{}
	for {
		e := PomDependency{}
		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		pm.Add(e)
	}
	return nil
}

// unionEnv 获取两个Env的并集 返回新的Env 存在相同属性时保留传入的第一个Env
func unionEnv(a, b PomEnv) PomEnv {
	env := PomEnv{Properties: PomProperties{}, DependencyManagement: PomManagement{}, DependencyExclusion: PomExclusions{}, DependencyManagementExclusion: PomExclusions{}}
	// 合并 Properties
	for k, v := range b.Properties {
		env.Properties[k] = v
	}
	for k, v := range a.Properties {
		env.Properties[k] = v
	}
	// 合并 DependencyManagement
	for k, v := range b.DependencyManagement {
		env.DependencyManagement[k] = v
	}
	for k, v := range a.DependencyManagement {
		env.DependencyManagement[k] = v
	}
	// 合并 Exclusion
	for k, v := range b.DependencyManagementExclusion {
		env.DependencyManagementExclusion[k] = v
	}
	for k, v := range a.DependencyManagementExclusion {
		env.DependencyManagementExclusion[k] = v
	}
	for k, v := range b.DependencyExclusion {
		env.DependencyExclusion[k] = v
	}
	for k, v := range a.DependencyExclusion {
		env.DependencyExclusion[k] = v
	}
	return env

}

func (pm PomExclusions) Add(dep PomDependency) {
	if pm == nil {
		pm = PomExclusions{}
	}
	trimPomDependency(&dep)
	k := dep.Index2()
	if k != ":" {
		pm[k] = struct{}{}
	}
}

func (pm PomManagement) Add(dep PomDependency) {
	if pm == nil {
		pm = PomManagement{}
	}
	trimPomDependency(&dep)
	k := dep.Index2()
	if k != ":" {
		if _, exist := pm[k]; !exist {
			pm[k] = dep
		}
	}
}

// Index2 is groupId:artifactId:classifier:type
func (pd PomDependency) Index2() string {
	if pd.Type == "jar" {
		pd.Type = ""
	}
	return fmt.Sprintf("%s:%s:%s:%s", pd.GroupId, pd.ArtifactId, pd.Classifier, pd.Type)
}

// Index3 is groupId:artifactId:classifier:type:version
func (pd PomDependency) Index3() string {
	return fmt.Sprintf("%s:%s", pd.Index2(), pd.Version)
}

// Index4 is groupId:artifactId:classifier:type:version:scope
func (pd PomDependency) Index4() string {
	return fmt.Sprintf("%s:%s", pd.Index3(), pd.Scope)
}

func (p Pom) Update(properties PomProperties, pd *PomDependency) {
	if strings.HasPrefix(pd.GroupId, "$") {
		pd.GroupId = p.GetProperty(properties, pd.GroupId)
	}
	if strings.HasPrefix(pd.ArtifactId, "$") {
		pd.ArtifactId = p.GetProperty(properties, pd.ArtifactId)
	}
	if strings.HasPrefix(pd.Version, "$") {
		pd.Version = p.GetProperty(properties, pd.Version)
	}
}

func (p Pom) UpdateWeak(env PomEnv, pd *PomDependency) {
	p.Update(env.Properties, pd)
	// management
	if d, exist := env.DependencyManagement[pd.Index2()]; exist {
		if pd.Version == "" {
			pd.Version = p.GetProperty(env.Properties, d.Version)
		}
		if pd.Scope == "" && d.Scope != "" {
			pd.Scope = d.Scope
		}
		if pd.Exclusions == nil && len(d.Exclusions) > 0 {
			pd.Exclusions = PomExclusions{}
		}
		for k, v := range d.Exclusions {
			pd.Exclusions[k] = v
		}
	}
}

// UpdateForce is complete dependency index by pom property
func (p Pom) UpdateForce(env PomEnv, pd *PomDependency) {
	p.UpdateWeak(env, pd)
	if d, exist := env.DependencyManagement[pd.Index2()]; exist {
		pd.Version = p.GetProperty(env.Properties, d.Version)
	}
}

func (m *Mvn) SetFile(f *model.FileInfo, p *Pom) {
	if f == nil {
		return
	}
	p.Fileinfo = f
	fixRelativePath := func(relative *string) {
		rp := *relative
		if rp == "" {
			rp = "../pom.xml"
		}
		rp = path.Join(path.Dir(f.Name), rp)
		// 路径下存在文件且不是当前文件则修改路径
		if _, ok := m.poms[*relative]; ok && f.Name != rp {
			*relative = rp
		}
	}
	// parent不为空则尝试补全parent相对路径
	if p.Parent.ArtifactId != "" {
		fixRelativePath(&p.Parent.RelativePath)
	}
}

// ReadPom is parse a pom file retuen Pom pointer
func (m Mvn) ReadPomFile(f *model.FileInfo, data []byte) (p *Pom) {
	data = regexp.MustCompile(`xml version="1.0" encoding="\S+"`).ReplaceAll(data, []byte(`xml version="1.0" encoding="UTF-8"`))
	p = &Pom{}
	p.Properties = PomProperties{}
	xml.Unmarshal(data, &p)
	m.SetFile(f, p)
	p.Update(p.Properties, &p.Parent)
	p.Update(p.Properties, &p.PomDependency)
	// parent添加默认路径
	if p.Parent.RelativePath == "" && strings.Contains(p.Parent.Index3(), "$") {
		if p.Fileinfo != nil {
			p.Parent.RelativePath = path.Join(path.Dir(p.Fileinfo.Name), "../pom.xml")
		}
	}
	if r, ok := m.getPomWithPath(p.Parent.RelativePath); ok {
		if p.Parent.ArtifactId == r.ArtifactId {
			p.PomEnv = unionEnv(p.PomEnv, r.PomEnv)
			p.Update(p.Properties, &p.Parent)
			p.Update(p.Properties, &p.PomDependency)
		}
	}
	trimPomDependency(&p.Parent)
	trimPomDependency(&p.PomDependency)
	p.Update(p.Properties, &p.Parent)
	if p.GroupId == "" && p.Parent.GroupId != "" {
		p.GroupId = p.Parent.GroupId
	}
	if p.Version == "" && p.Parent.Version != "" {
		p.Version = p.Parent.Version
	}
	p.Update(p.Properties, &p.PomDependency)
	trimPomDependency(&p.PomDependency)
	// 去掉版本号为区间的组件(如: [1.2.0-alpha])
	for i, d := range p.Dependencies {
		trimPomDependency(&d)
		p.Dependencies[i] = d
		if len(d.Version) > 2 && d.Version[0] == '[' && d.Version[len(d.Version)-1] == ']' && !strings.ContainsAny(d.Version, ",()") {
			p.Dependencies[i].Version = strings.Trim(p.Dependencies[i].Version, "[]")
		}
	}
	// mvn添加repos
	for _, repo := range p.Repositories {
		if _, ok := m.repos[repo]; !ok {
			m.repos[repo] = args.RepoConfig{
				Repo: repo,
			}
		}
	}
	return
}

// GetProperty is get property from pom properties
func (p Pom) GetProperty(properties PomProperties, key string) string {
	switch key {
	case "${project.version}", "${version}", "${pom.version}":
		return p.Version
	case "${project.groupId}", "${groupId}", "${pom.groupId}":
		return p.GroupId
	case "${project.artifactId}":
		return p.ArtifactId
	case "${project.parent.version}", "${parent.version}":
		return p.Parent.Version
	case "${project.parent.groupId}", "${parent.groupId}":
		return p.Parent.GroupId
	default:
		return regexp.MustCompile(`\$\{[^{}]*\}`).ReplaceAllStringFunc(key,
			func(s string) string {
				exist := map[string]struct{}{}
				for strings.HasPrefix(s, "$") {
					if _, ok := exist[s]; ok {
						break
					} else {
						exist[s] = struct{}{}
					}
					k := s[2 : len(s)-1]
					if v, ok := properties[k]; ok {
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
}

var reg = regexp.MustCompile(`\s`)

// trim 去除空白字符
func trim(s string) string {
	return reg.ReplaceAllString(s, "")
}

// trimPomDependency 去除空白字符
func trimPomDependency(p *PomDependency) {
	p.GroupId = trim(p.GroupId)
	p.ArtifactId = trim(p.ArtifactId)
	p.Version = trim(p.Version)
	p.Scope = trim(p.Scope)
	p.Classifier = trim(p.Classifier)
	p.Optional = trim(p.Optional)
	p.Type = trim(p.Type)
}
