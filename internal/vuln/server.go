/*
 * @Descripation: 从云服务获取漏洞
 * @Date: 2021-12-08 16:30:46
 */

package vuln

import (
	"encoding/json"
	"opensca/internal/client"
	"opensca/internal/logs"
	"opensca/internal/srt"
)

// GetServerVuln 从云服务获取漏洞
func GetServerVuln(deps []srt.Dependency) (vulns [][]*srt.Vuln, err error) {
	vulns = [][]*srt.Vuln{}
	data, err := json.Marshal(deps)
	if err != nil {
		logs.Error(err)
		return
	}
	data, err = client.Detect(data)
	if err != nil {
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
