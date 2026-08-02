package rock

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileUploadConfig 文件上传配置
type FileUploadConfig struct {
	// 文件大小限制 (字节)
	MaxFileSize int64

	// 允许的文件类型
	AllowedExtensions []string

	// 允许的MIME类型
	AllowedMimeTypes []string

	// 保存目录
	SaveDir string

	// 是否生成唯一文件名
	GenerateUniqueName bool

	// 文件名前缀
	FilenamePrefix string
}

// FileInfo 文件信息
type FileInfo struct {
	Header     *multipart.FileHeader `json:"header"`
	Filename   string                `json:"filename"`
	Extension  string                `json:"extension"`
	Size       int64                 `json:"size"`
	MIMEType   string                `json:"mime_type"`
	SavedPath  string                `json:"saved_path,omitempty"`
	URL        string                `json:"url,omitempty"`
	UploadTime time.Time             `json:"upload_time"`
}

// DefaultFileUploadConfig 获取默认文件上传配置
func DefaultFileUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		MaxFileSize:        10 * 1024 * 1024, // 10MB
		AllowedExtensions:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".txt"},
		AllowedMimeTypes:   []string{"image/jpeg", "image/png", "image/gif", "application/pdf", "text/plain"},
		SaveDir:            "uploads",
		GenerateUniqueName: true,
	}
}

// ValidateFile 验证文件
func ValidateFile(fh *multipart.FileHeader, config *FileUploadConfig) error {
	if fh == nil {
		return NewAppError(ErrBadRequest, "File header is nil")
	}

	// 检查文件大小
	if config.MaxFileSize > 0 && fh.Size > config.MaxFileSize {
		return NewAppErrorWithDetail(ErrBadRequest,
			fmt.Sprintf("File size exceeds limit: %d bytes", config.MaxFileSize),
			fmt.Sprintf("File size: %d bytes", fh.Size))
	}

	// 验证文件扩展名
	if len(config.AllowedExtensions) > 0 {
		ext := getFileExtension(fh.Filename)
		allowed := false
		for _, allowedExt := range config.AllowedExtensions {
			if strings.ToLower(allowedExt) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return NewAppErrorWithDetail(ErrBadRequest,
				fmt.Sprintf("File extension '%s' not allowed", ext),
				fmt.Sprintf("Allowed extensions: %v", config.AllowedExtensions))
		}
	}

	return nil
}

// GetFileMIMEType 获取文件MIME类型
func GetFileMIMEType(fh *multipart.FileHeader) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件头的前512字节来检测MIME类型
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, nil
}

// getFileExtension 获取文件扩展名（小写）
func getFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// shouldSaveToDisk 检查是否应该保存文件到磁盘
func shouldSaveToDisk(config *FileUploadConfig) bool {
	return config.SaveDir != ""
}

// ValidateMIMEType 验证MIME类型
func ValidateMIMEType(fh *multipart.FileHeader, config *FileUploadConfig) error {
	if len(config.AllowedMimeTypes) == 0 {
		return nil // 如果没有设置允许的MIME类型，跳过验证
	}

	mimeType, err := GetFileMIMEType(fh)
	if err != nil {
		return NewError(ErrBadRequest, "Failed to detect MIME type: %v", err)
	}

	allowed := false
	for _, allowedMime := range config.AllowedMimeTypes {
		if mimeType == allowedMime {
			allowed = true
			break
		}
	}

	if !allowed {
		return NewAppErrorWithDetail(ErrBadRequest,
			fmt.Sprintf("MIME type not allowed: %s", mimeType),
			fmt.Sprintf("Allowed MIME types: %v", config.AllowedMimeTypes))
	}

	return nil
}

// SaveSingleFile 保存单个文件
func SaveSingleFile(c Context, name string, config *FileUploadConfig) (*FileInfo, error) {
	fh, err := c.FormFile(name)
	if err != nil {
		return nil, NewError(ErrBadRequest, "Failed to get file: %v", err)
	}

	// 验证文件
	if err := ValidateFile(fh, config); err != nil {
		return nil, err
	}

	// 验证MIME类型
	if err := ValidateMIMEType(fh, config); err != nil {
		return nil, err
	}

	// 创建文件信息
	fileInfo := &FileInfo{
		Header:     fh,
		Filename:   fh.Filename,
		Extension:  getFileExtension(fh.Filename),
		Size:       fh.Size,
		UploadTime: time.Now(),
	}

	// 如果配置了保存目录，保存文件
	if shouldSaveToDisk(config) {
		savedPath, err := saveFileToDisk(fh, config)
		if err != nil {
			return nil, err
		}
		fileInfo.SavedPath = savedPath
	}

	return fileInfo, nil
}

// SaveMultipleFiles 保存多个文件
func SaveMultipleFiles(c Context, name string, config *FileUploadConfig) ([]*FileInfo, error) {
	// 解析multipart表单
	if err := c.ParseMultipartForm(config.MaxFileSize); err != nil {
		return nil, NewError(ErrBadRequest, "Failed to parse multipart form: %v", err)
	}

	// 获取文件头
	form := c.Request().MultipartForm
	if form == nil {
		return nil, NewAppError(ErrBadRequest, "No multipart form data found")
	}

	fhs, ok := form.File[name]
	if !ok || len(fhs) == 0 {
		return nil, NewAppError(ErrBadRequest, fmt.Sprintf("No files found with name: %s", name))
	}

	var fileInfos []*FileInfo
	for _, fh := range fhs {
		// 验证文件
		if err := ValidateFile(fh, config); err != nil {
			return nil, err
		}

		// 验证MIME类型
		if err := ValidateMIMEType(fh, config); err != nil {
			return nil, err
		}

		// 创建文件信息
		fileInfo := &FileInfo{
			Header:     fh,
			Filename:   fh.Filename,
			Extension:  getFileExtension(fh.Filename),
			Size:       fh.Size,
			UploadTime: time.Now(),
		}

		// 如果配置了保存目录，保存文件
		if shouldSaveToDisk(config) {
			savedPath, err := saveFileToDisk(fh, config)
			if err != nil {
				return nil, err
			}
			fileInfo.SavedPath = savedPath
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

// saveFileToDisk 保存文件到磁盘
func saveFileToDisk(fh *multipart.FileHeader, config *FileUploadConfig) (string, error) {
	// 创建保存目录
	if err := os.MkdirAll(config.SaveDir, 0755); err != nil {
		return "", NewError(ErrInternalServer, "Failed to create directory: %v", err)
	}

	// 生成文件名
	var filename string
	if config.GenerateUniqueName {
		ext := getFileExtension(fh.Filename)
		filename = fmt.Sprintf("%s%d%s",
			config.FilenamePrefix,
			time.Now().UnixNano(),
			ext)
	} else {
		filename = fh.Filename
	}

	// 确保文件名安全
	filename = filepath.Base(filename)
	if filename == "." || filename == "/" {
		return "", NewAppError(ErrBadRequest, "Invalid filename")
	}

	// 构建完整路径
	fullPath := filepath.Join(config.SaveDir, filename)

	// 保存文件
	src, err := fh.Open()
	if err != nil {
		return "", NewError(ErrInternalServer, "Failed to open file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", NewError(ErrInternalServer, "Failed to create file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", NewError(ErrInternalServer, "Failed to save file: %v", err)
	}

	return fullPath, nil
}

// GetUploadHandler 获取文件上传处理器
func GetUploadHandler(config *FileUploadConfig) HandlerFunc {
	if config == nil {
		config = DefaultFileUploadConfig()
	}

	return func(c Context) {
		// 设置最大内存用于multipart解析
		if config.MaxFileSize > 0 {
			c.Request().ParseMultipartForm(config.MaxFileSize)
		}

		c.Next()
	}
}

// Convenience methods for common upload scenarios

// UploadSingleImage 上传单张图片
func UploadSingleImage(c Context, name string) (*FileInfo, error) {
	config := DefaultFileUploadConfig()
	config.AllowedExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}
	config.AllowedMimeTypes = []string{"image/jpeg", "image/png", "image/gif"}
	config.SaveDir = "uploads/images"

	return SaveSingleFile(c, name, config)
}

// UploadSingleDocument 上传单个文档
func UploadSingleDocument(c Context, name string) (*FileInfo, error) {
	config := DefaultFileUploadConfig()
	config.AllowedExtensions = []string{".pdf", ".doc", ".docx", ".txt"}
	config.AllowedMimeTypes = []string{"application/pdf", "application/msword", "text/plain"}
	config.SaveDir = "uploads/documents"

	return SaveSingleFile(c, name, config)
}

// UploadMultipleImages 上传多张图片
func UploadMultipleImages(c Context, name string) ([]*FileInfo, error) {
	config := DefaultFileUploadConfig()
	config.AllowedExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}
	config.AllowedMimeTypes = []string{"image/jpeg", "image/png", "image/gif"}
	config.SaveDir = "uploads/images"

	return SaveMultipleFiles(c, name, config)
}
