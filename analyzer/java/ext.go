/*
 * @Descripation: mvn解析依赖树
 * @Date: 2021-12-16 10:10:13
 */

package java

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"util/args"
	"util/cache"
	"util/enum/language"
	"util/logs"
	"util/model"
	"util/temp"

	"github.com/pkg/errors"
)

// MvnDepTree 调用mvn解析项目获取依赖树
func MvnDepTree(path string, root *model.DepTree) {
	Len := len(root.Children)
	pwd := temp.GetPwd()
	os.Chdir(path)
	cmd := exec.Command("mvn", "dependency:tree", "--fail-never")
	out, _ := cmd.CombinedOutput()
	os.Chdir(pwd)
	// 统一替换换行符为\n
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\n\r"), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte("\r"), []byte("\n"))
	// 获取mvn解析内容
	lines := strings.Split(string(out), "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "[INFO] ")
	}
	// 捕获依赖树起始位置
	title := regexp.MustCompile(`--- [^\n]+ ---`)
	// 记录依赖树起始位置行号
	start := 0
	// 标记是否在依赖范围内树
	tree := false
	root.Direct = true
	// 获取mvn依赖树
	for i, line := range lines {
		if title.MatchString(line) {
			tree = true
			start = i
			continue
		}
		if tree && strings.Trim(line, "-") == "" {
			tree = false
			buildMvnDepTree(root, lines[start+1:i])
			for _, c := range root.Children {
				c.Direct = true
			}
			continue
		}
	}

	if len(root.Children) != Len {
		mvnSuccess = true
	}
	return
}

// buildMvnDepTree 构建mvn树
func buildMvnDepTree(root *model.DepTree, lines []string) {
	// 记录当前的顶点节点列表
	tops := []*model.DepTree{root}
	// 上一层级
	lastLevel := -1
	for _, line := range lines {
		// 计算层级
		level := 0
		for line[level*3+2] == ' ' {
			level++
		}
		tops = tops[:len(tops)-lastLevel+level-1]
		root = tops[len(tops)-1]
		tags := strings.Split(line[level*3:], ":")
		if len(tags) < 4 {
			logs.Error(errors.New("mvn parse error"))
			break
		}
		dep := model.NewDepTree(root)
		dep.Vendor = tags[0]
		dep.Name = tags[1]
		dep.Version = model.NewVersion(tags[3])
		dep.Language = language.Java
		tops = append(tops, dep)
		lastLevel = level
	}
}

// downloadPom 下载pom文件
func downloadPom(dep model.Dependency) (data []byte, err error) {
	tags := strings.Split(dep.Vendor, ".")
	tags = append(tags, dep.Name)
	tags = append(tags, dep.Version.Org)
	tags = append(tags, fmt.Sprintf("%s-%s.pom", dep.Name, dep.Version.Org))
	// 先扫描指定仓库
	for _, m := range args.Config.Maven {
		url := strings.TrimSuffix(m.Repo, `/`) + `/`
		url = url + strings.Join(tags, "/")
		name := m.User
		password := m.Password
		data, err = getFromRepo(url, name, password)
		if data == nil {
			continue
		}
		return
	}
	// 指定仓库都没有就去官方仓库查询
	d := `https://repo.maven.apache.org/maven2/`
	url := d + strings.Join(tags, "/")
	if rep, err := http.Get(url); err != nil {
		return nil, err
	} else {
		defer rep.Body.Close()
		if rep.StatusCode == 200 {
			return ioutil.ReadAll(rep.Body)
		}
	}
	// 应该走不到这里
	return nil, fmt.Errorf("download failure")
}

// 从私服库获取pom文件
func getFromRepo(url string, name string, password string) (data []byte, err error) {
	c := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, Timeout: time.Duration(1 * time.Second)}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	} else {
		resp.Request.SetBasicAuth(name, password)
		defer resp.Body.Close()
		logs.Debug(fmt.Sprintf("status code: %d url: %s", resp.StatusCode, url))
		if resp.StatusCode == 200 {
			return ioutil.ReadAll(resp.Body)
		}
	}
	return nil, fmt.Errorf("download from repository failure")
}

// getpom is get pom from index
func getpom(groupId, artifactId, version string) (p *Pom) {
	p = &Pom{Properties: PomProperties{}}
	if groupId == "" || artifactId == "" || version == "" {
		return nil
	}
	dep := model.Dependency{
		Vendor:  groupId,
		Name:    artifactId,
		Version: model.NewVersion(version),
	}
	data := cache.LoadCache(dep)
	if len(data) != 0 {
		return ReadPom(data)
	} else {
		// 无本地缓存下载pom文件
		if data, err := downloadPom(dep); err == nil {
			// 保存pom文件
			cache.SaveCache(dep, data)
			return ReadPom(data)
		} else {
			logs.Warn(err)
		}
	}
	return nil
}
