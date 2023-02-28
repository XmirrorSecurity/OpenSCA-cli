package java

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"util/args"
	"util/cache"
	"util/enum/language"
	"util/logs"
	"util/model"
)

// setCachePom 添加缓存的pom
func setCachePom(groupId, artifactId, version string, data []byte) {
	if groupId != "" && artifactId != "" && version != "" {
		if !strings.ContainsAny(groupId+artifactId+version, "${}^!@#<>") {
			cache.SaveCache(model.Dependency{
				Vendor:   groupId,
				Name:     artifactId,
				Version:  model.NewVersion(version),
				Language: language.Java,
			}, data)
		}
	}
}

// getCachePom 获取缓存的pom
func (m Mvn) getCachePom(groupId, artifactId, version string) (p *Pom, ok bool) {
	data := cache.LoadCache(model.Dependency{
		Vendor:   groupId,
		Name:     artifactId,
		Version:  model.NewVersion(version),
		Language: language.Java,
	})
	if len(data) > 0 {
		p = m.ReadPomFile(nil, data)
		p.Parent.RelativePath = ""
		return p, true
	}
	return nil, false
}

var statu404set = sync.Map{}

// doReq 发送请求并校验返回状态
func doReq(url, username, password string, do func(rep *http.Response)) bool {
	if n, exist := statu404set.Load(url); exist {
		if n.(int) >= 100 {
			statu404set.Delete(url)
		} else {
			statu404set.Store(url, n.(int)+1)
		}
		return false
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Warn(err)
	} else {
		if username != "" {
			req.SetBasicAuth(username, password)
		}
		if rep, err := (&http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}).Do(req); err == nil {
			defer rep.Body.Close()
			logs.Debug(fmt.Sprintf("status code: %d url: %s", rep.StatusCode, url))
			if rep.StatusCode == 200 {
				do(rep)
				return true
			} else if rep.StatusCode == 404 {
				statu404set.Store(url, 1)
			}
		}
	}
	return false
}

// downloadLastPom 尝试下载指定url下最新的pom文件(特指snapshot版本的pom)
func (m Mvn) downloadLastPom(url, username, password string) (p *Pom, ok bool) {
	poms := []string{}
	subs := []string{}
	url = strings.TrimSuffix(url, "/") + "/"
	// 通过maven-metadata.xml文件获取最新版本
	doReq(url+"maven-metadata.xml", username, password, func(rep *http.Response) {
		if data, err := io.ReadAll(rep.Body); err == nil {
			metadata := struct {
				GroupId      string `xml:"groupId"`
				ArtifactId   string `xml:"artifactId"`
				LastTime     string `xml:"versioning>lastUpdated"`
				SnapVersions []struct {
					Version string `xml:"value"`
					Time    string `xml:"updated"`
				} `xml:"versioning>snapshotVersions>snapshotVersion"`
			}{}
			xml.Unmarshal(data, &metadata)
			if metadata.LastTime == "" {
				return
			}
			for _, v := range metadata.SnapVersions {
				if v.Time == metadata.LastTime {
					pomdata := cache.LoadCache(model.Dependency{
						Vendor:  metadata.GroupId,
						Name:    metadata.ArtifactId,
						Version: model.NewVersion(v.Version),
					})
					if len(pomdata) > 0 {
						p = m.ReadPomFile(nil, pomdata)
						ok = true
					}
					poms = append(poms, fmt.Sprintf("%s-%s.pom", metadata.ArtifactId, v.Version))
					break
				}
			}
		}
	})
	if ok {
		return
	}
	if len(poms) == 0 {
		// 如果没通过metadata获取到pom文件则尝试通过查找上一层目录获取pom列表
		doReq(url, username, password, func(rep *http.Response) {
			if data, err := io.ReadAll(rep.Body); err == nil {
				if reg, err := regexp.Compile(`href="([^"]+)"`); err == nil {
					for _, group := range reg.FindAllSubmatch(data, -1) {
						if len(group) == 2 {
							if bytes.HasSuffix(group[1], []byte(".pom")) {
								poms = append(poms, path.Base(string(group[1])))
							} else if bytes.HasSuffix(group[1], []byte("/")) {
								// 过滤上级目录
								if strings.HasPrefix(url, string(group[1])) || bytes.Contains(group[1], []byte("../")) {
									continue
								}
								subs = append(subs, path.Base(string(group[1])))
							}
						}
					}
				}
			}
		})
	}
	if len(poms) > 0 {
		// 获取最后一个pom文件
		doReq(url+poms[len(poms)-1], username, password, func(rep *http.Response) {
			if data, err := io.ReadAll(rep.Body); err == nil {
				p, ok = m.ReadPomFile(nil, data), true
				setCachePom(p.GroupId, p.ArtifactId, p.Version, data)
			}
		})
	} else {
		// 从最后一个目录反向遍历，找到pom文件后停止
		for i := len(subs) - 1; i >= 0; i-- {
			if p, ok = m.downloadLastPom(url+subs[i], username, password); ok {
				break
			}
		}
	}
	return
}

// downloadFromRepo try download pom form mvn repos
func (m Mvn) downloadFromRepo(groupId, artifactId, version string) (p *Pom, ok bool) {
	if version == "" {
		return
	}
	pom_dir := fmt.Sprintf("%s/%s/%s", strings.ReplaceAll(groupId, ".", "/"), artifactId, version)
	pom_path := fmt.Sprintf("%s/%s-%s.pom", pom_dir, artifactId, version)
	// 添加pom中的repo
	repos := []args.RepoConfig{}
	for _, repo := range m.repos {
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo < repos[j].Repo
	})
	// 尝试获取缓存的pom
	p, ok = m.getCachePom(groupId, artifactId, version)
	if ok {
		return
	}
	for _, r := range repos {
		repo := strings.TrimSuffix(r.Repo, "/") + "/"
		if !doReq(repo+pom_path, r.User, r.Password, func(rep *http.Response) {
			if data, err := io.ReadAll(rep.Body); err == nil {
				setCachePom(groupId, artifactId, version, data)
				p, ok = m.ReadPomFile(nil, data), true
			}
		}) {
			// 快照版本尝试获取最后一个pom
			if strings.Contains(version, "snap") || strings.Contains(version, "SNAP") {
				p, ok = m.downloadLastPom(repo+pom_dir, r.User, r.Password)
			}
		}
		if ok {
			return
		}
	}
	return
}

// setPoms 添加POM文件
func (m Mvn) setPoms(fs ...*model.FileInfo) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, f := range fs {
		m.poms[f.Name] = f
	}
}

// getPomWithPath 辅助实现 relativePath 功能
func (m Mvn) getPomWithPath(relativePath string) (*Pom, bool) {
	if relativePath == "" {
		return nil, false
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	if strings.HasSuffix(relativePath, "/") {
		relativePath += "pom.xml"
	}
	if f, ok := m.poms[relativePath]; ok {
		p := m.ReadPomFile(f, f.Data)
		return p, ok
	}
	return nil, false
}

// GetPom is get pom from pomMap if exist else download pom
func (m Mvn) GetPom(p PomDependency) (pom *Pom) {
	if p, ok := m.getPomWithPath(p.RelativePath); ok {
		return p
	}
	if p.ArtifactId == "" || p.GroupId == "" || p.Version == "" {
		return nil
	}
	if strings.ContainsAny(p.Index3(), "${}") {
		return nil
	}
	var ok bool
	defer func() {
		if pom.PomEnv.Properties == nil {
			pom.PomEnv.Properties = PomProperties{}
		}
	}()
	pom, ok = m.getCachePom(p.GroupId, p.ArtifactId, p.Version)
	if ok {
		return pom
	}
	pom, ok = m.downloadFromRepo(p.GroupId, p.ArtifactId, p.Version)
	if ok {
		return pom
	}
	return &Pom{SimplePom: SimplePom{PomDependency: p}}
}
