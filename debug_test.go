package rock

import "testing"

func TestSetDebug(t *testing.T) {
	SetDebug(false)
	if debugEnabled {
		t.Error("SetDebug(false) 应关闭 debugEnabled")
	}
	SetDebug(true)
	if !debugEnabled {
		t.Error("SetDebug(true) 应开启 debugEnabled")
	}
	// 测试环境下 IsDebugging 恒为 true
	if !IsDebugging() {
		t.Error("test env 下 IsDebugging 应为 true")
	}
	SetDebug(false) // 还原，避免影响其他测试
}
