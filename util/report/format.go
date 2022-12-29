package report

import (
	"fmt"
	"os"
	"util/args"
	"util/enum/language"
	"util/logs"
	"util/model"
)

// 任务检查信息
type TaskInfo struct {
	ToolVersion string  `json:"tool_version"`
	AppName     string  `json:"app_name"`
	Size        int64   `json:"size"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	CostTime    float64 `json:"cost_time"`
	Error       error   `json:"-"`
	ErrorString string  `json:"error,omitempty"`
}

// format 按照输出内容格式化(不可逆)
func format(dep *model.DepTree) {
	// 去重
	if args.Config.Dedup {
		q := []*model.DepTree{dep}
		dm := map[string]*model.DepTree{}
		for len(q) > 0 {
			n := q[0]
			q = append(q[1:], n.Children...)
			// 去重
			k := fmt.Sprintf("%s:%s@%s#%s", n.Vendor, n.Name, n.Version.Org, n.Language.String())
			if d, ok := dm[k]; !ok {
				dm[k] = n
			} else {
				// 已存在相同组件，但是某些字段可能不一样

				// 临时解决部分组件homepage字段不显示问题
				// 因为去重时刚好把解析到homepage字段的组件去掉了
				// 其他字段可能也需要类似操作
				if n.HomePage != "" {
					d.HomePage = n.HomePage
				}
				// 是否有直接依赖
				if n.Direct {
					d.Direct = n.Direct
				}
				d.Paths = append(d.Paths, n.Path)
				// 从父组件中移除当前组件
				if n.Parent != nil {
					for i, c := range n.Parent.Children {
						if c.ID == n.ID {
							n.Parent.Children = append(n.Parent.Children[:i], n.Parent.Children[i+1:]...)
							break
						}
					}
				}
				// 将当前组件的子组件转移到已存在组件的子依赖中
				d.Children = append(d.Children, n.Children...)
				for _, c := range n.Children {
					c.Parent = d
				}
			}
		}
	}

	// 保留要导出的数据
	q := []*model.DepTree{dep}
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n.Language != language.None {
			n.LanguageStr = n.Language.String()
		}
		if n.Version != nil {
			n.VersionStr = n.Version.Org
		}
		if len(n.Paths) == 0 {
			if n.Path != "" {
				n.Paths = []string{n.Path}
			}
		}
		//不展示的字段置空
		n.Path = ""
		n.Language = language.None
		n.Version = nil
		n.ID = 0
	}
}

// Save 保存结果文件
func Save(data []byte, filepath string) {
	if len(data) > 0 {
		if f, err := os.Create(filepath); err != nil {
			logs.Error(err)
		} else {
			defer f.Close()
			f.Write(data)
		}
	}
}
