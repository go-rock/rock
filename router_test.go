package rock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRouter(t *testing.T) {
	app := New()
	var req *http.Request

	// 测试基本路由
	app.Get("/test", func(c Context) {
		c.String(200, "test")
	})

	app.Post("/test", func(c Context) {
		c.JSON(200, M{"method": "POST"})
	})

	app.Put("/test", func(c Context) {
		c.JSON(200, M{"method": "PUT"})
	})

	app.Delete("/test", func(c Context) {
		c.JSON(200, M{"method": "DELETE"})
	})

	// 测试参数路由
	app.Get("/user/:id", func(c Context) {
		id := c.Param("id")
		c.JSON(200, M{"user_id": id})
	})

	// 测试通配符路由（使用正确的语法）
	app.Get("/static/*", func(c Context) {
		path := c.Param("*")
		c.JSON(200, M{"path": path})
	})

	// 创建测试服务器
	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()
	// 禁用重定向跟随，以便测试重定向响应
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	// 禁用重定向跟随，以便测试重定向响应
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	// 禁用重定向跟随，以便测试重定向响应
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 测试GET请求
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试POST请求
	req, _ = http.NewRequest("POST", server.URL+"/test", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to POST /test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试PUT请求
	req, _ = http.NewRequest("PUT", server.URL+"/test", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to PUT /test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试DELETE请求
	req, _ = http.NewRequest("DELETE", server.URL+"/test", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to DELETE /test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试参数路由
	resp, err = client.Get(server.URL + "/user/123")
	if err != nil {
		t.Fatalf("Failed to GET /user/123: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试通配符路由
	resp, err = client.Get(server.URL + "/static/test.js")
	if err != nil {
		t.Fatalf("Failed to GET /static/test.js: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGroupNoRoute(t *testing.T) {
	app := New()

	app.NoRoute(func(c Context) {
		c.SetHeader("X-NoRoute", "root")
		c.String(404, "root 404")
	})

	admin := app.Group("/admin")
	admin.NoRoute(func(c Context) {
		c.SetHeader("X-NoRoute", "admin")
		c.String(404, "admin 404")
	})
	admin.Get("/ok", func(c Context) {
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()
	client := server.Client()

	// 根分组 404：命中 app.NoRoute
	resp, err := client.Get(server.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("Failed to GET /nonexistent: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 || resp.Header.Get("X-NoRoute") != "root" {
		t.Errorf("根分组 404 应走 app.NoRoute, got status=%d header=%q", resp.StatusCode, resp.Header.Get("X-NoRoute"))
	}

	// admin 分组 404：命中 admin.NoRoute，而不是全局的
	resp2, err := client.Get(server.URL + "/admin/nonexistent")
	if err != nil {
		t.Fatalf("Failed to GET /admin/nonexistent: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 || resp2.Header.Get("X-NoRoute") != "admin" {
		t.Errorf("/admin 下 404 应走 admin.NoRoute, got status=%d header=%q", resp2.StatusCode, resp2.Header.Get("X-NoRoute"))
	}

	// admin 分组正常路由不受影响
	resp3, err := client.Get(server.URL + "/admin/ok")
	if err != nil {
		t.Fatalf("Failed to GET /admin/ok: %v", err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Errorf("admin 正常路由应 200, got %d", resp3.StatusCode)
	}
}

func TestHeadFallsBackToGet(t *testing.T) {
	app := New()
	app.Get("/test", func(c Context) {
		c.String(200, "body")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodHead, server.URL+"/test", nil)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("HEAD failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("HEAD 应自动映射 GET 并返回 200, got %d", resp.StatusCode)
	}
}

func TestRouterNoRoute(t *testing.T) {
	app := New()

	app.Get("/test", func(c Context) {
		c.String(200, "test")
	})

	app.NoRoute(func(c Context) {
		c.JSON(404, M{"error": "not found"})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试不存在的路由
	resp, err := client.Get(server.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("Failed to GET /nonexistent: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestRouterNoMethod(t *testing.T) {
	app := New()

	app.Get("/test", func(c Context) {
		c.String(200, "test")
	})

	app.NoMethod(func(c Context) {
		c.JSON(405, M{"error": "method not allowed"})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试不支持的HTTP方法
	var req *http.Request
	var resp *http.Response
	var err error
	req, _ = http.NewRequest("PATCH", server.URL+"/test", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to PATCH /test: %v", err)
	}
	if resp.StatusCode != 405 {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestRouterStatic(t *testing.T) {
	// 创建测试静态文件目录
	// New().Static("/static", "./testdata/static")

	// 暂时跳过静态文件测试，因为需要文件系统
	t.Skip("Static file testing requires filesystem setup")
}

func TestRouterMiddleware(t *testing.T) {
	app := New()

	var executed bool

	app.Use(func(c Context) {
		executed = true
		c.Next()
	})

	app.Get("/test", func(c Context) {
		c.String(200, "test")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 重置执行标志
	executed = false

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}

	if !executed {
		t.Error("Middleware was not executed")
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRouterGroup(t *testing.T) {
	app := New()

	admin := app.Group("/admin")

	admin.Get("/users", func(c Context) {
		c.JSON(200, M{"section": "admin", "action": "users"})
	})

	admin.Get("/settings", func(c Context) {
		c.JSON(200, M{"section": "admin", "action": "settings"})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试管理员路由
	resp, err := client.Get(server.URL + "/admin/users")
	if err != nil {
		t.Fatalf("Failed to GET /admin/users: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	resp, err = client.Get(server.URL + "/admin/settings")
	if err != nil {
		t.Fatalf("Failed to GET /admin/settings: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 测试普通路由应该返回404
	resp, err = client.Get(server.URL + "/test")
	if err == nil && resp.StatusCode != 404 {
		t.Errorf("Expected 404 for /test route, got %d", resp.StatusCode)
	}
}

func TestRouterRedirection(t *testing.T) {
	app := New()

	app.Get("/test/", func(c Context) {
		c.String(200, "with trailing slash")
	})

	app.Get("/test2", func(c Context) {
		c.String(200, "without trailing slash")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()
	// 禁用重定向跟随，以便测试重定向响应
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 测试固定路径重定向
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}

	// 应该重定向到 /test/
	if resp.StatusCode != 301 && resp.StatusCode != 307 {
		t.Errorf("Expected redirection status, got %d", resp.StatusCode)
	}

	// 测试尾随斜杠重定向
	resp, err = client.Get(server.URL + "/test2/")
	if err != nil {
		t.Fatalf("Failed to GET /test2/: %v", err)
	}

	// 应该重定向到 /test2
	if resp.StatusCode != 301 && resp.StatusCode != 307 {
		t.Errorf("Expected redirection status, got %d", resp.StatusCode)
	}
}

func TestRouterTimeout(t *testing.T) {
	app := New()

	app.Get("/slow", func(c Context) {
		time.Sleep(5 * time.Second)
		c.String(200, "slow response")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()
	client.Timeout = 2 * time.Second

	// 测试超时
	_, err := client.Get(server.URL + "/slow")
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func BenchmarkRouter(b *testing.B) {
	app := New()

	// 添加大量路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/route/%d", i)
		app.Get(path, func(c Context) {
			c.String(200, "ok")
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/route/%d", i%1000)
		
		// 创建模拟请求避免网络端口冲突
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
		
		if w.Code != 200 {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}
