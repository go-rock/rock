package rock

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteSuccess(t *testing.T) {
	app := New()
	app.Get("/ok", func(c Context) {
		WriteSuccess(c, M{"a": 1})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/ok")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("WriteSuccess 应返回 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"success":true`) || !strings.Contains(string(body), `"a":1`) {
		t.Errorf("响应应含 success:true 和 data, got %s", body)
	}
}

func TestValidationErrorString(t *testing.T) {
	e := ValidationError{Field: "name", Message: "required"}
	s := e.Error()
	if !strings.Contains(s, "name") || !strings.Contains(s, "required") {
		t.Errorf("Error() 应包含字段和消息, got %q", s)
	}
}

func TestGetErrorMessage(t *testing.T) {
	if got := GetErrorMessage(ErrNotFound); got != "Not Found" {
		t.Errorf("ErrNotFound 应映射为 Not Found, got %q", got)
	}
	if got := GetErrorMessage(ErrorCode(999)); got != "Unknown Error" {
		t.Errorf("未知错误码应返回 Unknown Error, got %q", got)
	}
}

func TestAppErrorFormat(t *testing.T) {
	e := NewAppError(ErrBadRequest, "bad")
	if !strings.Contains(e.Error(), "400") || !strings.Contains(e.Error(), "bad") {
		t.Errorf("AppError.Error() 应含 code 和 message, got %q", e.Error())
	}
	e2 := NewAppErrorWithDetail(ErrBadRequest, "bad", "detail")
	if !strings.Contains(e2.Error(), "detail") {
		t.Errorf("带 Detail 的 Error() 应含 detail, got %q", e2.Error())
	}
}
