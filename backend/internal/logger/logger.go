package logger

import (
	"io"
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

	_ = os.MkdirAll("log", 0o755)
	file, err := os.OpenFile("log/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	var out io.Writer = os.Stdout
	if err == nil {
		out = io.MultiWriter(os.Stdout, file)
	}

	h := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: lv})
	return slog.New(h)
}
