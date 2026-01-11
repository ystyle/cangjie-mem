package version

// 版本信息（编译时通过 -ldflags 注入）
var (
	// Version 版本号
	Version = "v1.0.0-dev"
	// GitCommit Git 提交哈希
	GitCommit = "unknown"
	// BuildDate 构建日期
	BuildDate = "unknown"
)
