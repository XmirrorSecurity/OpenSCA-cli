package format

import (
	"encoding/json"
	"fmt"
	"io"
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
		Region struct {
			StartColumn int `json:"startColumn"`
			EndColumn   int `json:"endColumn"`
			StartLine   int `json:"startLine"`
			EndLine     int `json:"endLine"`
		} `json:"region"`
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
				location.PhysicalLocation.Region.StartColumn = 1
				location.PhysicalLocation.Region.EndColumn = 1
				location.PhysicalLocation.Region.StartLine = 1
				location.PhysicalLocation.Region.EndLine = 1
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
	outWrite(out, func(w io.Writer) error {
		return json.NewEncoder(w).Encode(s)
	})
}

func formatDesc(v *detail.VulnInfo) string {
	table := []struct {
		fmt string
		val string
	}{
		{"| id | %s |", v.Id},
		{"| --- | --- |", ""},
		{"| cve | %s |", v.Cve},
		{"| cnnvd | %s |", v.Cnnvd},
		{"| cnvd | %s |", v.Cnvd},
		{"| cwe | %s |", v.Cwe},
		{"| level | %s |", v.SecurityLevel()},
		{"| desc | %s |", v.Description},
		{"| suggestion | %s |", v.Suggestion},
	}
	var lines []string
	for _, line := range table {
		if strings.Contains(line.fmt, "%s") && line.val == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf(line.fmt, line.val))
	}
	return strings.Join(lines, "\n")
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
