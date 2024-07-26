package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

type RepoConfig struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func TrimRepo(repos ...RepoConfig) []RepoConfig {
	var newRepos []RepoConfig
	for _, repo := range repos {
		if repo.Url != "" {
			newRepos = append(newRepos, repo)
		}
	}
	return newRepos
}

func DownloadUrlFromRepos(route string, do func(repo RepoConfig, r io.Reader), repos ...RepoConfig) bool {

	repoSet := map[string]bool{}

	for _, repo := range repos {

		if repo.Url == "" {
			continue
		}
		if repoSet[repo.Url] {
			continue
		}
		repoSet[repo.Url] = true

		url := fmt.Sprintf("%s/%s", strings.TrimRight(repo.Url, "/"), strings.TrimLeft(route, "/"))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logs.Warn(err)
			return false
		}

		if repo.Username+repo.Password != "" {
			req.SetBasicAuth(repo.Username, repo.Password)
		}

		resp, err := HttpDownloadClient.Do(req)
		if err != nil {
			logs.Warn(err)
			continue
		}

		if resp.StatusCode != 200 {
			logs.Warnf("%d %s", resp.StatusCode, url)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		} else {
			logs.Debugf("%d %s", resp.StatusCode, url)
			do(repo, resp.Body)
			resp.Body.Close()
			return true
		}

	}
	return false
}
