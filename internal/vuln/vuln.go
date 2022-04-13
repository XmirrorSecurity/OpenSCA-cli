/*
 * @Descripation: 查找漏洞
 * @Date: 2021-11-16 21:05:05
 */

package vuln

import (
	"errors"
	"opensca/internal/args"
	"opensca/internal/srt"
)

// SearchVuln 查找漏洞
func SearchVuln(root *srt.DepTree) (err error) {
	queue := srt.NewQueue()
	queue.Push(root)
	deps := []*srt.DepTree{}
	for !queue.Empty() {
		node := queue.Pop().(*srt.DepTree)
		deps = append(deps, node)
		for _, child := range node.Children {
			queue.Push(child)
		}
	}
	localVulns := [][]*srt.Vuln{}
	serverVulns := [][]*srt.Vuln{}
	ds := make([]srt.Dependency, len(deps))
	for i, d := range deps {
		ds[i] = d.Dependency
	}
	if args.VulnDB != "" {
		localVulns = GetLocalVulns(ds)
	}
	if args.Url != "" && args.Token != "" {
		serverVulns, err = GetServerVuln(ds)
	} else if args.VulnDB == "" && args.Url == "" && args.Token != "" {
		err = errors.New("url is null")
	} else if args.VulnDB == "" && args.Url != "" && args.Token == "" {
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
		if len(serverVulns) != 0 {
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
