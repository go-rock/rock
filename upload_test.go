package rock

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileUpload(t *testing.T) {
	app := New()

	app.Post("/upload", func(c Context) {
		config := &FileUploadConfig{
			MaxFileSize:        10 * 1024 * 1024, // 10MB
			AllowedExtensions:  []string{".txt", ".jpg", ".png"},
			SaveDir:            "./test_uploads",
			GenerateUniqueName: true,
		}

		fileInfo, err := c.SaveSingleFile("file", config)
		if err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}

		c.JSON(200, M{
			"success":  true,
			"filename": fileInfo.Filename,
			"size":     fileInfo.Size,
			"path":     fileInfo.SavedPath,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建测试文件内容
	fileContent := []byte("test file content")
	body := &bytes.Buffer{}

	// 创建multipart表单
	writer := multipart.NewWriter(body)
	writer.WriteField("description", "test file")
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write(fileContent)
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 清理测试文件
	os.RemoveAll("./test_uploads")
}

func TestFileUploadMultiple(t *testing.T) {
	app := New()

	app.Post("/upload/multiple", func(c Context) {
		config := &FileUploadConfig{
			MaxFileSize:        5 * 1024 * 1024, // 5MB
			AllowedExtensions:  []string{".txt", ".jpg", ".png"},
			SaveDir:            "./test_uploads",
			GenerateUniqueName: true,
		}

		files, err := c.SaveMultipleFiles("files", config)
		if err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}

		c.JSON(200, M{
			"success": true,
			"count":   len(files),
			"files":   files,
		})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建多个文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("description", "multiple files test")

	part1, _ := writer.CreateFormFile("files", "file1.txt")
	part1.Write([]byte("content 1"))

	part2, _ := writer.CreateFormFile("files", "file2.txt")
	part2.Write([]byte("content 2"))

	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/upload/multiple", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload multiple files: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 清理测试文件
	os.RemoveAll("./test_uploads")
}

func TestFileUploadValidation(t *testing.T) {
	app := New()

	app.Post("/upload/validate", func(c Context) {
		config := &FileUploadConfig{
			MaxFileSize:       1024, // 1KB - 很小的限制
			AllowedExtensions: []string{".txt"},
			SaveDir:           "./test_uploads",
		}

		_, err := c.SaveSingleFile("file", config)
		if err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}

		c.JSON(200, M{"success": true})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 创建大文件（超过1KB限制）
	_largeContent := bytes.Repeat([]byte("a"), 2*1024) // 2KB
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write(_largeContent)

	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/upload/validate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload large file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for oversized file, got %d", resp.StatusCode)
	}

	// 清理测试文件
	os.RemoveAll("./test_uploads")
}

func TestFileUploadExtension(t *testing.T) {
	app := New()

	app.Post("/upload/ext", func(c Context) {
		config := &FileUploadConfig{
			MaxFileSize:       10 * 1024 * 1024,
			AllowedExtensions: []string{".txt", ".log"},
			SaveDir:           "./test_uploads",
		}

		_, err := c.SaveSingleFile("file", config)
		if err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}

		c.JSON(200, M{"success": true})
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试不允许的扩展名
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "test.exe")
	part.Write([]byte("executable"))

	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/upload/ext", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload file with forbidden extension: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for forbidden extension, got %d", resp.StatusCode)
	}

	// 清理测试文件
	os.RemoveAll("./test_uploads")
}

func TestStaticFile(t *testing.T) {
	// 创建测试静态文件
	testDir := "./test_static"
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFile := filepath.Join(testDir, "test.txt")
	os.WriteFile(testFile, []byte("Hello, World!"), 0644)

	app := New()
	app.Static("/static", testDir)

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试获取静态文件
	resp, err := client.Get(server.URL + "/static/test.txt")
	if err != nil {
		t.Fatalf("Failed to GET static file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(body))
	}

	// 测试获取不存在的文件
	resp, err = client.Get(server.URL + "/static/nonexistent.txt")
	if err != nil {
		t.Fatalf("Failed to GET nonexistent file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestStaticFileDirectory(t *testing.T) {
	// 创建测试目录结构
	testDir := "./test_static_dir"
	os.MkdirAll(filepath.Join(testDir, "subdir"), 0755)
	defer os.RemoveAll(testDir)

	// 创建文件
	os.WriteFile(filepath.Join(testDir, "root.txt"), []byte("root file"), 0644)
	os.WriteFile(filepath.Join(testDir, "subdir", "sub.txt"), []byte("sub file"), 0644)

	app := New()
	app.Static("/files", testDir)

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试根文件
	resp, err := client.Get(server.URL + "/files/root.txt")
	if err != nil {
		t.Fatalf("Failed to GET root file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for root file, got %d", resp.StatusCode)
	}

	// 测试子目录文件
	resp, err = client.Get(server.URL + "/files/subdir/sub.txt")
	if err != nil {
		t.Fatalf("Failed to GET subdir file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for subdir file, got %d", resp.StatusCode)
	}

	// 测试目录浏览（应该返回404，因为没有实现目录浏览）
	resp, err = client.Get(server.URL + "/files/")
	if err != nil {
		t.Fatalf("Failed to GET directory: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404 for directory browsing, got %d", resp.StatusCode)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	app := New()
	app.Use(Recovery())

	app.Get("/panic", func(c Context) {
		panic("test panic")
	})

	app.Get("/normal", func(c Context) {
		c.String(200, "normal response")
	})

	server := httptest.NewServer(app)
	defer server.Close()

	client := server.Client()

	// 测试正常请求
	resp, err := client.Get(server.URL + "/normal")
	if err != nil {
		t.Fatalf("Failed to GET /normal: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for normal request, got %d", resp.StatusCode)
	}

	// 测试panic恢复
	resp, err = client.Get(server.URL + "/panic")
	if err != nil {
		t.Fatalf("Failed to GET /panic: %v", err)
	}
	defer resp.Body.Close()

	// panic应该被恢复，返回500错误
	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500 for panic recovery, got %d", resp.StatusCode)
	}
}

func BenchmarkFileUpload(b *testing.B) {
	// 创建测试文件
	testFile := "./test_benchmark.txt"
	fileContent := bytes.Repeat([]byte("test"), 1024) // 4KB文件
	os.WriteFile(testFile, fileContent, 0644)
	defer os.Remove(testFile)

	app := New()

	uploadHandler := func(c Context) {
		config := &FileUploadConfig{
			MaxFileSize:       10 * 1024 * 1024,
			AllowedExtensions: []string{".txt"},
			SaveDir:           "./benchmark_uploads",
		}

		_, err := c.SaveSingleFile("file", config)
		if err != nil {
			c.JSON(400, M{"error": err.Error()})
			return
		}

		c.JSON(200, M{"success": true})
	}

	app.Post("/upload", uploadHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建模拟multipart请求避免网络端口冲突
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write(fileContent)

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
	}

	// 清理
	os.RemoveAll("./benchmark_uploads")
}
