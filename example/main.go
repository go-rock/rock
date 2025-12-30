package main

import (
	"fmt"
	"log"
	"time"

	"example/config"

	"github.com/go-rock/rock"
)

func main() {
	foo := "bar"
	fmt.Println(foo)
	app := rock.New()
	app.Use(rock.Recovery())
	app.Use(Logger())
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

	err := app.Run()
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
		"success":       true,
		"message":       fmt.Sprintf("成功上传 %d 个文件", len(files)),
		"uploaded_count": len(files),
		"files":         fileData,
	})
}
