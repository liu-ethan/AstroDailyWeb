package logger

import (
	"log/slog"
	"os"
)

// New 创建结构化 JSON 日志器。
// 参数：level - 日志级别，支持 debug/info/warn/error。
// 返回：*slog.Logger - 已初始化的日志实例。
func New(level string) *slog.Logger {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	return slog.New(h)
}
