package apperror

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error 返回错误字符串。
// 参数：无。
// 返回：string - 组合后的错误描述。
func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
	}
	return fmt.Sprintf("code=%d message=%s err=%v", e.Code, e.Message, e.Err)
}

// New 创建业务错误。
// 参数：code - 业务错误码；message - 错误文案。
// 返回：*AppError - 业务错误对象。
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装底层错误为业务错误。
// 参数：code - 业务错误码；message - 错误文案；err - 底层错误。
// 返回：*AppError - 业务错误对象。
func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

var (
	ErrBadRequest     = New(4000, "参数错误")
	ErrUnauthorized   = New(4010, "未授权")
	ErrTokenInvalid   = New(4011, "Token无效或已过期")
	ErrInternal       = New(5000, "系统繁忙，请稍后重试")
	ErrNotImplemented = New(5001, "功能待实现")
)
