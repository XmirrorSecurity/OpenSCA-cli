package java

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Pom is Project Object Model
type Pom struct {
	PomDependency
	Path                 string
	ParentPom            *Pom
	DependenciesPom      []*Pom
	Parent               PomDependency   `xml:"parent"`
	Dependencies         []PomDependency `xml:"dependencies>dependency"`
	DependencyManagement []PomDependency `xml:"dependencyManagement>dependencies>dependency"`
	Properties           PomProperties   `xml:"properties"`
	Repositories         []string        `xml:"repositories>repository>url"`
	Licenses             []string        `xml:"licenses>license>name"`
	Exclusion            map[string]map[string]struct{}
}

// PomDependency is dependency tag
type PomDependency struct {
	ArtifactId string          `xml:"artifactId"`
	GroupId    string          `xml:"groupId"`
	Version    string          `xml:"version"`
	Scope      string          `xml:"scope"`
	Optional   string          `xml:"optional"`
	Exclusions []PomDependency `xml:"exclusions>exclusion"`
}

// Deep is now pom in dependency tree level
func (p *Pom) Deep() int {
	deep := 0
	for p.ParentPom != nil {
		p = p.ParentPom
		deep++
	}
	return deep
}

// String is dependency tree when this pom top
func (pt *Pom) String() string {
	stack := []*Pom{pt}
	infos := []string{}
	for len(stack) > 0 {
		now := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		infos = append(infos, fmt.Sprintf("%s%s\n", strings.Repeat("\t", now.Deep()), now.Index3()))
		for i := len(now.DependenciesPom) - 1; i >= 0; i-- {
			c := now.DependenciesPom[i]
			stack = append(stack, c)
		}
	}
	return strings.Join(infos, "")
}

// Index2 is groupId:artifactId
func (pd PomDependency) Index2() string {
	return fmt.Sprintf("%s:%s", pd.GroupId, pd.ArtifactId)
}

// Index3 is groupId:artifactId:version
func (pd PomDependency) Index3() string {
	return fmt.Sprintf("%s:%s:%s", pd.GroupId, pd.ArtifactId, pd.Version)
}

// Complete is complete dependency index by pom property
func (p *Pom) Complete(pd *PomDependency) {
	if strings.HasPrefix(pd.GroupId, "$") {
		pd.GroupId = p.GetProperty(pd.GroupId)
	}
	if strings.HasPrefix(pd.ArtifactId, "$") {
		pd.ArtifactId = p.GetProperty(pd.ArtifactId)
	}
	if strings.HasPrefix(pd.Version, "$") {
		pd.Version = p.GetProperty(pd.Version)
	}
}

// PomProperties is properties tag
type PomProperties map[string]string

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
		(*pp)[e.XMLName.Local] = e.Value
	}
	return nil
}

// ReadPom is parse a pom file retuen Pom pointer
func ReadPom(data []byte) (p *Pom) {
	data = regexp.MustCompile(`xml version="1.0" encoding="\S+"`).ReplaceAll(data, []byte(`xml version="1.0" encoding="UTF-8"`))
	p = &Pom{Properties: PomProperties{}}
	xml.Unmarshal(data, &p)
	if p.GroupId == "" && p.Parent.GroupId != "" {
		p.GroupId = p.Parent.GroupId
	}
	if p.Version == "" && p.Parent.Version != "" {
		p.Version = p.Parent.Version
	}
	p.GroupId = strings.TrimSpace(p.GroupId)
	p.ArtifactId = strings.TrimSpace(p.ArtifactId)
	p.Version = strings.TrimSpace(p.Version)
	p.Complete(&p.PomDependency)
	return
}

// GetProperty is get property from pom properties
func (p *Pom) GetProperty(key string) string {
	switch key {
	case "${project.version}", "${version}", "${pom.version}":
		return p.Version
	case "${project.groupId}", "${groupId}", "${pom.groupId}":
		return p.GroupId
	case "${project.artifactId}", "${artifactId}", "${pom.artifactId}":
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
					if v, ok := p.Properties[k]; ok {
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
