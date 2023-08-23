package model

// // Vuln 组件漏洞
// type Vuln struct {
// 	Name            string `json:"name,omitempty" gorm:"column:name"`
// 	Id              string `json:"id" gorm:"column:id"`
// 	Cve             string `json:"cve_id,omitempty" gorm:"column:cve_id"`
// 	Cnnvd           string `json:"cnnvd_id,omitempty" gorm:"column:cnnvd_id"`
// 	Cnvd            string `json:"cnvd_id,omitempty" gorm:"column:cnvd_id"`
// 	Cwe             string `json:"cwe_id,omitempty" gorm:"column:cwe_id"`
// 	Description     string `json:"description,omitempty" gorm:"column:description"`
// 	DescriptionEn   string `json:"description_en,omitempty" gorm:"-"`
// 	Suggestion      string `json:"suggestion,omitempty" gorm:"column:suggestion"`
// 	AttackType      string `json:"attack_type,omitempty" gorm:"column:attack_type"`
// 	ReleaseDate     string `json:"release_date,omitempty" gorm:"column:release_date"`
// 	SecurityLevelId int    `json:"security_level_id" gorm:"column:security_level_id"`
// 	ExploitLevelId  int    `json:"exploit_level_id" gorm:"column:exploit_level_id"`
// }

// // NewVuln 创建Vuln
// func NewVuln() *Vuln {
// 	return &Vuln{}
// }

// LicenseInfo 许可证
type LicenseInfo struct {
	ShortName string `json:"name"`
	// TODO: expand
}
