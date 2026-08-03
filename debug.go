// Codes from https://github.com/gin-gonic/gin/blob/master/debug.go
package rock

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
)

type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

var DefaultWriter io.Writer = os.Stdout

// debugEnabled 显式调试开关，默认关闭。
// 生产环境保持关闭可避免路由表刷屏、以及向客户端暴露内部错误细节（见 WriteError）。
var debugEnabled = false

// SetDebug 显式开启/关闭调试输出。
// 测试环境（go test）下 IsDebugging 始终返回 true，不受此开关影响。
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

func IsDebugging() bool {
	// 测试环境始终开启，保证测试覆盖调试分支
	return debugEnabled || isInTestContext()
}

// isInTestContext 检测当前调用是否在 go test 环境中。
// 通过调用栈中是否存在 _test.go 文件判断，
// 不再依赖"函数名含 Test/Benchmark"这种脆弱启发式
// （生产二进制里函数名含 Test 也不会误判为调试环境）。
func isInTestContext() bool {
	// 检查环境变量（CI、测试标志等）
	if os.Getenv("CI") != "" || os.Getenv("TESTING") != "" {
		return true
	}

	// 检查调用栈中是否存在 _test.go 文件
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc) // 跳过当前函数
	if n == 0 {
		return false
	}

	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "_test.go") {
			return true
		}
		if !more {
			break
		}
	}

	return false
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// DebugPrintRouteFunc indicates debug log output format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())
		if DebugPrintRouteFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[ROCK-debug] "+format, values...)
	}
}
