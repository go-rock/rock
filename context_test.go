package rock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-rock/rock/binding"
)

func TestContextBasic(t *testing.T) {
	app := New()

	app.Get("/test", func(c Context) {
		// 测试设置数据
		c.Set("key", "value")
		if val, exists := c.Get("key"); !exists || val != "value" {
			t.Error("Failed to set/get data")
		}

		// 测试JSON响应
		c.JSON(200, M{
			"status": "ok",
			"method": c.GetMethod(),
			"path":   c.GetPath(),
		})
	})

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

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	// 验证响应内容
	body, _ := io.ReadAll(resp.Body)
	var result M
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", result["status"])
	}
}

func TestContextParams(t *testing.T) {
	app := New()

	app.Get("/user/:id", func(c Context) {
		id := c.Param("id")
		name, _ := c.GetQuery("name") // 从查询参数获取name

		c.JSON(200, M{
			"id":   id,
			"name": name,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/user/123?name=john")
	if err != nil {
		t.Fatalf("Failed to GET /user/123: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result M
	json.Unmarshal(body, &result)

	if result["id"] != "123" {
		t.Errorf("Expected id 123, got %v", result["id"])
	}
	if result["name"] != "john" {
		t.Errorf("Expected name john, got %v", result["name"])
	}
}

func TestContextQuery(t *testing.T) {
	app := New()

	app.Get("/search", func(c Context) {
		q := c.Query("q")
		limit := c.Query("limit")
		page := c.QueryInt("page")

		c.JSON(200, M{
			"query": q,
			"limit": limit,
			"page":  page,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/search?q=test&limit=10&page=2")
	if err != nil {
		t.Fatalf("Failed to GET /search: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result M
	json.Unmarshal(body, &result)

	if result["query"] != "test" {
		t.Errorf("Expected query test, got %v", result["query"])
	}
	if result["limit"] != "10" {
		t.Errorf("Expected limit 10, got %v", result["limit"])
	}
	// JSON 反序列化时数字会被解析为 float64
	if pageFloat, ok := result["page"].(float64); !ok || pageFloat != 2 {
		t.Errorf("Expected page 2, got %v", result["page"])
	}
}

func TestContextClientIP(t *testing.T) {
	app := New()

	app.Get("/ip", func(c Context) {
		ip := c.ClientIP()
		c.JSON(200, M{"ip": ip})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/ip")
	if err != nil {
		t.Fatalf("Failed to GET /ip: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result M
	json.Unmarshal(body, &result)

	// 测试客户端IP获取
	ip := result["ip"].(string)
	if ip == "" {
		t.Error("Failed to get client IP")
	}
}

func TestContextFormParsing(t *testing.T) {
	app := New()

	app.Post("/form", func(c Context) {
		if err := c.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
			c.JSON(400, M{"error": err.Error()})
			return
		}

		username := c.Query("username")
		email := c.Query("email")

		c.JSON(200, M{
			"username": username,
			"email":    email,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建表单数据
	form := url.Values{}
	form.Add("username", "testuser")
	form.Add("email", "test@example.com")

	resp, err := client.PostForm(server.URL+"/form", form)
	if err != nil {
		t.Fatalf("Failed to POST /form: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextMultipartForm(t *testing.T) {
	app := New()

	app.Post("/multipart", func(c Context) {
		if err := c.ParseMultipartForm(10 << 20); err != nil { // 10MB
			c.JSON(400, M{"error": err.Error()})
			return
		}

		// 测试表单字段
		name := c.Query("name")
		desc := c.Query("description")

		// 测试文件
		file, header, err := c.Request().FormFile("file")
		if err == nil {
			defer file.Close()
		}

		c.JSON(200, M{
			"name":        name,
			"description": desc,
			"file_name":   header.Filename,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文本字段
	writer.WriteField("name", "test file")
	writer.WriteField("description", "test description")

	// 添加文件
	fileWriter, _ := writer.CreateFormFile("file", "test.txt")
	fileWriter.Write([]byte("test file content"))

	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/multipart", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to POST /multipart: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextAbort(t *testing.T) {
	app := New()

	var middlewareExecuted bool
	var handlerExecuted bool

	app.Use(func(c Context) {
		middlewareExecuted = true
		c.Next()
		if c.StatusCode() == 200 {
			t.Error("Expected status to be changed by handler")
		}
	})

	app.Get("/abort", func(c Context) {
		handlerExecuted = true
		c.AbortWithStatusJSON(403, M{"error": "forbidden"})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/abort")
	if err != nil {
		t.Fatalf("Failed to GET /abort: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}

	if !middlewareExecuted {
		t.Error("Middleware was not executed")
	}

	if !handlerExecuted {
		t.Error("Handler was not executed")
	}
}

func TestContextRedirect(t *testing.T) {
	app := New()

	app.Get("/redirect", func(c Context) {
		c.Redirect("/target")
	})

	app.Get("/target", func(c Context) {
		c.String(200, "redirected")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试重定向
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // 不要自动跟随重定向
	}

	resp, err := client.Get(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("Failed to GET /redirect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 302 && resp.StatusCode != 307 {
		t.Errorf("Expected redirection status, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if !strings.Contains(location, "/target") {
		t.Errorf("Expected redirect to /target, got %s", location)
	}
}

func TestContextWriter(t *testing.T) {
	app := New()

	app.Get("/write", func(c Context) {
		// 测试直接写入
		n, err := c.Write([]byte("hello"))
		if err != nil {
			t.Errorf("Failed to write: %v", err)
		}
		if n != 5 {
			t.Errorf("Expected to write 5 bytes, got %d", n)
		}
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/write")
	if err != nil {
		t.Fatalf("Failed to GET /write: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(body))
	}
}

func TestContextFail(t *testing.T) {
	app := New()

	app.Get("/fail", func(c Context) {
		c.Fail(400, "bad request")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/fail")
	if err != nil {
		t.Fatalf("Failed to GET /fail: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestContextString(t *testing.T) {
	app := New()

	app.Get("/string", func(c Context) {
		c.String(200, "Hello %s", "World")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/string")
	if err != nil {
		t.Fatalf("Failed to GET /string: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", string(body))
	}
}

func TestContextXML(t *testing.T) {
	app := New()

	app.Get("/xml", func(c Context) {
		type User struct {
			Name string `xml:"name"`
			Age  int    `xml:"age"`
		}

		user := User{Name: "John", Age: 30}
		c.XML(200, user)
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	resp, err := client.Get(server.URL + "/xml")
	if err != nil {
		t.Fatalf("Failed to GET /xml: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "xml") {
		t.Errorf("Expected XML content type, got %s", ct)
	}
}

func TestContextContextPool(t *testing.T) {
	app := New()

	var firstRequestID string
	var secondRequestID string

	app.Use(func(c Context) {
		// 设置请求ID
		requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
		c.Set("request_id", requestID)
		c.Next()

		if c.Request().URL.Path == "/first" {
			if val, exists := c.Get("request_id"); exists {
				firstRequestID = val.(string)
			}
		} else if c.Request().URL.Path == "/second" {
			if val, exists := c.Get("request_id"); exists {
				secondRequestID = val.(string)
			}
		}
	})

	app.Get("/first", func(c Context) {
		c.String(200, "first")
	})

	app.Get("/second", func(c Context) {
		c.String(200, "second")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 第一个请求
	resp1, _ := client.Get(server.URL + "/first")
	resp1.Body.Close()

	// 第二个请求
	resp2, _ := client.Get(server.URL + "/second")
	resp2.Body.Close()

	// 验证请求ID不同（对象池正确工作）
	if firstRequestID == "" || secondRequestID == "" {
		t.Error("Request IDs were not set")
	}

	if firstRequestID == secondRequestID {
		t.Error("Context pool did not work correctly - request IDs are the same")
	}
}

func TestContextMustMethods(t *testing.T) {
	app := New()

	app.Post("/must", func(c Context) {
		// 测试MustPostInt
		page := c.MustPostInt("page", 1)
		if page != 10 {
			t.Errorf("Expected page 10, got %d", page)
		}

		// 测试MustPostString
		name := c.MustPostString("name", "default")
		if name != "test" {
			t.Errorf("Expected name test, got %s", name)
		}

		// 测试MustPostInt - 应该从POST表单数据中获取
		limit := c.MustPostInt("limit", 20)
		if limit != 5 {
			t.Errorf("Expected limit 5, got %d", limit)
		}

		// 测试默认值
		defaultPage := c.MustPostInt("missing", 99)
		if defaultPage != 99 {
			t.Errorf("Expected default page 99, got %d", defaultPage)
		}

		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建表单数据
	form := url.Values{}
	form.Add("page", "10")
	form.Add("name", "test")
	form.Add("limit", "5")

	resp, err := client.PostForm(server.URL+"/must", form)
	if err != nil {
		t.Fatalf("Failed to POST /must: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextPoolNoValueLeak(t *testing.T) {
	app := New()

	// 请求 A：往 ctx.Values() 写入数据（模拟中间件存用户ID/鉴权信息等）
	reqA := httptest.NewRequest("GET", "/a", nil)
	wA := httptest.NewRecorder()
	cA := app.createContext(wA, reqA)
	cA.Values().Set("user_id", "42")
	cA.ViewData("title", "secret-page")

	// 请求 A 结束，放回池
	app.pool.Put(cA)

	// 请求 B：复用同一个 pool 里的 Ctx（sync.Pool 会优先返回刚 Put 的对象）
	reqB := httptest.NewRequest("GET", "/b", nil)
	wB := httptest.NewRecorder()
	cB := app.createContext(wB, reqB)

	if got := cB.Values().Get("user_id"); got != nil {
		t.Fatalf("values 跨请求泄漏! 请求B看到了请求A写入的 user_id=%v", got)
	}
	if got := cB.GetViewData(); got != nil {
		t.Fatalf("ViewData 跨请求泄漏! 请求B看到了请求A的 view data=%v", got)
	}
}

func TestContextStatusLazyWriteHeader(t *testing.T) {
	app := New()

	// 先设置状态码再改，最终以 JSON 的参数为准，且只写一次头
	app.Get("/override", func(c Context) {
		c.Status(http.StatusNotFound)
		c.JSON(http.StatusOK, M{"ok": true})
	})

	// 先 Status 再裸写 body，应以 Status 设置的状态码发送
	app.Get("/raw", func(c Context) {
		c.Status(http.StatusCreated)
		c.Write([]byte("created"))
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/override")
	if err != nil {
		t.Fatalf("Failed to GET /override: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("JSON 的 code 参数应为最终状态码 200, got %d", resp.StatusCode)
	}

	resp2, err := server.Client().Get(server.URL + "/raw")
	if err != nil {
		t.Fatalf("Failed to GET /raw: %v", err)
	}
	body, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("Status+Write 应发送 201, got %d", resp2.StatusCode)
	}
	if string(body) != "created" {
		t.Errorf("Expected body 'created', got %q", string(body))
	}
}

func TestContextMustAccessors(t *testing.T) {
	app := New()
	app.Get("/users/:id", func(c Context) {
		c.JSON(200, M{
			"id": c.MustParamInt("id", 0),
			"q1": c.MustQueryInt("n", 5),
			"q2": c.MustQueryString("s", "def"),
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	// 正常参数
	resp, _ := server.Client().Get(server.URL + "/users/42?n=7&s=hello")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `"id":42`) || !strings.Contains(string(body), `"q1":7`) || !strings.Contains(string(body), `"q2":"hello"`) {
		t.Errorf("正常参数解析错误: %s", body)
	}

	// 非法参数 → 回退默认值
	resp2, _ := server.Client().Get(server.URL + "/users/abc")
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if !strings.Contains(string(body2), `"id":0`) || !strings.Contains(string(body2), `"q1":5`) || !strings.Contains(string(body2), `"q2":"def"`) {
		t.Errorf("非法参数应回退默认值: %s", body2)
	}
}

func TestContextAttachmentInline(t *testing.T) {
	app := New()
	app.Get("/att", func(c Context) {
		c.Attachment(strings.NewReader("file-data"), "a.txt")
	})
	app.Get("/inl", func(c Context) {
		c.Inline(strings.NewReader("file-data"), "a.txt")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, _ := server.Client().Get(server.URL + "/att")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "file-data" {
		t.Errorf("Attachment body 错误: %q", body)
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Disposition"), "attachment;") {
		t.Errorf("Attachment 应设置 Content-Disposition, got %q", resp.Header.Get("Content-Disposition"))
	}

	resp2, _ := server.Client().Get(server.URL + "/inl")
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if string(body2) != "file-data" {
		t.Errorf("Inline body 错误: %q", body2)
	}
	if !strings.HasPrefix(resp2.Header.Get("Content-Disposition"), "inline;") {
		t.Errorf("Inline 应设置 Content-Disposition, got %q", resp2.Header.Get("Content-Disposition"))
	}
}

func TestContextApplicationAndView(t *testing.T) {
	app := New()
	app.Get("/", func(c Context) {
		if c.Application() != app {
			t.Error("Application() 应返回 app")
		}
		if c.GetView() != app.GetView() {
			t.Error("GetView() 应返回 app 的 view")
		}
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()
	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("应 200, got %d", resp.StatusCode)
	}
}

func TestResetRequestClearsState(t *testing.T) {
	c := &Ctx{}
	c.newContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/a", nil))
	c.values.Set("k", "v")
	c.data = M{"x": 1}
	c.params = Map{"p": "1"}
	c.handlers = []HandlerFunc{func(Context) {}}
	c.statusCode = 201

	c.ResetRequest(httptest.NewRequest("GET", "/b", nil))

	if c.Path != "/b" {
		t.Errorf("Path 应重置为 /b, got %q", c.Path)
	}
	if c.values.Get("k") != nil {
		t.Error("ResetRequest 应清空 values")
	}
	if c.data != nil || c.params != nil || c.handlers != nil {
		t.Error("ResetRequest 应清空 data/params/handlers")
	}
	if c.statusCode != http.StatusOK {
		t.Errorf("statusCode 应重置为 200, got %d", c.statusCode)
	}
}

func TestContextDataMethods(t *testing.T) {
	app := New()
	app.Get("/", func(c Context) {
		c.SetData(M{"x": 1})
		if d := c.Data(); d["x"] != 1 {
			t.Errorf("SetData/Data 错误: %v", d)
		}
		c.Set("y", 2)
		if v, ok := c.Get("y"); !ok || v != 2 {
			t.Errorf("Set/Get 错误: %v %v", v, ok)
		}
		c.String(200, "ok")
	})

	server := httptest.NewServer(app)
	defer server.Close()
	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("应 200, got %d", resp.StatusCode)
	}
}

func TestClientIPTrustProxy(t *testing.T) {
	app := New()
	app.Get("/ip", func(c Context) { c.String(200, "%s", c.ClientIP()) })

	server := httptest.NewServer(app)
	defer server.Close()

	// 默认不信任代理头，应使用 RemoteAddr 而非 X-Forwarded-For
	req, _ := http.NewRequest("GET", server.URL+"/ip", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("Failed to GET /ip: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) == "203.0.113.7" {
		t.Error("默认应忽略 X-Forwarded-For，使用 RemoteAddr")
	}

	// 开启信任后应使用代理头
	app.SetTrustProxy(true)
	req2, _ := http.NewRequest("GET", server.URL+"/ip", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.7")
	resp2, err := server.Client().Do(req2)
	if err != nil {
		t.Fatalf("Failed to GET /ip: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if string(body2) != "203.0.113.7" {
		t.Errorf("开启 TrustProxy 后应使用 X-Forwarded-For, got %q", body2)
	}
}

func TestGetQueryPresence(t *testing.T) {
	app := New()
	app.Get("/q", func(c Context) {
		val, ok := c.GetQuery("foo")
		c.JSON(200, M{"val": val, "ok": ok})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	// ?foo= 应视为"存在但为空"，而不是不存在
	resp, err := server.Client().Get(server.URL + "/q?foo=")
	if err != nil {
		t.Fatalf("Failed to GET /q: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Val string `json:"val"`
		Ok  bool   `json:"ok"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if !result.Ok || result.Val != "" {
		t.Errorf("?foo= 应返回 (\"\", true), got (val=%q, ok=%v)", result.Val, result.Ok)
	}
}

func TestDecodeBodyReusable(t *testing.T) {
	app := New()
	app.Post("/bind", func(c Context) {
		type Req struct {
			Name string `json:"name"`
		}
		var a, b Req
		if err := c.ShouldBind(&a); err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}
		// 同一请求内第二次 ShouldBind 不应因 body 已消费而失败
		if err := c.ShouldBind(&b); err != nil {
			c.JSON(400, M{"error": "second bind: " + err.Error()})
			return
		}
		if a.Name != "x" || b.Name != "x" {
			c.JSON(400, M{"error": "data mismatch"})
			return
		}
		c.JSON(200, M{"ok": true})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Post(server.URL+"/bind", "application/json", strings.NewReader(`{"name":"x"}`))
	if err != nil {
		t.Fatalf("Failed to POST /bind: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("二次 ShouldBind 应成功, got %d: %s", resp.StatusCode, body)
	}
}

func TestWriteErrorValidation(t *testing.T) {
	app := New()
	app.Get("/validate", func(c Context) {
		req := struct {
			Name string `binding:"required"`
		}{}
		// Name 为空 → binding.Validate 返回 validator.ValidationErrors
		if err := binding.Validate(&req); err != nil {
			WriteError(c, 400, err)
			return
		}
		c.JSON(200, M{"ok": true})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/validate")
	if err != nil {
		t.Fatalf("Failed to GET /validate: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 400 {
		t.Errorf("校验错误应返回 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "Validation failed") {
		t.Errorf("应返回 Validation failed, got %s", body)
	}
	if strings.Contains(string(body), "Internal Server Error") {
		t.Error("校验错误不应被误报为 Internal Server Error")
	}
}

func BenchmarkContext(b *testing.B) {
	app := New()

	handler := func(c Context) {
		c.JSON(200, M{
			"path":   c.GetPath(),
			"method": c.GetMethod(),
			"status": c.StatusCode(),
		})
	}

	app.Get("/test", handler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建模拟请求避免网络端口冲突
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
	}
}
