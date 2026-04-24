package response

import "net/http"

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 生成成功响应。
// 参数：data - 响应数据。
// 返回：Envelope - 统一响应结构。
func Success(data interface{}) Envelope {
	return Envelope{Code: 200, Message: "success", Data: data}
}

// SuccessMessage 生成带自定义文案的成功响应。
// 参数：message - 成功提示；data - 响应数据。
// 返回：Envelope - 统一响应结构。
func SuccessMessage(message string, data interface{}) Envelope {
	return Envelope{Code: 200, Message: message, Data: data}
}

// Fail 生成失败响应。
// 参数：code - 业务错误码；message - 错误描述。
// 返回：Envelope - 统一响应结构。
func Fail(code int, message string) Envelope {
	return Envelope{Code: code, Message: message, Data: nil}
}

// HTTPStatusFromBizCode 将业务错误码映射为 HTTP 状态码。
// 参数：code - 业务错误码。
// 返回：int - HTTP 状态码。
func HTTPStatusFromBizCode(code int) int {
	switch {
	case code == 200:
		return http.StatusOK
	case code == 401 || code == 4010 || code == 4011:
		return http.StatusUnauthorized
	case code >= 400 && code < 5000:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
