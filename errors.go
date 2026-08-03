package rock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ErrorCode 定义常见的错误代码
type ErrorCode int

const (
	// 4xx 客户端错误
	ErrBadRequest     ErrorCode = 400
	ErrUnauthorized   ErrorCode = 401
	ErrForbidden      ErrorCode = 403
	ErrNotFound       ErrorCode = 404
	ErrMethodNotAllow ErrorCode = 405
	ErrUnprocessable  ErrorCode = 422

	// 5xx 服务器错误
	ErrInternalServer ErrorCode = 500
	ErrBadGateway     ErrorCode = 502
)

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%d: %s - %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithDetail 创建带有详细信息的应用错误
func NewAppErrorWithDetail(code ErrorCode, message, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// HTTPError HTTP错误响应结构
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// ErrorResponse 统一的错误响应格式
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   HTTPError `json:"error"`
}

// WriteError 写入错误响应
func WriteError(c Context, statusCode int, err error) {
	var httpErr HTTPError

	// 根据错误类型设置响应
	switch e := err.(type) {
	case *AppError:
		httpErr = HTTPError{
			Code:    int(e.Code),
			Message: e.Message,
			Detail:  e.Detail,
		}
	case *ValidationError:
		httpErr = HTTPError{
			Code:    int(ErrBadRequest),
			Message: "Validation failed",
			Detail:  e.Error(),
		}
	default:
		// 未知错误类型
		httpErr = HTTPError{
			Code:    statusCode,
			Message: err.Error(),
		}
	}

	// 先设置响应头，再设置状态码并写入 body。
	// 状态码为懒写入（见 Ctx.writeHeader），c.Write 会在首次写入时发送响应头。
	c.SetHeader("Content-Type", "application/json")

	// 写入错误响应
	response := ErrorResponse{
		Success: false,
		Error:   httpErr,
	}

	body, err := json.Marshal(response)
	if err != nil {
		// 如果序列化失败，至少写入文本错误
		c.SetHeader("Content-Type", "text/plain")
		c.Status(statusCode)
		c.Write([]byte(fmt.Sprintf("Error: %s", httpErr.Message)))
		return
	}

	c.Status(statusCode)
	c.Write(body)
}

// writeJSONResponse 统一的JSON响应写入方法
func writeJSONResponse(writer io.Writer, data interface{}) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(data)
}

// NewError 创建带有格式化消息的应用错误
func NewError(code ErrorCode, format string, args ...interface{}) *AppError {
	return NewAppError(code, fmt.Sprintf(format, args...))
}

// WriteSuccess 写入成功响应
func WriteSuccess(c Context, data interface{}) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// ValidateRequest 验证请求参数
func ValidateRequest(c Context, params map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for field, rules := range params {
		if rules == nil || rules == "" {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: "field is required",
			})
		}
	}

	return errors
}

// HandlePanic 处理panic恢复
func HandlePanic(c Context, message interface{}) {
	// 记录错误日志
	if app := c.App(); app != nil && app.logger != nil {
		app.logger.Errorf("Panic recovered: %v", message)
	}

	// 返回500错误
	WriteError(c, http.StatusInternalServerError,
		NewAppError(ErrInternalServer, "Internal Server Error"))
}

// CommonErrorMessages 通用错误消息
var CommonErrorMessages = map[ErrorCode]string{
	ErrBadRequest:     "Bad Request",
	ErrUnauthorized:   "Unauthorized",
	ErrForbidden:      "Forbidden",
	ErrNotFound:       "Not Found",
	ErrMethodNotAllow: "Method Not Allowed",
	ErrUnprocessable:  "Unprocessable Entity",
	ErrInternalServer: "Internal Server Error",
	ErrBadGateway:     "Bad Gateway",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code ErrorCode) string {
	if msg, ok := CommonErrorMessages[code]; ok {
		return msg
	}
	return "Unknown Error"
}
