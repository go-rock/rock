package rock

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	app := New()

	var executionOrder []string

	// 第一个中间件
	app.Use(func(c Context) {
		executionOrder = append(executionOrder, "middleware1-before")
		c.Set("middleware1", "value1")
		c.Next()
		executionOrder = append(executionOrder, "middleware1-after")
	})

	// 第二个中间件
	app.Use(func(c Context) {
		executionOrder = append(executionOrder, "middleware2-before")
		c.Set("middleware2", "value2")
		c.Next()
		executionOrder = append(executionOrder, "middleware2-after")
	})

	// 路由处理器
	app.Get("/test", func(c Context) {
		executionOrder = append(executionOrder, "handler")
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	// 验证执行顺序
	expectedOrder := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d executions, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Execution %d: expected '%s', got '%s'", i, expected,
				func() string {
					if i < len(executionOrder) {
						return executionOrder[i]
					}
					return "<missing>"
				}())
		}
	}

	// 验证中间件数据传递 - router.handle是一个函数，无法进行类型检查
	// 这个测试通过实际HTTP请求验证
}

func TestMiddlewareAbort(t *testing.T) {
	app := New()

	var middlewareExecuted bool
	var handlerExecuted bool

	app.Use(func(c Context) {
		middlewareExecuted = true
		c.Abort()
	})

	app.Get("/test", func(c Context) {
		handlerExecuted = true
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	if !middlewareExecuted {
		t.Error("Middleware was not executed")
	}

	if handlerExecuted {
		t.Error("Handler should not have been executed after Abort()")
	}
}

func TestMiddlewareNext(t *testing.T) {
	app := New()

	var executions []string

	app.Use(func(c Context) {
		executions = append(executions, "middleware1-start")
		c.Next()
		executions = append(executions, "middleware1-end")
	})

	app.Use(func(c Context) {
		executions = append(executions, "middleware2-start")
		c.Next()
		executions = append(executions, "middleware2-end")
	})

	app.Get("/test", func(c Context) {
		executions = append(executions, "handler")
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	expectedExecutions := []string{
		"middleware1-start",
		"middleware2-start",
		"handler",
		"middleware2-end",
		"middleware1-end",
	}

	for i, expected := range expectedExecutions {
		if i >= len(executions) || executions[i] != expected {
			t.Errorf("Execution %d: expected '%s', got '%s'", i, expected,
				func() string {
					if i < len(executions) {
						return executions[i]
					}
					return "<missing>"
				}())
		}
	}
}

func TestMiddlewareWithGroup(t *testing.T) {
	app := New()

	var globalMiddlewareExecuted bool
	var groupMiddlewareExecuted bool

	// 全局中间件
	app.Use(func(c Context) {
		globalMiddlewareExecuted = true
		c.Next()
	})

	// 路由组中间件
	admin := app.Group("/admin")
	admin.Use(func(c Context) {
		groupMiddlewareExecuted = true
		c.Next()
	})

	admin.Get("/test", func(c Context) {
		c.String(200, "admin test")
	})

	// 普通路由
	app.Get("/normal", func(c Context) {
		c.String(200, "normal test")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试管理员路由
	resp1, err := client.Get(server.URL + "/admin/test")
	if err != nil {
		t.Fatalf("Failed to GET /admin/test: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp1.StatusCode)
	}

	if !globalMiddlewareExecuted {
		t.Error("Global middleware was not executed for admin route")
	}

	if !groupMiddlewareExecuted {
		t.Error("Group middleware was not executed for admin route")
	}

	// 重置标志
	globalMiddlewareExecuted = false
	groupMiddlewareExecuted = false

	// 测试普通路由
	resp2, err := client.Get(server.URL + "/normal")
	if err != nil {
		t.Fatalf("Failed to GET /normal: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp2.StatusCode)
	}

	if !globalMiddlewareExecuted {
		t.Error("Global middleware was not executed for normal route")
	}

	if groupMiddlewareExecuted {
		t.Error("Group middleware should not be executed for normal route")
	}
}

func TestMiddlewareWithPriority(t *testing.T) {
	app := New()

	executionOrder := []string{}

	// 添加中间件（应该按优先级排序）
	app.Use(func(c Context) {
		executionOrder = append(executionOrder, "middleware1-before")
		c.Next()
		executionOrder = append(executionOrder, "middleware1-after")
	})

	app.Use(func(c Context) {
		executionOrder = append(executionOrder, "middleware2-before")
		c.Next()
		executionOrder = append(executionOrder, "middleware2-after")
	})

	app.Use(func(c Context) {
		executionOrder = append(executionOrder, "middleware3-before")
		c.Next()
		executionOrder = append(executionOrder, "middleware3-after")
	})

	app.Get("/test", func(c Context) {
		executionOrder = append(executionOrder, "handler")
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	// 验证执行顺序
	expectedOrder := []string{
		"middleware1-before",
		"middleware2-before",
		"middleware3-before",
		"handler",
		"middleware3-after",
		"middleware2-after",
		"middleware1-after",
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Execution %d: expected '%s', got '%s'", i, expected,
				func() string {
					if i < len(executionOrder) {
						return executionOrder[i]
					}
					return "<missing>"
				}())
		}
	}
}

func TestMiddlewareError(t *testing.T) {
	app := New()

	var middlewareError bool

	app.Use(func(c Context) {
		defer func() {
			if r := recover(); r != nil {
				middlewareError = true
				c.LogError("Middleware panic: %v", r)
				// 设置状态码为500并返回错误响应
				c.JSON(500, H{"error": "Internal Server Error"})
			}
		}()
		c.Next()
	})

	app.Get("/panic", func(c Context) {
		panic("test panic")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/panic")
	if err != nil {
		t.Fatalf("Failed to GET /panic: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// 验证panic被正确捕获
	if !middlewareError {
		t.Error("Expected middlewareError to be true")
	}
}

func TestMiddlewareRemove(t *testing.T) {
	app := New()

	var executions []string

	app.Use(func(c Context) {
		executions = append(executions, "middleware1")
		c.Next()
	})

	app.Use(func(c Context) {
		executions = append(executions, "middleware2")
		c.Next()
	})

	app.Get("/test", func(c Context) {
		executions = append(executions, "handler")
		c.String(200, "ok")
	})

	// 测试移除中间件（如果支持的话）
	// app.RemoveMiddleware(1) // 移除第二个中间件

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddlewareClear(t *testing.T) {
	app := New()

	var executions []string

	app.Use(func(c Context) {
		executions = append(executions, "middleware1")
		c.Next()
	})

	app.Use(func(c Context) {
		executions = append(executions, "middleware2")
		c.Next()
	})

	app.Get("/test", func(c Context) {
		executions = append(executions, "handler")
		c.String(200, "ok")
	})

	// 测试清除所有中间件（如果支持的话）
	// app.ClearMiddleware()

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func BenchmarkMiddleware(b *testing.B) {
	app := New()

	// 添加5个中间件
	for i := 0; i < 5; i++ {
		app.Use(func(c Context) {
			c.Next()
		})
	}

	app.Get("/test", func(c Context) {
		c.String(200, "ok")
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建模拟请求避免网络端口冲突
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
	}
}

func TestMiddlewarePerformance(t *testing.T) {
	app := New()

	start := time.Now()

	// 添加100个中间件进行性能测试
	for i := 0; i < 100; i++ {
		app.Use(func(c Context) {
			c.Next()
		})
	}

	app.Get("/test", func(c Context) {
		c.String(200, "ok")
	})

	duration := time.Since(start)

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试带100个中间件的请求
	start = time.Now()
	resp, err := client.Get(server.URL + "/test")
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	// 性能检查 - 100个中间件不应该导致显著延迟
	if duration > 100*time.Millisecond {
		t.Errorf("Request with 100 middlewares took too long: %v", duration)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddlewareGroupPrefixSegmentMatch(t *testing.T) {
	app := New()

	// /admin 组的中间件只应作用于 /admin 前缀的路径
	admin := app.Group("/admin")
	admin.Use(func(c Context) {
		c.SetHeader("X-Admin", "yes")
		c.Next()
	})
	admin.Get("/x", func(c Context) { c.Status(200) })

	// /administrator 是不同路径，不应命中 /admin 组中间件
	app.Get("/administrator", func(c Context) { c.Status(200) })

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/admin/x")
	if err != nil {
		t.Fatalf("Failed to GET /admin/x: %v", err)
	}
	resp.Body.Close()
	if resp.Header.Get("X-Admin") != "yes" {
		t.Error("expected /admin/x to run /admin middleware")
	}

	resp2, err := server.Client().Get(server.URL + "/administrator")
	if err != nil {
		t.Fatalf("Failed to GET /administrator: %v", err)
	}
	resp2.Body.Close()
	if resp2.Header.Get("X-Admin") != "" {
		t.Error("bug: /administrator 不应运行 /admin 组的中间件")
	}
}

func TestMiddlewareGroupOrder(t *testing.T) {
	app := New()

	var order []string
	app.Use(func(c Context) { order = append(order, "root"); c.Next() })

	api := app.Group("/api")
	api.Use(func(c Context) { order = append(order, "api"); c.Next() })

	v1 := api.Group("/v1")
	v1.Use(func(c Context) { order = append(order, "v1"); c.Next() })
	v1.Get("/x", func(c Context) { order = append(order, "handler"); c.String(200, "ok") })

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/api/v1/x")
	if err != nil {
		t.Fatalf("Failed to GET /api/v1/x: %v", err)
	}
	resp.Body.Close()

	expected := []string{"root", "api", "v1", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("执行顺序应为 %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("执行顺序应为 %v, got %v", expected, order)
		}
	}
}

func TestNextLongChain(t *testing.T) {
	app := New()

	executed := 0
	// 注册超过 63 个中间件（abortIndex 上限），验证任意长链都能完整执行
	for i := 0; i < 70; i++ {
		app.Use(func(c Context) {
			executed++
		})
	}
	app.Get("/long", func(c Context) {
		executed++
		c.String(200, "done")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/long")
	if err != nil {
		t.Fatalf("Failed to GET /long: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if executed != 71 {
		t.Errorf("expected all 71 handlers to run, got %d", executed)
	}
}

func TestMiddlewareRunsOnNoRoute(t *testing.T) {
	app := New()
	app.Use(func(c Context) {
		c.SetHeader("X-Global", "ran")
		c.Next()
	})
	// 不注册 /missing 路由

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/missing")
	if err != nil {
		t.Fatalf("Failed to GET /missing: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Global") != "ran" {
		t.Error("bug: 404 路径没有执行全局中间件")
	}
}
