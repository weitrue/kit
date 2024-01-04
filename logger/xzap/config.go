package xzap

// Config 配置信息
type Config struct {
	ServiceName string
	Mode        string
	Path        string
	Level       string
	Compress    bool
	KeepDays    int
}
