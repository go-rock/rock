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

func IsDebugging() bool {
	// 在测试环境中开启调试输出，生产环境中关闭
	// 通过检查调用堆栈来判断是否在测试环境中
	return isInTestContext()
}

// isInTestContext 检测当前调用是否在测试环境中
func isInTestContext() bool {
	// 检查环境变量（CI、测试标志等）
	if os.Getenv("CI") != "" || os.Getenv("TESTING") != "" {
		return true
	}

	// 检查调用堆栈中是否包含测试函数
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc) // 跳过当前函数和IsDebugging函数
	if n == 0 {
		return false
	}

	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		funcName := frame.Function
		
		// 检查函数名是否包含测试相关标识
		if strings.Contains(funcName, "Test") || 
		   strings.Contains(funcName, "Benchmark") ||
		   strings.Contains(funcName, "testing.") {
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
