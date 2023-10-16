package java

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/java/xml"
)

type Pom struct {
	PomDependency
	Parent               PomDependency    `xml:"parent"`
	Properties           PomProperties    `xml:"properties"`
	DependencyManagement []*PomDependency `xml:"dependencyManagement>dependencies>dependency"`
	Dependencies         []*PomDependency `xml:"dependencies>dependency"`
	Modules              []string         `xml:"modules>module"`
	Repositories         []string         `xml:"repositories>repository>url"`
	Mirrors              []string         `xml:"mirrors>mirror>url"`
	Licenses             []string         `xml:"licenses>license>name"`
	Profiles             []Pom            `xml:"profiles>profile"`
	File                 *model.File      `xml:"-" json:"-"`
}

type PomDependency struct {
	ArtifactId   string           `xml:"artifactId"`
	GroupId      string           `xml:"groupId"`
	Version      string           `xml:"version"`
	Type         string           `xml:"type"`
	Scope        string           `xml:"scope"`
	Classifier   string           `xml:"classifier"`
	RelativePath string           `xml:"relativePath"`
	Optional     bool             `xml:"optional"`
	Exclusions   []*PomDependency `xml:"exclusions>exclusion"`
	Define       *Pom             `xml:"-"`
	RefProperty  *Property        `xml:"-"`
	Start        int              `xml:",start"`
	End          int              `xml:",end"`
}

type Property struct {
	Key    string
	Value  string
	Define *Pom
	Start  int `xml:",start"`
	End    int `xml:",end"`
}

type PomProperties map[string]*Property

func (pp *PomProperties) UnmarshalXML(d *xml.Decoder, s xml.StartElement) error {
	*pp = PomProperties{}
	for {
		e := struct {
			XMLName xml.Name
			Value   string `xml:",chardata"`
			Start   int    `xml:",start"`
			End     int    `xml:",end"`
		}{}
		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		(*pp)[e.XMLName.Local] = &Property{
			Key:   e.XMLName.Local,
			Value: e.Value,
			Start: e.Start,
			End:   e.End,
		}
	}
	return nil
}

func (pd PomDependency) NeedExclusion(dep PomDependency) bool {
	check := func(s1, s2 string) bool {
		return s1 == "" || s1 == "*" || s1 == s2
	}
	for _, e := range pd.Exclusions {
		if check(e.GroupId, dep.GroupId) &&
			check(e.ArtifactId, dep.ArtifactId) &&
			check(e.Version, dep.Version) {
			return true
		}
	}
	return false
}

// GAV is groupId:artifactId:version
func (pd PomDependency) GAV() string {
	return fmt.Sprintf("%s:%s:%s", pd.GroupId, pd.ArtifactId, pd.Version)
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

func ReadPom(reader io.Reader) *Pom {

	data, err := io.ReadAll(reader)
	if err != nil {
		logs.Warn(err)
		return nil
	}

	data = regexp.MustCompile(`xml version="1.0" encoding="\S+"`).ReplaceAll(data, []byte(`xml version="1.0" encoding="UTF-8"`))
	p := &Pom{Properties: PomProperties{}}
	xml.Unmarshal(data, &p)

	trimSpace(&p.Parent)
	trimSpace(&p.PomDependency)

	if p.GroupId == "" {
		p.GroupId = p.Parent.GroupId
	}
	if p.Version == "" {
		p.Version = p.Parent.Version
	}

	p.Parent.Define = p
	for _, v := range p.Properties {
		v.Define = p
	}
	for _, d := range p.DependencyManagement {
		d.Define = p
	}
	for _, d := range p.Dependencies {
		trimSpace(d)
		d.Define = p
	}

	// 添加内置属性
	p.Properties["project.groupId"] = &Property{Key: "project.groupId", Value: p.GroupId}
	p.Properties["groupId"] = &Property{Key: "groupId", Value: p.GroupId}
	p.Properties["pom.groupId"] = &Property{Key: "pom.groupId", Value: p.GroupId}
	p.Properties["project.parent.groupId"] = &Property{Key: "project.parent.groupId", Value: p.Parent.GroupId}
	p.Properties["parent.groupId"] = &Property{Key: "parent.groupId", Value: p.Parent.GroupId}
	p.Properties["project.version"] = &Property{Key: "project.version", Value: p.Version}
	p.Properties["version"] = &Property{Key: "version", Value: p.Version}
	p.Properties["pom.version"] = &Property{Key: "pom.version", Value: p.Version}
	p.Properties["project.parent.version"] = &Property{Key: "project.parent.version", Value: p.Parent.Version}
	p.Properties["parent.version"] = &Property{Key: "parent.version", Value: p.Parent.Version}

	// 添加第一个Profile
	if len(p.Profiles) > 0 {
		profile := p.Profiles[0]
		for k, v := range profile.Properties {
			if _, ok := p.Properties[k]; !ok {
				p.Properties[k] = v
			}
		}
		p.Dependencies = append(p.Dependencies, profile.Dependencies...)
		p.DependencyManagement = append(p.DependencyManagement, profile.DependencyManagement...)
		p.Modules = append(p.Modules, profile.Modules...)
		p.Mirrors = append(p.Mirrors, profile.Mirrors...)
		p.Repositories = append(p.Repositories, profile.Repositories...)
	}

	// 处理版本范围
	for _, d := range p.Dependencies {
		if strings.ContainsAny(d.Version, "()[]") {
			if !strings.Contains(d.Version, ",") {
				d.Version = strings.Trim(d.Version, "()[]")
			} else {
				lr := strings.Split(strings.Split(d.Version, "!")[0], ",")
				if len(lr) == 2 {
					l, r := lr[0], lr[1]
					if strings.HasPrefix(l, "[") {
						d.Version = strings.TrimLeft(l, "[")
					}
					if strings.HasSuffix(r, "]") {
						d.Version = strings.TrimRight(l, "]")
					}
				}
			}
		}
	}

	// 存在厂商和组件相同的依赖时保留最后声明的
	depSet := map[string]bool{}
	for i := len(p.Dependencies) - 1; i >= 0; i-- {
		if depSet[p.Dependencies[i].Index2()] {
			p.Dependencies = append(p.Dependencies[:i], p.Dependencies[i+1:]...)
		} else {
			depSet[p.Dependencies[i].Index2()] = true
		}
	}

	return p
}

func (p *Pom) Update(dep *PomDependency) {
	var ref *Property
	dep.GroupId, ref = p.update(dep.GroupId)
	if ref != nil {
		dep.RefProperty = ref
	}
	dep.Version, ref = p.update(dep.Version)
	if ref != nil {
		dep.RefProperty = ref
	}
}

var propertyReg = regexp.MustCompile(`\$\{[^{}]*\}`)

func (p *Pom) update(value string) (val string, ref *Property) {
	val = propertyReg.ReplaceAllStringFunc(value,
		func(s string) string {
			exist := map[string]struct{}{}
			for strings.HasPrefix(s, "$") {
				if _, ok := exist[s]; ok {
					break
				} else {
					exist[s] = struct{}{}
				}
				k := s[2 : len(s)-1]
				if v, ok := p.Properties[k]; ok {
					if len(v.Value) > 0 {
						s = v.Value
						ref = v
						continue
					}
				}
				break
			}
			return s
		})
	return
}

var reg = regexp.MustCompile(`\s`)

func trimSpace(p *PomDependency) {
	if p == nil {
		return
	}
	trim := func(s string) string {
		return reg.ReplaceAllString(s, "")
	}
	p.GroupId = trim(p.GroupId)
	p.ArtifactId = trim(p.ArtifactId)
	p.Version = trim(p.Version)
	p.Scope = trim(p.Scope)
	p.Classifier = trim(p.Classifier)
	p.Type = trim(p.Type)
}

func (dep PomDependency) Check() bool {
	return !(dep.ArtifactId == "" || dep.GroupId == "" || dep.Version == "" || strings.Contains(dep.GAV(), "$"))
}

// ImportPath 引入路径
func (dep PomDependency) ImportPath() []PomDependency {
	paths := []PomDependency{dep}
	pom := dep.Define
	pomset := map[*Pom]bool{}
	for pom != nil {
		if pomset[pom] {
			break
		}
		pomset[pom] = true
		paths = append(paths, pom.PomDependency)
		pom = pom.Define
	}
	return paths
}

// ImportPathStack 引入路径栈
func (dep PomDependency) ImportPathStack() string {
	var importPaths []string
	paths := dep.ImportPath()
	for i, d := range paths {
		importPath := fmt.Sprintf("[%s]", d.Index4())
		if i > 0 {
			pre := paths[i-1]
			importPath += fmt.Sprintf("(line:%d-%d)", pre.Start, pre.End)
		}
		if d.RefProperty != nil {
			if d.RefProperty.Define != nil {
				importPath += fmt.Sprintf("#[%s](line:%d)", d.RefProperty.Define.Index4(), d.RefProperty.Start)
			}
			importPath += fmt.Sprintf("${%s}=%s", d.RefProperty.Key, d.RefProperty.Value)
		}
		importPaths = append(importPaths, importPath)
	}
	return strings.Join(importPaths, "<=")
}
