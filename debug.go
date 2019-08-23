package rock

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strings"
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return true
}

var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

func (app *App) debugPrintRoute(httpMethod, absolutePath string, handlers []Handler) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		h := LastHandler(handlers)
		handlerName := nameOfFunction(h)
		if DebugPrintRouteFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

var DefaultWriter io.Writer = os.Stdout

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[ROCK-debug] "+format, values...)
	}
}
