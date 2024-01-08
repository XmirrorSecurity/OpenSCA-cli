package format

import (
	"encoding/json"
	"fmt"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"io"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/detail"
)

type sarifReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool struct {
		Driver struct {
			Name           string      `json:"name"`
			Version        string      `json:"version"`
			InformationUri string      `json:"informationUri"`
			Rules          []sarifRule `json:"rules"`
		} `json:"driver"`
	} `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifRule struct {
	Id               string                    `json:"id"`
	ShortDescription sarifRuleShortDescription `json:"shortDescription"`
	FullDescription  sarifRuleFullDescription  `json:"fullDescription"`
	Help             sarifRuleHelp             `json:"help"`
	Properties       sarifRuleProperties       `json:"properties"`
}

type sarifRuleShortDescription struct {
	Text string `json:"text"`
}

type sarifRuleFullDescription struct {
	Text string `json:"text"`
}

type sarifRuleHelp struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

type sarifRuleProperties struct {
	Tags []string `json:"tags"`
}

type sarifResult struct {
	RuleId  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			Uri   string `json:"uri"`
			Index int    `json:"index,omitempty"`
		} `json:"artifactLocation"`
		// Region struct {
		// 	StartLine   int `json:"startLine"`
		// 	StartColumn int `json:"startColumn"`
		// } `json:"region"`
	} `json:"physicalLocation"`
}

func Sarif(report Report, out string) {

	s := sarifReport{
		Version: "2.1.0",
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0.json",
	}

	run := sarifRun{}
	run.Tool.Driver.Name = "opensca-cli"
	run.Tool.Driver.Version = strings.TrimLeft(report.TaskInfo.ToolVersion, "vV")
	run.Tool.Driver.InformationUri = "https://opensca.xmirror.cn"

	report.ForEach(func(n *detail.DepDetailGraph) bool {
		for _, vuln := range n.Vulnerabilities {
			run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, sarifRule{
				Id: vuln.Id,
				ShortDescription: sarifRuleShortDescription{
					Text: fmt.Sprintf("[%s] 组件 %s 中存在 %s", vuln.SecurityLevel(), n.Dep.Key()[:strings.LastIndex(n.Dep.Key(), ":")], vuln.Name),
				},
				Help: sarifRuleHelp{
					Markdown: formatDesc(vuln),
				},
				Properties: sarifRuleProperties{
					Tags: formatTags(n),
				},
			},
			)

			if vuln.Id == "" {
				continue
			}
			result := sarifResult{
				RuleId: "XM1001",
				Level:  "warning",
			}
			result.Message.Text = fmt.Sprintf("引入的组件 %s 中存在 %s", n.Dep.Key()[:strings.LastIndex(n.Dep.Key(), ":")], vuln.Name)
			for i, path := range n.Paths {
				location := sarifLocation{}
				if truncIndex := strings.Index(path, "["); truncIndex > 0 {
					path = path[:truncIndex]
					path = strings.TrimPrefix(path, filepath.Base(report.TaskInfo.AppName))
					path = strings.Trim(path, `\/`)
				}
				location.PhysicalLocation.ArtifactLocation.Uri = path
				location.PhysicalLocation.ArtifactLocation.Index = i
				result.Locations = append(result.Locations, location)
			}
			run.Results = append(run.Results, result)
		}
		return true
	})

	s.Runs = []sarifRun{run}
	outWrite(out, func(w io.Writer) {
		err := json.NewEncoder(w).Encode(s)
		if err != nil {
			logs.Error(err)
		}
	})
}

func formatDesc(v *detail.Vuln) string {
	text := fmt.Sprintf("| id | %s |", v.Id)
	text = fmt.Sprintf("%s\n| --- | --- |", text)
	if v.Cve != "" {
		text = fmt.Sprintf("%s\n| cve | %s |", text, v.Cve)
	}
	if v.Cnnvd != "" {
		text = fmt.Sprintf("%s\n| cnnvd | %s |", text, v.Cnnvd)
	}
	if v.Cnvd != "" {
		text = fmt.Sprintf("%s\n| cnvd | %s |", text, v.Cnvd)
	}
	if v.Cwe != "" {
		text = fmt.Sprintf("%s\n| cwe | %s |", text, v.Cwe)
	}
	text = fmt.Sprintf("%s\n| level | %s |", text, v.SecurityLevel())
	text = fmt.Sprintf("%s\n| desc | %s |", text, v.Description)
	text = fmt.Sprintf("%s\n| suggestion | %s |", text, v.Suggestion)

	return text
}

func formatTags(dg *detail.DepDetailGraph) []string {
	tags := []string{}
	exists := make(map[string]bool)
	tags = append(tags, "security")
	tags = append(tags, "Use-Vulnerable-and-Outdated-Components")
	if dg.Dep.Language != "" {
		tags = append(tags, dg.Dep.Language)
		exists[dg.Dep.Language] = true
	}
	for _, v := range dg.Vulnerabilities {
		if v.Cve != "" && !exists[v.Cve] {
			tags = append(tags, v.Cve)
			exists[v.Cve] = true
		}
		if v.AttackType != "" && !exists[v.AttackType] {
			tags = append(tags, v.AttackType)
			exists[v.AttackType] = true
		}
	}
	return tags
}
