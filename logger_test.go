package rock

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	// 创建缓冲输出以捕获日志
	var buf bytes.Buffer
	app := New()

	// 设置日志输出到缓冲
	app.SetLoggerOutput(&buf)
	app.SetLogLevel(LevelDebug)

	// 测试基本日志功能
	app.logger.Debug("This is a debug message")
	app.logger.Info("This is an info message")
	app.logger.Warn("This is a warning message")
	app.logger.Error("This is an error message")

	output := buf.String()
	if output == "" {
		t.Error("Expected log output, got empty string")
	}

	// 验证输出包含预期内容
	if !bytes.Contains(buf.Bytes(), []byte("This is a debug message")) {
		t.Error("Expected debug message in output")
	}
}

func TestRequestLogging(t *testing.T) {
	// 创建缓冲输出以捕获请求日志
	var buf bytes.Buffer
	app := New()

	// 设置日志输出到缓冲并启用请求日志
	app.SetLoggerOutput(&buf)
	app.EnableRequestLog(true)
	app.SetLogLevel(LevelInfo)

	// 创建一个简单的路由
	app.Get("/test", func(c Context) {
		c.String(200, "ok")
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "TestAgent/1.0")

	// 执行请求
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// 验证请求被处理
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 验证请求日志被记录
	output := buf.String()
	if output == "" {
		t.Error("Expected request log output, got empty string")
	}

	if !bytes.Contains(buf.Bytes(), []byte("GET /test")) {
		t.Error("Expected GET /test in request log")
	}
}

func TestContextLogging(t *testing.T) {
	// 创建缓冲输出以捕获日志
	var buf bytes.Buffer
	app := New()

	// 设置日志输出到缓冲
	app.SetLoggerOutput(&buf)
	app.SetLogLevel(LevelInfo)

	// 创建一个使用Context日志的路由
	app.Get("/test", func(c Context) {
		c.LogDebug("Debug from context: %s", "test")
		c.LogInfo("Info from context: %s", "test")
		c.LogWarn("Warning from context: %s", "test")
		c.LogError("Error from context: %s", "test")
		c.String(200, "ok")
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 执行请求
	app.ServeHTTP(w, req)

	// 验证Context日志被记录
	output := buf.String()
	if output == "" {
		t.Error("Expected context log output, got empty string")
	}

	if !bytes.Contains(buf.Bytes(), []byte("Info from context: test")) {
		t.Error("Expected context info message in output")
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	app := New()
	app.SetLoggerOutput(&buf)

	// 测试不同的日志级别
	app.SetLogLevel(LevelError)
	app.logger.Debug("This should not appear")
	app.logger.Error("This should appear")

	if bytes.Contains(buf.Bytes(), []byte("This should not appear")) {
		t.Error("Debug message should not appear when level is Error")
	}

	if !bytes.Contains(buf.Bytes(), []byte("This should appear")) {
		t.Error("Error message should appear when level is Error")
	}
}

func TestLoggerWithCaller(t *testing.T) {
	var buf bytes.Buffer
	app := New()
	app.SetLoggerOutput(&buf)

	// 测试带调用者信息的日志
	app.logger.SetCallerInfo(true)
	app.logger.Info("Test message with caller")

	output := buf.String()
	if output == "" {
		t.Error("Expected log output with caller info")
	}
}
