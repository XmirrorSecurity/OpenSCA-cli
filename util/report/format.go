package report

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"util/args"
	"util/enum/language"
	"util/logs"
	"util/model"
)

// 任务检查信息
type TaskInfo struct {
	ToolVersion string  `json:"tool_version" xml:"tool_version" `
	AppName     string  `json:"app_name" xml:"app_name" `
	Size        int64   `json:"size" xml:"size" `
	StartTime   string  `json:"start_time" xml:"start_time" `
	EndTime     string  `json:"end_time" xml:"end_time" `
	CostTime    float64 `json:"cost_time" xml:"cost_time" `
	Error       error   `json:"-" xml:"-" `
	ErrorString string  `json:"error,omitempty" xml:"error,omitempty" `
}

// Format 按照输出内容格式化(不可逆)
func Format(dep *model.DepTree) {
	q := []*model.DepTree{dep}
	// 保留要导出的数据
	for len(q) > 0 {
		n := q[0]
		q = append(q[1:], n.Children...)
		if n.Language != language.None {
			n.LanguageStr = n.Language.String()
		}
		if n.Version != nil {
			n.VersionStr = n.Version.Org
		}
		if n.Path != "" {
			n.Paths = []string{n.Path}
		}
		n.Language = language.None
		n.Version = nil
	}
	// 去重
	if args.Config.Dedup {
		q = []*model.DepTree{dep}
		dm := map[string]*model.DepTree{}
		for len(q) > 0 {
			n := q[0]
			q = append(q[1:], n.Children...)
			// 去重
			k := fmt.Sprintf("%s:%s@%s#%s", n.Vendor, n.Name, n.VersionStr, n.LanguageStr)
			if d, ok := dm[k]; !ok {
				dm[k] = n
			} else {
				// 临时解决部分组件homepage字段不显示问题
				// 因为去重时刚好把解析到homepage字段的组件去掉了
				// 其他字段可能也需要类似操作
				if n.HomePage != "" {
					d.HomePage = n.HomePage
				}
				// 已存在相同组件
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
}

func outWrite(do func(io.Writer)) {
	out := args.Config.Out

	if out == "" {
		do(os.Stdout)
		return
	}

	pwd, _ := os.Getwd()
	fmt.Printf("Working directory: %s, Output file: %s\n", pwd, out)
	// 尝试创建导出文件目录
	if err := os.MkdirAll(filepath.Dir(out), fs.ModePerm); err != nil {
		logs.Warn(err)
		fmt.Println(err)
		return
	}
	w, err := os.Create(out)
	if err != nil {
		logs.Warn(err)
	} else {
		defer w.Close()
		do(w)
	}
}
