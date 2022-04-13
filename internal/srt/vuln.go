/*
 * @Descripation: 漏洞信息
 * @Date: 2021-12-11 14:54:46
 */

package srt

// Vuln 组件漏洞
type Vuln struct {
	Name            string `json:"name,omitempty"`
	Id              string `json:"id"`
	Cve             string `json:"cve_id,omitempty"`
	Cnnvd           string `json:"cnnvd_id,omitempty"`
	Cnvd            string `json:"cnvd_id,omitempty"`
	Cwe             string `json:"cwe_id,omitempty"`
	Description     string `json:"description,omitempty"`
	DescriptionEn   string `json:"description_en,omitempty"`
	Suggestion      string `json:"suggestion,omitempty"`
	AttackType      string `json:"attack_type,omitempty"`
	ReleaseDate     string `json:"release_date,omitempty"`
	SecurityLevelId int    `json:"security_level_id"`
	ExploitLevelId  int    `json:"exploit_level_id"`
}

// NewVuln 创建Vuln
func NewVuln() *Vuln {
	return &Vuln{}
}
