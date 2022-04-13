/*
 * @Descripation: 解析gem文件
 * @Date: 2021-11-30 15:03:08
 */

package ruby

import (
	"regexp"
	"sort"
	"strings"
	"util/model"
)

// parseGemfileLock 解析Gemfile.lock文件
func parseGemfileLock(root *model.DepTree, file *model.FileData) (deps []*model.DepTree) {
	deps = []*model.DepTree{}
	data := string(file.Data)
	subreg := regexp.MustCompile(`[ ]{6}(\S+)`)
	reg := regexp.MustCompile(`[ ]{4}(\S+) \(([\d.]+)\)`)
	// 记录组件信息
	// 当前依赖信息
	var now *model.DepTree
	// map[name]dep
	depMap := map[string]*model.DepTree{}
	// 记录有无父组件
	parentMap := map[string]struct{}{}
	// map[id]subname
	subMap := map[int64][]string{}
	for _, line := range strings.Split(data, "\n") {
		if subreg.MatchString(line) {
			match := subreg.FindStringSubmatch(line)
			if len(match) == 2 && now != nil {
				name := match[1]
				subMap[now.ID] = append(subMap[now.ID], name)
			}
		} else if reg.MatchString(line) {
			match := reg.FindStringSubmatch(line)
			if len(match) == 3 {
				name, ver := match[1], match[2]
				parentMap[name] = struct{}{}
				if dep, ok := depMap[name]; !ok {
					now = model.NewDepTree(nil)
					now.Name = name
					now.Version = model.NewVersion(ver)
					depMap[name] = now
				} else {
					now = dep
				}
			}
		}
	}
	// 构建依赖树
	// 查找直接依赖
	names := []string{}
	// 找出没有父组件的组件
	for _, subs := range subMap {
		for _, n := range subs {
			delete(parentMap, n)
		}
	}
	for name := range parentMap {
		names = append(names, name)
	}
	sort.Strings(names)
	// 添加直接依赖
	q := model.NewQueue()
	for _, name := range names {
		dep := depMap[name]
		dep.Parent = root
		root.Children = append(root.Children, dep)
		deps = append(deps, dep)
		q.Push(dep)
	}
	// 层级遍历添加间接依赖
	for !q.Empty() {
		dep := q.Pop().(*model.DepTree)
		subs := subMap[dep.ID]
		sort.Strings(subs)
		for _, name := range subs {
			if sub, ok := depMap[name]; ok && sub.Parent == nil {
				sub.Parent = dep
				dep.Children = append(dep.Children, sub)
				q.Push(sub)
			}
		}
	}
	return
}
