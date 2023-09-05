package sbom

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

func ParseDsdx(f *model.File) *model.DepGraph {

	dependencies := map[string][]string{}
	depIdMap := map[string]*model.DepGraph{}
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	var group, name, version, id, language string
	f.ReadLine(func(line string) {
		i := strings.Index(line, ":")
		if strings.HasPrefix(line, "#") || i == -1 {
			if id != "" {
				dep := _dep(group, name, version)
				dep.Language = model.Language(language)
				depIdMap[id] = dep
			}
			group, name, version, id, language = "", "", "", "", ""
			return
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		switch k {
		case "ComponentName":
			name = v
		case "ComponentGroup":
			group = v
		case "ComponentVersion":
			group = v
		case "ComponentLanguage":
			language = v
		case "Dependencies":
			json.Unmarshal([]byte(v), &dependencies)
		}
	})

	for parentId, childrenIds := range dependencies {
		parent, ok := depIdMap[parentId]
		if !ok {
			continue
		}
		for _, childId := range childrenIds {
			parent.AppendChild(depIdMap[childId])
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
	_dep := model.NewDepGraphMap(nil, func(s ...string) *model.DepGraph {
		return &model.DepGraph{
			Vendor:  s[0],
			Name:    s[1],
			Version: s[2],
		}
	}).LoadOrStore

	for _, c := range doc.Components {
		dep := _dep(c.Group, c.Name, c.Version)
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
