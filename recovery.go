package rock

import (
	"fmt"
	"runtime"
	"strings"
)

// trace 获取堆栈跟踪信息
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		str.WriteString(fmt.Sprintf("\n\t%s:%d", frame.File, frame.Line))
		if !more {
			break
		}
	}

	return str.String()
}

// Recovery 中间件：处理panic恢复
func Recovery() HandlerFunc {
	return func(c Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				// 记录错误日志
				fmt.Printf("%s\n\n", trace(message))
				
				// 使用统一的错误处理
				HandlePanic(c, message)
			}
		}()

		c.Next()
	}
}
