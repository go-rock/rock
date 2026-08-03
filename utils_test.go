package rock

import (
	"net/http"
	"testing"
)

func TestEnsureTemplateName(t *testing.T) {
	e := &fakeViewEngine{ext: ".html"}

	if got := EnsureTemplateName("home", e); got != "home.html" {
		t.Errorf("应补扩展名, got %q", got)
	}
	if got := EnsureTemplateName("home.html", e); got != "home.html" {
		t.Errorf("已有扩展名不应重复, got %q", got)
	}
	if got := EnsureTemplateName("/admin/login", e); got != "admin/login.html" {
		t.Errorf("应剥前导斜杠并补扩展名, got %q", got)
	}
	if got := EnsureTemplateName("", e); got != "" {
		t.Errorf("空名称应返回空, got %q", got)
	}
}

func TestDetectContentType(t *testing.T) {
	// 已知扩展名返回对应的 MIME
	if got := detectContentType("a.txt"); got == "" {
		t.Error("a.txt 应有 MIME 类型")
	}
	// 未知扩展名回退 octet-stream
	if got := detectContentType("a.unknownext"); got != OctetStream {
		t.Errorf("未知扩展名应回退 octet-stream, got %q", got)
	}
}

func TestWrapHandler(t *testing.T) {
	r := NewRouter()

	if _, err := r.wrapHandler(HandlerFunc(func(Context) {})); err != nil {
		t.Errorf("HandlerFunc 不应报错: %v", err)
	}
	if _, err := r.wrapHandler(func(Context) {}); err != nil {
		t.Errorf("func(Context) 不应报错: %v", err)
	}
	if _, err := r.wrapHandler(func(http.ResponseWriter, *http.Request) {}); err != nil {
		t.Errorf("func(http.ResponseWriter, *http.Request) 不应报错: %v", err)
	}
	if _, err := r.wrapHandler(123); err == nil {
		t.Error("未知类型应返回错误")
	}
}

// 验证 HandlerFunc/MiddlewareFunc 类型别名可用
var _ HandlerFunc = MiddlewareFunc(func(Context) {})
