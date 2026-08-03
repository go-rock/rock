package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"example/config"

	"github.com/go-rock/rock"
)

func main() {
	app := rock.New()
	app.Use(rock.Recovery())
	app.SetLogLevel(rock.LevelDebug)

	// app.Use(Logger())
	app.NoRoute(func(c rock.Context) {
		// c.JSON(404, rock.M{"msg": "404 not found"})
		c.Status(404)
		c.HTML("404")
	})

	// app.NoMethod(func(c rock.Context) {
	// 	// c.JSON(404, rock.M{"msg": "404 not found"})
	// 	c.Status(405)
	// 	c.String(405, "Method not allowed")
	// })

	config.Setup(app)
	app.Static("/assets", "./static")
	app.Static("/themes", "./themes/assets")
	// base.HTMLRender(render.Default())
	// app.RegisterView(render.New(render.ViewConfig{
	// 	ViewDir:   "./templates/pg2/",
	// 	Extension: ".html",
	// }))
	// app.RegisterView(rock.NewHtmlEngine("html engine"))
	app.Get("/", Home)

	// app.Get("/posts/:id", Post)
	admin := app.Group("/admin")
	admin.Use(auth())

	// admin.RegisterView(rock.NewPgEngine("pg engine"))
	// admin.RegisterView(render.New(render.ViewConfig{
	// 	ViewDir:   "./views/",
	// 	Extension: ".html",
	// }))

	admin.NoRoute(func(c rock.Context) {
		c.JSON(404, rock.M{"msg": "404 not found"})
	})

	// app.GetHTMLRender().SetViewDir("./tem/")
	// app.GetHTMLRender().SetViewDir("./template/")
	{
		// app.GetHTMLRender().SetViewDir("./tem/")
		admin.Get("/login", AdminLogin)
		admin.Post("/login", AdminLogin)
	}

	app.Get("/panic", func(c rock.Context) {
		names := []string{"geektutu"}
		c.String(200, names[100])
	})

	// 文件上传测试路由
	app.Get("/upload", UploadForm)
	app.Post("/upload", UploadHandler)

	// 多文件上传测试
	app.Get("/upload/multiple", MultipleUploadForm)
	app.Post("/upload/multiple", MultipleUploadHandler)

	// Store键值存储系统路由
	app.Get("/store/demo", StoreDemo)
	app.Get("/store/get/:key", StoreGetHandler)
	app.Post("/store/set", StoreSetHandler)
	app.Get("/store/immutable", StoreImmutableDemo)
	app.Get("/store/default", StoreDefaultDemo)
	app.Get("/store/concurrent", StoreConcurrentDemo)

	// 日志系统演示路由
	app.Get("/logger/demo", LoggerDemo)
	app.Get("/logger/levels", LoggerLevelsDemo)
	app.Get("/logger/request", LoggerRequestDemo)
	app.Get("/logger/context", LoggerContextDemo)
	app.Get("/logger/config", LoggerConfigDemo)

	// 简单测试各个日志级别的路由
	app.Get("/test/debug", func(c rock.Context) {
		rock.Debug("这是Debug级别日志 - 最详细的调试信息")
		rock.Debugf("格式化Debug日志: %s", "测试数据")
		c.JSON(200, rock.M{"message": "Debug日志已记录", "level": "DEBUG"})
	})

	app.Get("/test/info", func(c rock.Context) {
		rock.Info("这是Info级别日志 - 一般性信息记录")
		rock.Infof("格式化Info日志: %s", "测试数据")
		c.JSON(200, rock.M{"message": "Info日志已记录", "level": "INFO"})
	})

	app.Get("/test/warn", func(c rock.Context) {
		rock.Warn("这是Warn级别日志 - 警告信息")
		rock.Warnf("格式化Warn日志: %s", "测试数据")
		c.JSON(200, rock.M{"message": "Warn日志已记录", "level": "WARN"})
	})

	app.Get("/test/error", func(c rock.Context) {
		rock.Error("这是Error级别日志 - 错误信息")
		rock.Errorf("格式化Error日志: %s", "测试数据")
		c.JSON(200, rock.M{"message": "Error日志已记录", "level": "ERROR"})
	})

	// 监听端口读 config.json 的 "port"，缺省回退 :8989
	port := config.Config.GetString("port")
	if port == "" {
		port = ":8989"
	}
	err := app.Run(port)
	if err != nil {
		panic(err)
	}
}

func Home(c rock.Context) {
	// c.JSON(200, rock.H{"msg": "ok"})
	log.Println(c.GetView(), "HOME")
	c.HTML("home")
}

type Error struct {
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

// admin
func AdminLogin(c rock.Context) {
	log.Println("admin auth action")
	log.Println(c.GetView().Engine.Ext())

	error := &Error{Name: "render error", Msg: "Error msg"}
	// error := rock.M{"Msg": "xiao"}
	// c.JSON(http.StatusOK, rock.H{"msg": "admin login"})
	// c.Status(422)
	c.HTML("admin/login", rock.M{"data": error})
}

// Api
// func ApiIndex(c rock.Context) {
// 	c.JSON(200, rock.H{"msg": "api v1 index"})
// }

// middlewares
// func onlyForApi() rock.HandlerFunc {
// 	return func(c rock.Context) {
// 		// Start timer
// 		t := time.Now()
// 		// if a server error occurred
// 		c.Fail(500, "Internal Server Error")
// 		// Calculate resolution time
// 		log.Printf("Api only code [%d] %s in %v for group api", c.StatusCode(), c.Request().RequestURI, time.Since(t))
// 	}
// }

func auth() rock.HandlerFunc {
	return func(c rock.Context) {
		log.Println("auth before")
		// 模拟认证检查 - 实际应用中应该检查session、token等
		isAuthenticated := false // 这里是模拟，实际应该根据具体逻辑判断
		if !isAuthenticated {
			c.JSON(401, rock.H{"msg": "require admin"})
			c.Abort()
		}
		c.Next()
		log.Println("auth after")
	}
}

func Logger() rock.HandlerFunc {
	return func(c rock.Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s %s in %v", c.StatusCode(), c.Request().Method, c.Request().RequestURI, time.Since(t))
	}
}

// 文件上传处理函数

// UploadForm 显示单文件上传表单
func UploadForm(c rock.Context) {
	c.HTML("upload")
}

// UploadHandler 处理单文件上传
func UploadHandler(c rock.Context) {
	// 创建文件上传配置
	config := &rock.FileUploadConfig{
		MaxFileSize:        10 * 1024 * 1024, // 10MB
		AllowedExtensions:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt"},
		SaveDir:            "./uploads", // 保存到uploads目录
		GenerateUniqueName: true,
		FilenamePrefix:     "upload_",
	}

	// 上传文件
	fileInfo, err := c.SaveSingleFile("file", config)
	if err != nil {
		c.JSON(400, rock.H{
			"success": false,
			"message": fmt.Sprintf("上传失败: %v", err),
		})
		return
	}

	// 返回成功结果
	c.JSON(200, rock.H{
		"success": true,
		"message": "文件上传成功",
		"data": rock.H{
			"filename":    fileInfo.Filename,
			"size":        fileInfo.Size,
			"extension":   fileInfo.Extension,
			"saved_path":  fileInfo.SavedPath,
			"upload_time": fileInfo.UploadTime,
		},
	})
}

// MultipleUploadForm 显示多文件上传表单
func MultipleUploadForm(c rock.Context) {
	c.HTML("multiple_upload")
}

// MultipleUploadHandler 处理多文件上传
func MultipleUploadHandler(c rock.Context) {
	// 创建文件上传配置
	config := &rock.FileUploadConfig{
		MaxFileSize:        5 * 1024 * 1024, // 5MB per file
		AllowedExtensions:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
		SaveDir:            "./uploads/multiple", // 保存到uploads/multiple目录
		GenerateUniqueName: true,
		FilenamePrefix:     "multi_",
	}

	// 上传多个文件
	files, err := c.SaveMultipleFiles("files", config)
	if err != nil {
		c.JSON(400, rock.H{
			"success": false,
			"message": fmt.Sprintf("多文件上传失败: %v", err),
		})
		return
	}

	// 整理返回数据
	var fileData []rock.H
	for _, fileInfo := range files {
		fileData = append(fileData, rock.H{
			"filename":   fileInfo.Filename,
			"size":       fileInfo.Size,
			"extension":  fileInfo.Extension,
			"saved_path": fileInfo.SavedPath,
		})
	}

	// 返回成功结果
	c.JSON(200, rock.H{
		"success":        true,
		"message":        fmt.Sprintf("成功上传 %d 个文件", len(files)),
		"uploaded_count": len(files),
		"files":          fileData,
	})
}

// ==================== Store键值存储系统演示 ====================

// 全局Store实例
var store = rock.NewStore()

// StoreDemo 基本的Store使用演示
func StoreDemo(c rock.Context) {
	// 设置各种类型的值
	store.Set("string_value", "Hello Rock Framework!")
	store.Set("int_value", 42)
	store.Set("float_value", 3.14159)
	store.Set("bool_value", true)
	store.Set("slice_value", []string{"apple", "banana", "cherry"})
	store.Set("map_value", rock.M{
		"name":     "Rock",
		"version":  "1.0",
		"features": []string{"fast", "simple", "powerful"},
	})

	c.HTML("store")
}

// StoreGetHandler 根据key获取值
func StoreGetHandler(c rock.Context) {
	key := fmt.Sprintf("%v", c.Param("key"))
	value := store.Get(key)

	if value == nil {
		c.JSON(404, rock.H{
			"success": false,
			"message": fmt.Sprintf("Key '%s' not found", key),
			"key":     key,
		})
		return
	}

	c.JSON(200, rock.H{
		"success": true,
		"key":     key,
		"value":   value,
		"type":    fmt.Sprintf("%T", value),
	})
}

// StoreSetHandler 通过POST设置键值对
func StoreSetHandler(c rock.Context) {
	// 获取JSON数据
	var data struct {
		Key       string      `json:"key"`
		Value     interface{} `json:"value"`
		Immutable bool        `json:"immutable"`
	}

	// 使用ShouldBind方法进行JSON绑定
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(400, rock.H{
			"success": false,
			"message": fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	if data.Key == "" {
		c.JSON(400, rock.H{
			"success": false,
			"message": "Key cannot be empty",
		})
		return
	}

	// 保存到Store
	if data.Immutable {
		store.Save(data.Key, data.Value, true)
		c.JSON(200, rock.H{
			"success":   true,
			"message":   fmt.Sprintf("Immutable value saved for key '%s'", data.Key),
			"key":       data.Key,
			"value":     data.Value,
			"immutable": true,
		})
	} else {
		store.Set(data.Key, data.Value)
		c.JSON(200, rock.H{
			"success":   true,
			"message":   fmt.Sprintf("Value saved for key '%s'", data.Key),
			"key":       data.Key,
			"value":     data.Value,
			"immutable": false,
		})
	}
}

// StoreImmutableDemo 不可变数据演示
func StoreImmutableDemo(c rock.Context) {
	// 设置一个不可变的切片
	originalSlice := []string{"original", "data"}
	store.Save("immutable_slice", originalSlice, true)

	// 尝试修改原始切片
	originalSlice[0] = "modified"
	originalSlice = append(originalSlice, "new_item")

	// 获取Store中的值（应该是原始值）
	storedSlice := store.Get("immutable_slice")

	// 设置一个不可变的映射
	originalMap := rock.M{
		"key1": "value1",
		"key2": "value2",
	}
	store.Save("immutable_map", originalMap, true)

	// 修改原始映射
	originalMap["key1"] = "modified_value"
	originalMap["key3"] = "new_value"

	storedMap := store.Get("immutable_map")

	result := rock.M{
		"message":        "不可变数据演示",
		"original_slice": originalSlice,
		"stored_slice":   storedSlice,
		"original_map":   originalMap,
		"stored_map":     storedMap,
		"explanation":    "不可变数据在Store中保持原始值，不会受到外部修改影响",
	}

	c.JSON(200, result)
}

// 日志系统演示处理器

// LoggerDemo 基本日志功能演示
func LoggerDemo(c rock.Context) {
	// 记录各种级别的日志
	rock.Debug("这是调试日志 - LoggerDemo 入口")
	rock.Info("这是信息日志 - 正在演示基本日志功能")
	rock.Warn("这是警告日志 - 用户正在访问日志演示页面")
	rock.Error("这是错误日志 - 模拟一个错误情况")
	rock.Infof("致命级别 - 程序无法继续的错误（演示用Info级别）")

	// 记录格式化日志
	rock.Infof("当前时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	rock.Debugf("请求路径: %s", c.Request().URL.Path)
	rock.Warnf("用户代理: %s", c.Request().Header.Get("User-Agent"))

	// 使用不同日志级别的便捷函数
	rock.Info("这是一条信息日志")
	rock.Warn("这是一条警告日志")
	rock.Error("这是一条错误日志")

	// 记录到文件（如果配置了文件输出）
	app := c.App()
	if app != nil && app.Logger() != nil {
		app.Logger().Info("在日志演示中记录到文件")
	}

	// 切换到HTML视图展示日志功能
	c.HTML("logger")
}

// LoggerLevelsDemo 日志级别演示
func LoggerLevelsDemo(c rock.Context) {
	app := c.App()
	logger := app.Logger()

	// 获取当前日志级别
	currentLevel := logger.GetLevel()
	rock.Infof("当前日志级别: %s", currentLevel)

	// 演示不同级别的日志
	logger.Debug("调试级别 - 最详细的日志信息")
	logger.Info("信息级别 - 一般性信息记录")
	logger.Warn("警告级别 - 可能的问题警告")
	logger.Error("错误级别 - 错误情况记录")
	logger.Infof("致命级别 - 程序无法继续的错误（演示用Info级别）")

	// 演示格式化日志
	logger.Debugf("格式化调试: 数字 %d, 字符串 %s", 123, "test")
	logger.Infof("格式化信息: 浮点数 %.2f", 3.14159)
	logger.Warnf("格式化警告: 布尔值 %v", true)

	// 切换日志级别并演示
	rock.Info("=== 切换到Debug级别 ===")
	logger.SetLevel(rock.LevelDebug)
	logger.Debug("现在应该能看到这条调试日志")
	logger.Info("信息日志 - Debug级别下可见")

	rock.Info("=== 切换回Info级别 ===")
	logger.SetLevel(rock.LevelInfo)
	logger.Debug("现在不应该看到这条调试日志")
	logger.Info("信息日志 - Info级别下可见")

	rock.Warn("=== 切换到Warn级别 ===")
	logger.SetLevel(rock.LevelWarn)
	logger.Info("现在不应该看到这条信息日志")
	logger.Warn("警告日志 - Warn级别下可见")

	// 恢复到Info级别
	logger.SetLevel(rock.LevelInfo)

	c.JSON(200, rock.M{
		"success":       true,
		"message":       "日志级别演示完成",
		"current_level": currentLevel.String(),
		"explanation":   "日志级别从详细到严重：Debug < Info < Warn < Error < Fatal",
	})
}

// LoggerRequestDemo 请求日志演示
func LoggerRequestDemo(c rock.Context) {
	app := c.App()
	logger := app.Logger()

	// 启用请求日志
	logger.EnableRequestLog(true)
	rock.Info("启用请求日志记录")

	// 记录请求相关日志
	rock.Infof("处理请求: %s %s", c.Request().Method, c.Request().URL.Path)
	rock.Infof("客户端IP: %s", c.Request().RemoteAddr)
	rock.Infof("User-Agent: %s", c.Request().Header.Get("User-Agent"))

	// 模拟业务逻辑处理
	rock.Info("开始执行业务逻辑...")
	time.Sleep(100 * time.Millisecond)
	rock.Info("业务逻辑执行完成")

	// 记录请求处理时间
	startTime := time.Now()
	time.Sleep(50 * time.Millisecond)
	processingTime := time.Since(startTime)

	rock.Infof("请求处理耗时: %v", processingTime)

	// 模拟不同的响应状态
	statusCode := 200
	if c.Query("error") == "true" {
		statusCode = 500
		rock.Error("模拟服务器内部错误")
	} else if c.Query("warn") == "true" {
		statusCode = 400
		rock.Warn("模拟客户端错误")
	}

	// 注意：框架会自动记录请求日志，无需手动调用

	c.JSON(200, rock.M{
		"success":         true,
		"message":         "请求日志演示完成",
		"method":          c.Request().Method,
		"path":            c.Request().URL.Path,
		"status_code":     statusCode,
		"processing_time": processingTime.String(),
	})
}

// LoggerContextDemo Context日志演示
func LoggerContextDemo(c rock.Context) {
	// 使用Context的日志方法
	c.LogDebug("这是来自Context的调试日志")
	c.LogInfo("这是来自Context的信息日志")
	c.LogWarn("这是来自Context的警告日志")
	c.LogError("这是来自Context的错误日志")

	// 记录请求上下文信息
	c.LogInfo("请求路径: %s", c.Request().URL.Path)
	c.LogInfo("请求方法: %s", c.Request().Method)
	c.LogInfo("查询参数: %v", c.Request().URL.Query())

	// 记录处理过程
	c.LogDebug("开始处理请求...")
	c.LogDebug("验证参数...")
	c.LogDebug("执行业务逻辑...")
	c.LogDebug("准备响应...")
	c.LogInfo("请求处理完成")

	// 记录响应信息
	c.LogInfo("响应状态码: %d", c.StatusCode())
	c.LogInfo("响应内容类型: %s", c.Request().Header.Get("Content-Type"))

	// 模拟复杂处理流程
	steps := []string{
		"初始化",
		"参数验证",
		"权限检查",
		"数据处理",
		"结果验证",
		"响应准备",
	}

	for i, step := range steps {
		c.LogDebug("步骤 %d: %s", i+1, step)
		time.Sleep(10 * time.Millisecond)
	}

	c.LogInfo("Context日志演示完成 - 所有步骤已记录")

	c.JSON(200, rock.M{
		"success":     true,
		"message":     "Context日志演示完成",
		"steps_count": len(steps),
		"explanation": "Context日志方法可以直接在请求上下文中使用",
	})
}

// LoggerConfigDemo 日志配置演示
func LoggerConfigDemo(c rock.Context) {
	app := c.App()
	logger := app.Logger()

	// 获取当前配置
	currentLevel := logger.GetLevel()
	rock.Infof("当前日志级别: %s", currentLevel)

	// 演示输出重定向
	var buf bytes.Buffer
	logger.AddOutput(&buf)
	rock.Info("添加了新的日志输出目标")

	// 测试新的输出
	logger.Debug("调试日志 - 写入缓冲")
	logger.Info("信息日志 - 写入缓冲")
	logger.Warn("警告日志 - 写入缓冲")

	// 检查缓冲内容
	bufferContent := buf.String()
	rock.Infof("缓冲输出内容长度: %d 字符", len(bufferContent))

	// 移除缓冲输出
	logger.SetOutputs(os.Stdout) // 只输出到标准输出
	rock.Info("移除了缓冲输出，现在只输出到控制台")

	// 演示日志级别切换
	levels := []rock.LogLevel{
		rock.LevelDebug,
		rock.LevelInfo,
		rock.LevelWarn,
		rock.LevelError,
	}

	levelNames := []string{"Debug", "Info", "Warn", "Error"}

	for i, level := range levels {
		rock.Infof("设置日志级别为: %s", levelNames[i])
		logger.SetLevel(level)

		// 测试当前级别能看到的日志
		logger.Debug("调试日志 - 级别: " + levelNames[i])
		logger.Info("信息日志 - 级别: " + levelNames[i])
		logger.Warn("警告日志 - 级别: " + levelNames[i])
		logger.Error("错误日志 - 级别: " + levelNames[i])
	}

	// 恢复到原始级别
	logger.SetLevel(currentLevel)
	rock.Info("已恢复到原始日志级别")

	// 演示请求日志开关
	rock.Info("=== 演示请求日志开关 ===")
	logger.EnableRequestLog(true)
	rock.Info("请求日志已启用")

	logger.EnableRequestLog(false)
	rock.Info("请求日志已禁用")

	// 恢复到原始配置
	logger.EnableRequestLog(true)

	c.JSON(200, rock.M{
		"success":               true,
		"message":               "日志配置演示完成",
		"original_level":        currentLevel.String(),
		"buffer_content_length": len(bufferContent),
		"configured_levels":     levelNames,
		"explanation":           "演示了日志器的各种配置选项",
	})
}

// StoreDefaultDemo 默认值演示
func StoreDefaultDemo(c rock.Context) {
	// 尝试获取不存在的key，返回默认值
	nonExistent := store.GetDefault("non_existent_key", "default_value")
	nonExistentInt := store.GetDefault("non_existent_int", 100)
	nonExistentSlice := store.GetDefault("non_existent_slice", []string{"default", "values"})

	result := rock.M{
		"message": "默认值演示",
		"data": rock.M{
			"non_existent_string": nonExistent,
			"non_existent_int":    nonExistentInt,
			"non_existent_slice":  nonExistentSlice,
		},
		"explanation": "GetDefault方法在key不存在时返回指定的默认值",
	}

	c.JSON(200, result)
}

// StoreConcurrentDemo 并发访问演示
func StoreConcurrentDemo(c rock.Context) {
	done := make(chan bool, 10)

	// 启动多个goroutine同时访问Store
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 每个goroutine设置和读取不同的key
			key := fmt.Sprintf("concurrent_key_%d", id)
			value := fmt.Sprintf("value_from_goroutine_%d", id)

			store.Set(key, value)
			retrieved := store.Get(key)

			log.Printf("Goroutine %d: Set '%s' = '%s', Retrieved = '%s'",
				id, key, value, retrieved)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 展示并发访问结果
	concurrentData := rock.M{}
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("concurrent_key_%d", i)
		value := store.Get(key)
		if value != nil {
			concurrentData[key] = value
		}
	}

	result := rock.M{
		"message":     "并发访问演示完成",
		"data":        concurrentData,
		"total_keys":  len(concurrentData),
		"explanation": "Store支持并发读写，使用RWMutex保证线程安全",
	}

	c.JSON(200, result)
}
