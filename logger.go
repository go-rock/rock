package rock

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kataras/golog"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String 实现Stringer接口
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志器接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	// 设置日志级别
	SetLevel(level LogLevel)
	GetLevel() LogLevel

	// 添加输出目标
	AddOutput(output io.Writer)

	// 请求日志
	RequestLog(method, path, ip, userAgent string, statusCode int, latency time.Duration)
}

// RockLogger Rock框架的日志器实现
type RockLogger struct {
	logger     *golog.Logger
	level      LogLevel
	outputs    []io.Writer
	mu         sync.RWMutex
	requestLog bool
}

// NewLogger 创建新的日志器
func NewLogger() *RockLogger {
	logger := golog.New()
	logger.Level = golog.DebugLevel

	// 设置默认输出
	outputs := []io.Writer{os.Stdout}

	rl := &RockLogger{
		logger:  logger,
		level:   LevelDebug, // 统一设置为DEBUG级别
		outputs: outputs,
	}

	return rl
}

// NewLoggerWithConfig 使用配置创建日志器
func NewLoggerWithConfig(level LogLevel, outputs []io.Writer, enableRequestLog bool) *RockLogger {
	logger := golog.New()

	// 转换日志级别
	switch level {
	case LevelDebug:
		logger.Level = golog.DebugLevel
	case LevelInfo:
		logger.Level = golog.InfoLevel
	case LevelWarn:
		logger.Level = golog.WarnLevel
	case LevelError:
		logger.Level = golog.ErrorLevel
	case LevelFatal:
		logger.Level = golog.FatalLevel
	}

	rl := &RockLogger{
		logger:     logger,
		level:      level,
		outputs:    make([]io.Writer, 0),
		requestLog: enableRequestLog,
	}

	// 设置输出 - 避免重复添加
	if outputs != nil && len(outputs) > 0 {
		// 清空默认输出，只添加指定的输出
		logger.SetOutput(io.Discard) // 清空默认输出
		for _, output := range outputs {
			rl.outputs = append(rl.outputs, output)
			logger.AddOutput(output)
		}
	}

	return rl
}

// Debug 调试日志
func (rl *RockLogger) Debug(args ...interface{}) {
	if rl.shouldLog(LevelDebug) {
		rl.logger.Debug(args...)
	}
}

// Debugf 格式化调试日志
func (rl *RockLogger) Debugf(format string, args ...interface{}) {
	if rl.shouldLog(LevelDebug) {
		rl.logger.Debugf(format, args...)
	}
}

// Info 信息日志
func (rl *RockLogger) Info(args ...interface{}) {
	if rl.shouldLog(LevelInfo) {
		rl.logger.Info(args...)
	}
}

// Infof 格式化信息日志
func (rl *RockLogger) Infof(format string, args ...interface{}) {
	if rl.shouldLog(LevelInfo) {
		rl.logger.Infof(format, args...)
	}
}

// Warn 警告日志
func (rl *RockLogger) Warn(args ...interface{}) {
	if rl.shouldLog(LevelWarn) {
		rl.logger.Warn(args...)
	}
}

// Warnf 格式化警告日志
func (rl *RockLogger) Warnf(format string, args ...interface{}) {
	if rl.shouldLog(LevelWarn) {
		rl.logger.Warnf(format, args...)
	}
}

// Error 错误日志
func (rl *RockLogger) Error(args ...interface{}) {
	if rl.shouldLog(LevelError) {
		rl.logger.Error(args...)
	}
}

// Errorf 格式化错误日志
func (rl *RockLogger) Errorf(format string, args ...interface{}) {
	if rl.shouldLog(LevelError) {
		rl.logger.Errorf(format, args...)
	}
}

// Fatal 致命错误日志
func (rl *RockLogger) Fatal(args ...interface{}) {
	if rl.shouldLog(LevelFatal) {
		rl.logger.Fatal(args...)
	}
}

// Fatalf 格式化致命错误日志
func (rl *RockLogger) Fatalf(format string, args ...interface{}) {
	if rl.shouldLog(LevelFatal) {
		rl.logger.Fatalf(format, args...)
	}
}

// SetLevel 设置日志级别
func (rl *RockLogger) SetLevel(level LogLevel) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.level = level

	// 同步到golog
	switch level {
	case LevelDebug:
		rl.logger.Level = golog.DebugLevel
	case LevelInfo:
		rl.logger.Level = golog.InfoLevel
	case LevelWarn:
		rl.logger.Level = golog.WarnLevel
	case LevelError:
		rl.logger.Level = golog.ErrorLevel
	case LevelFatal:
		rl.logger.Level = golog.FatalLevel
	}
}

// GetLevel 获取日志级别
func (rl *RockLogger) GetLevel() LogLevel {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.level
}

// AddOutput 添加输出目标
func (rl *RockLogger) AddOutput(output io.Writer) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// 检查是否已经添加过相同的输出
	for _, existing := range rl.outputs {
		if existing == output {
			return // 避免重复添加相同的输出
		}
	}
	
	rl.outputs = append(rl.outputs, output)
	rl.logger.AddOutput(output)
}

// SetOutputs 设置输出目标
func (rl *RockLogger) SetOutputs(outputs ...io.Writer) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// 清空现有输出
	rl.outputs = outputs
	// 重置logger的输出
	rl.logger.SetOutput(io.Discard)
	
	// 添加新的输出
	for _, output := range outputs {
		rl.logger.AddOutput(output)
	}
}

// EnableRequestLog 启用或禁用请求日志
func (rl *RockLogger) EnableRequestLog(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.requestLog = enabled
}

// SetCallerInfo 设置是否包含调用者信息
func (rl *RockLogger) SetCallerInfo(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// 这里可以扩展以支持调用者信息
	// 目前只是占位符实现
}

// shouldLog 检查是否应该记录此级别的日志
func (rl *RockLogger) shouldLog(level LogLevel) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return level >= rl.level
}

// RequestLog 记录请求日志
func (rl *RockLogger) RequestLog(method, path, ip, userAgent string, statusCode int, latency time.Duration) {
	if !rl.requestLog {
		return
	}

	// 根据状态码选择日志级别
	var level LogLevel
	switch {
	case statusCode >= 500:
		level = LevelError
	case statusCode >= 400:
		level = LevelWarn
	default:
		level = LevelInfo
	}

	if rl.shouldLog(level) {
		msg := fmt.Sprintf("%s %s %s %d %s %s",
			method,
			path,
			ip,
			statusCode,
			latency.String(),
			userAgent,
		)

		switch level {
		case LevelError:
			rl.logger.Error(msg)
		case LevelWarn:
			rl.logger.Warn(msg)
		default:
			rl.logger.Info(msg)
		}
	}
}

// GetCaller 获取调用者信息
func GetCaller() (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "", 0, ""
	}

	function = runtime.FuncForPC(pc).Name()
	function = extractFunctionName(function)
	file = extractFileName(file)

	return file, line, function
}

// extractFunctionName 提取函数名
func extractFunctionName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// extractFileName 提取文件名
func extractFileName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// WithCaller 带调用者信息的日志记录器
type LoggerWithCaller struct {
	logger Logger
}

// Debug 带调用者信息的调试日志
func (lc *LoggerWithCaller) Debug(args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	finalArgs := make([]interface{}, 0, len(args)+1)
	finalArgs = append(finalArgs, prefix)
	finalArgs = append(finalArgs, args...)

	lc.logger.Debug(finalArgs...)
}

// Debugf 带调用者信息的格式化调试日志
func (lc *LoggerWithCaller) Debugf(format string, args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	lc.logger.Debugf(prefix+format, args...)
}

// Info 带调用者信息的信息日志
func (lc *LoggerWithCaller) Info(args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	finalArgs := make([]interface{}, 0, len(args)+1)
	finalArgs = append(finalArgs, prefix)
	finalArgs = append(finalArgs, args...)

	lc.logger.Info(finalArgs...)
}

// Infof 带调用者信息的格式化信息日志
func (lc *LoggerWithCaller) Infof(format string, args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	lc.logger.Infof(prefix+format, args...)
}

// Warn 带调用者信息的警告日志
func (lc *LoggerWithCaller) Warn(args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	finalArgs := make([]interface{}, 0, len(args)+1)
	finalArgs = append(finalArgs, prefix)
	finalArgs = append(finalArgs, args...)

	lc.logger.Warn(finalArgs...)
}

// Warnf 带调用者信息的格式化警告日志
func (lc *LoggerWithCaller) Warnf(format string, args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	lc.logger.Warnf(prefix+format, args...)
}

// Error 带调用者信息的错误日志
func (lc *LoggerWithCaller) Error(args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	finalArgs := make([]interface{}, 0, len(args)+1)
	finalArgs = append(finalArgs, prefix)
	finalArgs = append(finalArgs, args...)

	lc.logger.Error(finalArgs...)
}

// Errorf 带调用者信息的格式化错误日志
func (lc *LoggerWithCaller) Errorf(format string, args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	lc.logger.Errorf(prefix+format, args...)
}

// Fatal 带调用者信息的致命错误日志
func (lc *LoggerWithCaller) Fatal(args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	finalArgs := make([]interface{}, 0, len(args)+1)
	finalArgs = append(finalArgs, prefix)
	finalArgs = append(finalArgs, args...)

	lc.logger.Fatal(finalArgs...)
}

// Fatalf 带调用者信息的格式化致命错误日志
func (lc *LoggerWithCaller) Fatalf(format string, args ...interface{}) {
	file, line, function := GetCaller()
	prefix := fmt.Sprintf("[%s:%d %s] ", file, line, function)

	lc.logger.Fatalf(prefix+format, args...)
}

// SetLevel 设置日志级别
func (lc *LoggerWithCaller) SetLevel(level LogLevel) {
	lc.logger.SetLevel(level)
}

// GetLevel 获取日志级别
func (lc *LoggerWithCaller) GetLevel() LogLevel {
	return lc.logger.GetLevel()
}

// AddOutput 添加输出目标
func (lc *LoggerWithCaller) AddOutput(output io.Writer) {
	lc.logger.AddOutput(output)
}

// RequestLog 记录请求日志
func (lc *LoggerWithCaller) RequestLog(method, path, ip, userAgent string, statusCode int, latency time.Duration) {
	lc.logger.RequestLog(method, path, ip, userAgent, statusCode, latency)
}

// WrapLoggerWithCaller 包装日志器以包含调用者信息
func WrapLoggerWithCaller(logger Logger) *LoggerWithCaller {
	return &LoggerWithCaller{
		logger: logger,
	}
}

// 默认全局日志器
var defaultLogger *RockLogger

func init() {
	defaultLogger = NewLogger()
}

// SetDefaultLogger 设置默认日志器
func SetDefaultLogger(logger *RockLogger) {
	defaultLogger = logger
}

// GetDefaultLogger 获取默认日志器
func GetDefaultLogger() *RockLogger {
	return defaultLogger
}

// 便捷函数

// Debug 调试日志
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}
