package model

// 任务检测参数
type TaskArg struct {
	// 检测数据源 文件路径或url 兼容http(s)|ftp|file
	DataOrigin string
	// 检测对象名称 用于结果展示 缺省时取DataOrigin尾单词
	Name string
	// 超时时间 单位s
	Timeout int
}

// 任务检测结果
type TaskResult struct {
	// 检测目标名
	AppName string `json:"app_name" xml:"app_name" `
	// 检测文件大小
	Size int64 `json:"size" xml:"size" `
	// 任务开始时间
	StartTime string `json:"start_time" xml:"start_time" `
	// 任务结束时间
	EndTime string `json:"end_time" xml:"end_time" `
	// 任务检测耗时 单位s
	CostTime float64 `json:"cost_time" xml:"cost_time" `
	// 依赖图根节点
	DepRoot []*DepGraph `json:"-" xml:"-"`
	// 错误信息
	Error error `json:"-" xml:"-"`
}
