package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/model"
)

func ParseDsdx(f *model.File) *model.DepGraph {

	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(func(s ...string) string {
		return s[0]
	}, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:   s[1],
			Name:     s[2],
			Version:  s[3],
			Language: model.Language(s[4]),
		}
	}).LoadOrStore

	// 记录dsdx中的tag信息
	tags := map[string]string{}
	checkAndSet := func(k, v string) {
		if _, ok := tags[k]; ok {
			depIdMap[tags["id"]] = _dep(tags["id"], tags["group"], tags["name"], tags["version"], tags["language"])
			tags = map[string]string{}
		}
		tags[k] = strings.TrimSpace(v)
	}
	// 记录依赖关系
	dependencies := map[string][]string{}

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
	depIdMap[tags["id"]] = _dep(tags["id"], tags["group"], tags["name"], tags["version"], tags["language"])

	if len(depIdMap) == 0 {
		return nil
	}

	for parent, children := range dependencies {
		for _, child := range children {
			depIdMap[parent].AppendChild(depIdMap[child])
		}
	}

	var roots []*model.DepGraph
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 && dep.Name != "" {
			roots = append(roots, dep)
		}
	}

	if len(roots) == 1 {
		return roots[0]
	}

	root := &model.DepGraph{Path: f.Relpath()}
	for _, r := range roots {
		root.AppendChild(r)
	}
	return root
}

func ParseDsdxJson(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		json.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(doc)
}

func ParseDsdxXml(f *model.File) *model.DepGraph {
	doc := &model.DsdxDocument{}
	f.OpenReader(func(reader io.Reader) {
		xml.NewDecoder(reader).Decode(doc)
	})
	return parseDsdxDoc(doc)
}

func parseDsdxDoc(doc *model.DsdxDocument) *model.DepGraph {

	if doc == nil || doc.DSDXVersion == "" {
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

	root := &model.DepGraph{}
	for _, dep := range depIdMap {
		if len(dep.Parents) == 0 {
			root.AppendChild(dep)
		}
	}

	return root
}
