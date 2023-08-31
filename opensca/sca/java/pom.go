package java

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

type Pom struct {
	PomDependency
	Parent               PomDependency    `xml:"parent"`
	Properties           PomProperties    `xml:"properties"`
	DependencyManagement []*PomDependency `xml:"dependencyManagement>dependencies>dependency"`
	Dependencies         []*PomDependency `xml:"dependencies>dependency"`
	Modules              []string         `xml:"modules>module"`
	Repositories         []MvnRepo        `xml:"repositories>repository"`
	Mirrors              []MvnRepo        `xml:"mirrors>mirror"`
	Licenses             []string         `xml:"licenses>license>name"`
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
	// Start        int              `xml:",start"`
	// End          int              `xml:",end"`
	Define *Pom `xml:"-"`
}

type Property struct {
	Value string
	// Start  int
	// End    int
	Define *Pom
}

type PomProperties map[string]*Property

func (pp *PomProperties) UnmarshalXML(d *xml.Decoder, s xml.StartElement) error {
	*pp = PomProperties{}
	for {
		e := struct {
			XMLName xml.Name
			Value   string `xml:",chardata"`
			// Start   int    `xml:",start" json:"-"`
			// End     int    `xml:",end" json:"-"`
		}{}
		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		(*pp)[e.XMLName.Local] = &Property{
			Value: e.Value,
			// Start: e.Start,
			// End:   e.End,
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

	return p
}

func (p *Pom) Update(dep *PomDependency) {
	dep.Version = p.GetProperty(dep.Version)
}

var propertyReg = regexp.MustCompile(`\$\{[^{}]*\}`)

func (p *Pom) GetProperty(key string) string {
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
		return propertyReg.ReplaceAllStringFunc(key,
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
							continue
						}
					}
					break
				}
				return s
			})
	}
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
