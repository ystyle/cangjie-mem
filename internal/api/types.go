package api

import "encoding/json"

// APIResponse 统一 API 响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string, code ...string) *APIResponse {
	errInfo := &ErrorInfo{Message: message}
	if len(code) > 0 && code[0] != "" {
		errInfo.Code = code[0]
	}
	return &APIResponse{
		Success: false,
		Error:   errInfo,
	}
}

// JSONBytes 将响应转换为 JSON 字节
func (r *APIResponse) JSONBytes() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
