package python

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"util/bar"
	"util/ex"
	"util/logs"
	"util/model"
	"util/temp"
)

var reg1 *regexp.Regexp
var regGit *regexp.Regexp
var replacer *strings.Replacer

func init() {
	reg1 = regexp.MustCompile(`^\w`)
	regGit = regexp.MustCompile(`\/([\w-]+)\.git`)
	replacer =  strings.NewReplacer("# via","","\r",""," ","","#","")
}

func parseRequirementsin(root *model.DepTree, file *model.FileInfo) {
	// 检查python环境
	if _, err := ex.CheckPython(ex.Python); err != nil {
		return
	}
	strArry := []string{}
	temp.DoInTempDir(func(tempdir string) {
		// 安装piptools
		if _, err := ex.Do(ex.PipinstallPiptoos, tempdir); err != nil {
			logs.Error(err)
			return
		}
		// 删除虚拟环境
		defer ex.Do(ex.RemoveVirtualCmd, tempdir)
		// 获取输出数据
		strArry = getOutData(file, tempdir)
	})
	// 解析输出数据构建依赖树
	parseOutData(root, strArry)
}

// 解析各组件所打印的信息
func parseOutData(root *model.DepTree, strs []string) {
	// 直接依赖
	directMap := map[string]*model.DepTree{}
	childMap := map[*model.DepTree]map[string]struct{}{}
	for _, str := range strs {
		lines := strings.Split(str, "\n")
		for i, line := range lines {
			if reg1.MatchString(line) {
				lines = lines[i:]
				break
			}
		}
		// parentsMap一个组件名对应其所有父组件名
		var parentsMap = make(map[string][]string)
		cur := model.NewDepTree(nil)
		nodes := []string{}
		depMap := map[string]*model.DepTree{}
		for _, line := range lines {
			if strings.Contains(line, "==") {
				// 在输出内容"=="符号左右对应名字与版本号
				cur = model.NewDepTree(nil)
				line = strings.TrimSuffix(line, "\r")
				nv := strings.Split(line, `==`)
				if len(nv) == 2 {
					cur.Name = strings.TrimSpace(nv[0])
					cur.Version = model.NewVersion(strings.TrimSpace(nv[1]))
					depMap[cur.Name] = cur
					m := make(map[string]struct{})
					childMap[cur] = m
					nodes = append(nodes, cur.Name)
				}
			} else if strings.Contains(line, "#") {
				// "#"符号后有父组件名字信息
				line = replacer.Replace(line)
				if line == "" {
					continue
				}
				parentsMap[cur.Name] = append(parentsMap[cur.Name], line)
			}
		}
		depMap[cur.Name] = cur
		nodes = append(nodes, cur.Name)
		for _, name := range nodes {
			if _,ok := depMap[name]; !ok {
				continue
			}
			parNames := parentsMap[name]
			for _, parName := range parNames {
				if len(parNames) == 1 && strings.Contains(parName, "requirements") {
					if dep, ok := depMap[name]; ok {
						directMap[dep.Name] = dep
					}
				}
				if _,ok := depMap[parName]; !ok {
					continue
				}
				parent := depMap[parName]
				dep := depMap[name]
				if m,ok := childMap[dep]; ok {
					if _,ok := m[dep.Name];ok {
						continue
					}
					m[dep.Name] = struct{}{}
				}
				parent.Children = append(parent.Children, dep)
				dep.Parent = parent
			}
		}
	}
	withRoot(root,directMap)
}

// 所有直接依赖连接至root
func withRoot(root *model.DepTree,directMap map[string]*model.DepTree) {
	direct := []*model.DepTree{}
	for _, n := range directMap {
		direct = append(direct, n)
	}
	sort.Slice(direct, func(i, j int) bool {
		return direct[i].Name < direct[j].Name
	})
	for _, d := range direct {
		root.Children = append(root.Children, d)
		d.Parent = root
	}
}

// 获取打印数据
func getOutData(file *model.FileInfo, dir string) []string {
	s := string(file.Data)
	strList := []string{}
	reqpath := path.Join(dir, `requirements.in`)
	out, err := os.Create(reqpath)
	if err != nil {
		logs.Error(err)
		return strList
	}
	out.Close()
	for _, v := range strings.Split(s, "\n") {
		// 少部分情况会有git连接
		if regGit.MatchString(v) {
			res := regGit.FindStringSubmatch(v)
			if len(res) == 2 {
				bar.PipCompile.Add(1)
				strList = append(strList, getSingleModStr(reqpath, res[1]))
				continue
			}
		}
		// 一般情况下字母开头的行内容都是组件名
		if reg1.MatchString(v) {
			bar.PipCompile.Add(1)
			strList = append(strList, getSingleModStr(reqpath, v))
		}
	}
	return strList
}

// 将组件名与版本号写入requirements.in文件单独调用pip-compile，获取打印数据
func getSingleModStr(reqpath string, elem string) string {
	f, err := os.OpenFile(reqpath, os.O_RDWR, 0755)
	if err != nil {
		return ""
	}
	f.Seek(0, 0)
	f.Truncate(0)
	f.WriteString(elem)
	f.Close()
	if str, err := ex.Do(ex.PipCompilein, path.Dir(reqpath)); err != nil {
		logs.Error(err)
		logs.Error(fmt.Errorf("get info err:%s", elem))
		return ""
	} else {
		return str
	}
}