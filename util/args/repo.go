package args

type RepoConfig struct {
	Repo     string `json:"repo"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// GetRepoConfig 获取仓库配置
func GetRepoConfig() map[string]RepoConfig {
	cfg := map[string]RepoConfig{}
	for _, r := range Config.Maven {
		cfg[r.Repo] = r
	}
	return cfg
}

type OriginConfig struct {
	Dsn string `json:"dsn"`
}
