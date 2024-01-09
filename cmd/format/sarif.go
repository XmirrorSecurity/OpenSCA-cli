package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"

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
	Name             string                    `json:"name"`
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

	vulnInfos := map[string]*detail.VulnInfo{}

	report.ForEach(func(n *detail.DepDetailGraph) bool {
		for _, vuln := range n.Vulnerabilities {

			if vuln.Id == "" {
				continue
			}

			vulnInfos[vuln.Id] = &detail.VulnInfo{Vuln: vuln, Language: n.Language}

			result := sarifResult{
				RuleId: vuln.Id,
				Level:  "warning",
			}
			result.Message.Text = fmt.Sprintf("引入的组件 %s 中存在 %s", n.Dep.Key()[:strings.LastIndex(n.Dep.Key(), ":")], vuln.Name)
			for i, path := range n.Paths {
				location := sarifLocation{}
				location.PhysicalLocation.ArtifactLocation.Uri = path
				location.PhysicalLocation.ArtifactLocation.Index = i
				// location.PhysicalLocation.Region.StartColumn = 1
				// location.PhysicalLocation.Region.StartLine = 1
				result.Locations = append(result.Locations, location)
			}

			run.Results = append(run.Results, result)
		}
		return true
	})

	for _, vuln := range vulnInfos {
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, sarifRule{
			Id:               vuln.Id,
			Name:             vuln.Name,
			ShortDescription: sarifRuleShortDescription{Text: vuln.Name},
			FullDescription:  sarifRuleFullDescription{Text: vuln.Description},
			Help:             sarifRuleHelp{Markdown: formatDesc(vuln)},
			Properties:       sarifRuleProperties{Tags: formatTags(vuln)},
		})
	}

	s.Runs = []sarifRun{run}
	outWrite(out, func(w io.Writer) {
		err := json.NewEncoder(w).Encode(s)
		if err != nil {
			logs.Error(err)
		}
	})
}

func formatDesc(v *detail.VulnInfo) string {
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

func formatTags(v *detail.VulnInfo) []string {
	tags := []string{"security", "Use-Vulnerable-and-Outdated-Components", v.Cve, v.Cwe, v.AttackType, v.Language}
	for i := 0; i < len(tags); {
		if tags[i] == "" {
			tags = append(tags[:i], tags[i+1:]...)
		} else {
			i++
		}
	}
	return tags
}
