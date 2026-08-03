package rock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupPatchAndOptions(t *testing.T) {
	app := New()
	app.Patch("/p", func(c Context) { c.String(200, "patch") })
	app.Options("/o", func(c Context) { c.String(200, "options") })

	server := httptest.NewServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPatch, server.URL+"/p", nil)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("PATCH failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("PATCH /p 应 200, got %d", resp.StatusCode)
	}

	req2, _ := http.NewRequest(http.MethodOptions, server.URL+"/o", nil)
	resp2, err := server.Client().Do(req2)
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Errorf("OPTIONS /o 应 200, got %d", resp2.StatusCode)
	}
}

func TestGroupUseFuncAndPriority(t *testing.T) {
	app := New()
	var order []string

	g := app.Group("/g")
	g.UseFunc(func(c Context) { order = append(order, "uf"); c.Next() })
	g.UseWithPriority(0, func(c Context) { order = append(order, "p0"); c.Next() })
	g.Get("/x", func(c Context) { order = append(order, "h"); c.String(200, "ok") })

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/g/x")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	resp.Body.Close()

	// p0 通过 UseWithPriority(0) 插到最前
	expected := []string{"p0", "uf", "h"}
	if len(order) != len(expected) {
		t.Fatalf("执行顺序应为 %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("执行顺序应为 %v, got %v", expected, order)
		}
	}
}

func TestGroupRemoveAndClearMiddleware(t *testing.T) {
	app := New()

	g := app.Group("/g")
	g.Use(func(c Context) { c.SetHeader("X-A", "1"); c.Next() })
	g.Use(func(c Context) { c.SetHeader("X-B", "1"); c.Next() })
	g.RemoveMiddleware(0) // 移除第一个
	g.Get("/x", func(c Context) { c.String(200, "ok") })

	cg := app.Group("/c")
	cg.Use(func(c Context) { c.SetHeader("X-C", "1"); c.Next() })
	cg.ClearMiddleware()
	cg.Get("/y", func(c Context) { c.String(200, "ok") })

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/g/x")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	resp.Body.Close()
	if resp.Header.Get("X-A") != "" {
		t.Error("RemoveMiddleware(0) 后 X-A 不应存在")
	}
	if resp.Header.Get("X-B") != "1" {
		t.Error("X-B 应保留")
	}

	resp2, err := server.Client().Get(server.URL + "/c/y")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.Header.Get("X-C") != "" {
		t.Error("ClearMiddleware 后 X-C 不应存在")
	}
}

func TestGroupSetRenderNoop(t *testing.T) {
	app := New()
	// SetRender 是空实现，只确保不 panic
	app.SetRender(nil)
}
