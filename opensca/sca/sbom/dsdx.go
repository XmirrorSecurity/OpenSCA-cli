package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func ParseDsdx(f *model.File) *model.DepGraph {
	return parseDsdxDoc(f, ReadDsdx(f))
}

func ParseDsdxJson(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(f, doc)
}

func ParseDsdxXml(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(f, doc)
}

func parseDsdxDoc(f *model.File, doc *model.DsdxDocument) *model.DepGraph {

	if doc == nil || len(doc.Components) == 0 {
		return nil
	}

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[1],
			Name:    s[2],
			Version: s[3],
		}
	}).LoadOrStore

	for _, c := range doc.Components {
		dep := _dep(c.ID, c.Group, c.Name, c.Version)
		dep.Language = model.Language(c.Language)
		dep.Licenses = c.License
		depIdMap[c.ID] = dep
	}

	for parentId, childrenIds := range doc.Dependencies {
		parent, ok := depIdMap[parentId]
		if !ok {
			continue
		}
		for _, id := range childrenIds {
			parent.AppendChild(depIdMap[id])
		}
	}

	root := &model.DepGraph{Path: f.Relpath()}
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}

// ReadDsdx 读取dsdx文件
func ReadDsdx(f *model.File) *model.DsdxDocument {

	dsdx := &model.DsdxDocument{Dependencies: model.DsdxDependencies{}}

	// 记录依赖关系
	dependencies := map[string][]string{}
	// 记录dsdx中的tag信息
	tags := map[string]string{}

	checkAndSet := func(k, v string) {
		if _, ok := tags[k]; ok {
			dsdx.Components = append(dsdx.Components, model.DsdxComponent{
				ID:       tags["id"],
				Group:    tags["group"],
				Name:     tags["name"],
				Version:  tags["version"],
				Language: tags["language"],
			})
			tags = map[string]string{}
		}
		tags[k] = strings.TrimSpace(v)
	}

	f.ReadLine(func(line string) {
		i := strings.Index(line, ":")
		if strings.HasPrefix(line, "#") || i == -1 {
			return
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		switch k {
		case "ComponentID":
			checkAndSet("id", v)
		case "ComponentName":
			checkAndSet("name", v)
		case "ComponentGroup":
			checkAndSet("group", v)
		case "ComponentVersion":
			checkAndSet("version", v)
		case "ComponentLanguage":
			checkAndSet("language", v)
		case "Dependencies":
			json.Unmarshal([]byte(v), &dependencies)
		}
	})
	checkAndSet("id", "")

	for parent, children := range dependencies {
		dsdx.Dependencies[parent] = children
	}

	return dsdx
}
