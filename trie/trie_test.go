package trie

import (
	"strings"
	"testing"
)

// --- 基础匹配 ---

func TestStaticMatch(t *testing.T) {
	tr := New()
	tr.Define("/a").Handle("GET", "h1")
	tr.Define("/a/b").Handle("GET", "h2")

	if m := tr.Match("/a"); m.Node == nil || m.Node.GetHandler("GET") != "h1" {
		t.Error("/a 应命中 h1")
	}
	if m := tr.Match("/a/b"); m.Node == nil || m.Node.GetHandler("GET") != "h2" {
		t.Error("/a/b 应命中 h2")
	}
	// 未匹配的路径
	if m := tr.Match("/a/c"); m.Node != nil {
		t.Error("/a/c 不应命中")
	}
}

func TestRootMatch(t *testing.T) {
	tr := New()
	tr.Define("/").Handle("GET", "root")
	if m := tr.Match("/"); m.Node == nil || m.Node.GetHandler("GET") != "root" {
		t.Error("/ 应命中 root")
	}
}

// --- 命名参数 ---

func TestNamedParam(t *testing.T) {
	tr := New()
	tr.Define("/users/:id").Handle("GET", "h")

	m := tr.Match("/users/42")
	if m.Node == nil {
		t.Fatal("应命中 /users/:id")
	}
	if m.Params["id"] != "42" {
		t.Errorf("id 参数应为 42, got %v", m.Params["id"])
	}
	// 参数缺失不应命中
	if m := tr.Match("/users"); m.Node != nil {
		t.Error("/users 不应命中 /users/:id")
	}
}

func TestMultipleNamedParams(t *testing.T) {
	tr := New()
	tr.Define("/users/:id/posts/:postId").Handle("GET", "h")

	m := tr.Match("/users/7/posts/99")
	if m.Node == nil {
		t.Fatal("应命中")
	}
	if m.Params["id"] != "7" || m.Params["postId"] != "99" {
		t.Errorf("params 错误: %v", m.Params)
	}
}

// --- 通配符 ---

func TestCatchAllParam(t *testing.T) {
	tr := New()
	tr.Define("/files/:path*").Handle("GET", "h")

	m := tr.Match("/files/a/b/c.txt")
	if m.Node == nil {
		t.Fatal("应命中 :path*")
	}
	if m.Params["path"] != "a/b/c.txt" {
		t.Errorf("path 通配符应捕获 a/b/c.txt, got %v", m.Params["path"])
	}
}

func TestStarWildcard(t *testing.T) {
	tr := New()
	tr.Define("/static/*").Handle("GET", "h")

	if m := tr.Match("/static/anything/here"); m.Node == nil {
		t.Error("/static/* 应命中 /static/anything/here")
	}
	// 空尾段：/static/ 与 /static 都不应命中 /static/*
	if m := tr.Match("/static/"); m.Node != nil {
		t.Error("/static/ 不应命中 /static/*（空尾段）")
	}
	if m := tr.Match("/static"); m.Node != nil {
		t.Error("/static 不应命中 /static/*（缺一段）")
	}
}

// --- 正则参数 ---

func TestRegexParam(t *testing.T) {
	tr := New()
	tr.Define("/users/:id(\\d+)").Handle("GET", "h")

	if m := tr.Match("/users/42"); m.Node == nil || m.Params["id"] != "42" {
		t.Error("数字 id 应命中")
	}
	if m := tr.Match("/users/abc"); m.Node != nil {
		t.Error("非数字 id 不应命中正则参数")
	}
}

func TestRegexParamOrdering(t *testing.T) {
	// 同一位置既有正则参数又有静态段，静态优先
	tr := New()
	tr.Define("/users/me").Handle("GET", "static")
	tr.Define("/users/:id(\\d+)").Handle("GET", "regex")

	if m := tr.Match("/users/me"); m.Node.GetHandler("GET") != "static" {
		t.Error("静态段应优先于正则参数")
	}
	if m := tr.Match("/users/42"); m.Node.GetHandler("GET") != "regex" {
		t.Error("数字应命中正则参数")
	}
}

// --- 字面量冒号 ---

func TestLiteralColon(t *testing.T) {
	tr := New()
	tr.Define("/x/::id").Handle("GET", "h")

	m := tr.Match("/x/:id")
	if m.Node == nil {
		t.Fatal("::id 应字面匹配 /x/:id")
	}
	if _, has := m.Params["id"]; has {
		t.Error("::id 不应捕获参数")
	}
}

// --- 后缀参数 ---

func TestSuffixParam(t *testing.T) {
	tr := New()
	tr.Define("/a/:file+.json").Handle("GET", "h")

	m := tr.Match("/a/data.json")
	if m.Node == nil {
		t.Fatal("应命中 :file+.json")
	}
	if m.Params["file"] != "data" {
		t.Errorf("file 参数应为 data, got %v", m.Params["file"])
	}
	// 不满足后缀的不应命中
	if m := tr.Match("/a/data.txt"); m.Node != nil {
		t.Error("data.txt 不应命中 :file+.json")
	}
}

// --- 大小写 ---

func TestIgnoreCaseDefault(t *testing.T) {
	tr := New() // 默认 IgnoreCase=true
	tr.Define("/users/:id").Handle("GET", "h")

	if m := tr.Match("/USERS/42"); m.Node == nil {
		t.Error("默认应忽略大小写")
	}
}

func TestCaseSensitive(t *testing.T) {
	tr := New(Options{}) // IgnoreCase=false
	tr.Define("/users/:id").Handle("GET", "h")

	if m := tr.Match("/USERS/42"); m.Node != nil {
		t.Error("Options{} 应区分大小写")
	}
	if m := tr.Match("/users/42"); m.Node == nil {
		t.Error("小写路径应命中")
	}
}

// --- 尾斜杠重定向 TSR ---

func TestTrailingSlashRemove(t *testing.T) {
	tr := New()
	tr.Define("/foo").Handle("GET", "h")

	m := tr.Match("/foo/")
	if m.Node != nil {
		t.Error("匹配到 /foo/ 时 Node 应为 nil（触发重定向）")
	}
	if m.TSR != "/foo" {
		t.Errorf("TSR 应为 /foo, got %q", m.TSR)
	}
}

func TestTrailingSlashAdd(t *testing.T) {
	tr := New()
	tr.Define("/foo/").Handle("GET", "h")

	m := tr.Match("/foo")
	if m.TSR != "/foo/" {
		t.Errorf("TSR 应为 /foo/, got %q", m.TSR)
	}
}

// --- 固定路径重定向 FPR ---

func TestFixedPathRedirect(t *testing.T) {
	tr := New()
	tr.Define("/a/b").Handle("GET", "h")

	m := tr.Match("/a//b")
	if m.Node != nil {
		t.Error("匹配 /a//b 时 Node 应为 nil")
	}
	if m.FPR != "/a/b" {
		t.Errorf("FPR 应为 /a/b, got %q", m.FPR)
	}
}

// --- 多方法 ---

func TestMultipleMethods(t *testing.T) {
	tr := New()
	node := tr.Define("/api")
	node.Handle("GET", "g")
	node.Handle("POST", "p")

	if node.GetHandler("GET") != "g" || node.GetHandler("POST") != "p" {
		t.Error("方法处理器存取错误")
	}
	if node.GetHandler("DELETE") != nil {
		t.Error("DELETE 未注册应为 nil")
	}
	allow := node.GetAllow()
	if !strings.Contains(allow, "GET") || !strings.Contains(allow, "POST") {
		t.Errorf("Allow 应包含 GET 和 POST, got %q", allow)
	}
	methods := node.GetMethods()
	if len(methods) != 2 {
		t.Errorf("应有 2 个方法, got %v", methods)
	}
}

func TestDuplicateMethodPanics(t *testing.T) {
	tr := New()
	node := tr.Define("/dup")
	node.Handle("GET", "g")
	assertPanics(t, func() { node.Handle("GET", "again") })
}

// --- 端点枚举 ---

func TestGetEndpoints(t *testing.T) {
	tr := New()
	tr.Define("/a")
	tr.Define("/a/b")
	tr.Define("/c")
	// 中间节点不应算端点
	if n := len(tr.GetEndpoints()); n != 3 {
		t.Errorf("应有 3 个端点, got %d", n)
	}
}

// --- 错误输入（应 panic） ---

func TestMatchPanicsWithoutSlash(t *testing.T) {
	tr := New()
	assertPanics(t, func() { tr.Match("a") })
}

func TestDefinePanicsOnMultiSlash(t *testing.T) {
	tr := New()
	assertPanics(t, func() { tr.Define("/a//b") })
}

func TestDefinePanicsPatternAfterWildcard(t *testing.T) {
	tr := New()
	assertPanics(t, func() { tr.Define("/a/*/b") })
}

func TestDefinePanicsInvalidParamName(t *testing.T) {
	tr := New()
	// 参数名必须是 [0-9A-Za-z_]，"." 非法
	assertPanics(t, func() { tr.Define("/a/:bad.name") })
}

func TestDefinePanicsEmptyColon(t *testing.T) {
	tr := New()
	// ":" 单独出现时参数名为空，不应 index-out-of-range，而是干净地 panic
	assertPanics(t, func() { tr.Define("/:") })
}

func TestDefinePanicsEmptySuffixName(t *testing.T) {
	tr := New()
	// ":+json"：后缀把参数名剥空，同样应干净 panic
	assertPanics(t, func() { tr.Define("/a/:+json") })
}

// --- 辅助 ---

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("应触发 panic，但没有")
		}
	}()
	f()
}
