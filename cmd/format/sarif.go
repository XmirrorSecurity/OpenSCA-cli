package format

import (
	"encoding/json"
	"fmt"
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
			Name           string `json:"name"`
			Version        string `json:"version"`
			InformationUri string `json:"informationUri"`
		} `json:"driver"`
	} `json:"tool"`
	Results []sarifResult `json:"results"`
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
			if vuln.Id == "" {
				continue
			}
			result := sarifResult{
				RuleId: "XM1001",
				Level:  "warning",
			}
			result.Message.Text = fmt.Sprintf("id:%s cve:%s cwe:%s level:%s\ndesc:%s\nsuggestion:%s", vuln.Id, vuln.Cve, vuln.Cwe, vuln.SecurityLevel(), vuln.Description, vuln.Suggestion)
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
		json.NewEncoder(w).Encode(s)
	})
}
