/*
 * @Descripation: 查找漏洞
 * @Date: 2021-11-16 21:05:05
 */

package vuln

import (
	"errors"
	"util/args"
	"util/model"
)

// SearchVuln 查找漏洞
func SearchVuln(root *model.DepTree) (err error) {
	queue := model.NewQueue()
	queue.Push(root)
	deps := []*model.DepTree{}
	for !queue.Empty() {
		node := queue.Pop().(*model.DepTree)
		deps = append(deps, node)
		for _, child := range node.Children {
			queue.Push(child)
		}
	}
	localVulns := [][]*model.Vuln{}
	serverVulns := [][]*model.Vuln{}
	serverVulnsV2 := make(map[int64][]*model.Vuln)
	ds := make([]model.Dependency, len(deps))
	for i, d := range deps {
		ds[i] = d.Dependency
	}
	if args.Config.VulnDB != "" {
		localVulns = GetLocalVulns(ds)
	}
	if args.Config.Url != "" && args.Config.Token != "" {
		if args.Config.V2 {
			serverVulnsV2, err = GetServerVulnV2(root)
		} else {
			serverVulns, err = GetServerVuln(ds)
		}
	} else if args.Config.VulnDB == "" && args.Config.Url == "" && args.Config.Token != "" {
		err = errors.New("url is null")
	} else if args.Config.VulnDB == "" && args.Config.Url != "" && args.Config.Token == "" {
		err = errors.New("token is null")
	}
	for i, dep := range deps {
		// 合并本地和云端库搜索的漏洞
		exist := map[string]struct{}{}
		if len(localVulns) != 0 {
			for _, vuln := range localVulns[i] {
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
				}
			}
		}
		if args.Config.V2 {
			if len(serverVulnsV2) != 0 {
				if _, ok := serverVulnsV2[dep.ID]; ok {
					for _, vuln := range serverVulnsV2[dep.ID] {
						if _, ok := exist[vuln.Id]; !ok {
							exist[vuln.Id] = struct{}{}
							dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
						}
					}
				}
			}
		} else if len(serverVulns) != 0 {
			for _, vuln := range serverVulns[i] {
				if _, ok := exist[vuln.Id]; !ok {
					exist[vuln.Id] = struct{}{}
					dep.Vulnerabilities = append(dep.Vulnerabilities, vuln)
				}
			}
		}
	}
	return
}
