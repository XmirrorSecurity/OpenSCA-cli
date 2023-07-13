/*
 * @Description: 从云服务获取漏洞
 * @Date: 2021-12-08 16:30:46
 */

package vuln

import (
	"encoding/json"
	"fmt"
	"util/client"
	"util/logs"
	"util/model"
)

// GetServerVuln 从云服务获取漏洞
func GetServerVuln(deps []model.Dependency) (vulns [][]*model.Vuln, err error) {
	vulns = [][]*model.Vuln{}
	data, err := json.Marshal(deps)
	if err != nil {
		logs.Error(err)
		return
	}
	data, err = client.Detect("vuln", data)
	if err != nil {
		fmt.Printf("\n%s", err.Error())
		return vulns, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &vulns)
		if err != nil {
			logs.Error(err)
		}
	}
	return
}

// GetServerLicense 从云服务获取许可证
func GetServerLicense(deps []model.Dependency) (lics [][]model.LicenseInfo, err error) {
	lics = [][]model.LicenseInfo{}
	data, err := json.Marshal(deps)
	if err != nil {
		logs.Error(err)
		return
	}
	data, err = client.Detect("license", data)
	if err != nil {
		fmt.Printf("\n%s", err.Error())
		return lics, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &lics)
		if err != nil {
			logs.Error(err)
		}
	}
	return
}
