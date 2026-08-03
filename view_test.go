package rock

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeViewEngine 实现 rock.ViewEngine 接口，用于测试视图链路
type fakeViewEngine struct {
	ext string
	dir string
}

func (f *fakeViewEngine) Name() string        { return "fake" }
func (f *fakeViewEngine) Ext() string         { return f.ext }
func (f *fakeViewEngine) SetViewDir(d string) { f.dir = d }
func (f *fakeViewEngine) GetViewDir() string  { return f.dir }
func (f *fakeViewEngine) ExecuteWriter(w io.Writer, filename string, data interface{}) error {
	// 与 rock-pongo2 一致：引擎自己负责扩展名补全
	filename = EnsureTemplateName(filename, f)
	fmt.Fprintf(w, "render:%s:%v", filename, data)
	return nil
}

func TestRegisterViewAndHTML(t *testing.T) {
	app := New()
	fe := &fakeViewEngine{ext: ".html"}
	app.RegisterView(fe)

	app.Get("/", func(c Context) {
		c.HTML("home", M{"a": 1})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("模板渲染应 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type 应为 text/html, got %q", ct)
	}
	if string(body) != "render:home.html:map[a:1]" {
		t.Errorf("渲染输出不符, got %q", body)
	}
}

func TestHTMLWithoutEngine(t *testing.T) {
	app := New()
	app.Get("/", func(c Context) {
		c.HTML("home")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("未注册视图引擎应返回 500, got %d", resp.StatusCode)
	}
}

func TestAppView(t *testing.T) {
	app := New()
	fe := &fakeViewEngine{ext: ".html"}
	app.RegisterView(fe)

	var buf bytes.Buffer
	if err := app.View(&buf, "home", M{"a": 1}); err != nil {
		t.Fatalf("app.View 失败: %v", err)
	}
	if buf.String() != "render:home.html:map[a:1]" {
		t.Errorf("app.View 输出不符, got %q", buf.String())
	}
}

func TestAppViewMissingEngine(t *testing.T) {
	app := New()
	var buf bytes.Buffer
	if err := app.View(&buf, "home", nil); err == nil {
		t.Error("未注册视图引擎时 app.View 应报错")
	}
}

func TestViewRegisterAndRegistered(t *testing.T) {
	v := &View{}
	if v.Registered() {
		t.Error("初始不应 Registered")
	}
	fe := &fakeViewEngine{}
	v.Register(fe)
	if !v.Registered() {
		t.Error("注册后应 Registered")
	}
	if v.Engine != fe {
		t.Error("Engine 应为注册的引擎")
	}
}

func TestCtxViewEngine(t *testing.T) {
	app := New()
	fe := &fakeViewEngine{ext: ".html"}

	app.Get("/", func(c Context) {
		c.ViewEngine(fe) // 直接设置，不走 RegisterView 中间件
		c.HTML("home", nil)
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "render:home.html:<nil>" {
		t.Errorf("ctx.ViewEngine 设置后应渲染, got %q", body)
	}
}

func TestAppGetViewAndLogger(t *testing.T) {
	app := New()
	v := app.GetView()
	if v.Registered() {
		t.Error("未注册不应 Registered")
	}
	app.RegisterView(&fakeViewEngine{})
	v = app.GetView()
	if !v.Registered() {
		t.Error("注册后应 Registered")
	}
	if app.Logger() == nil {
		t.Error("app.Logger() 不应为 nil")
	}
}
