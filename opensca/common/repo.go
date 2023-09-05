package common

type RepoConfig struct {
	Url      string `json:"repo"`
	Username string `json:"user"`
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
