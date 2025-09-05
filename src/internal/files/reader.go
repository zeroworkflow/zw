package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFileSize = 1024 * 1024 // 1MB limit per file
	maxTotalSize = 5 * 1024 * 1024 // 5MB total limit
)

// FileContent represents a file with its content
type FileContent struct {
	Path    string
	Content string
	Size    int64
}

// Reader handles reading and processing files
type Reader struct {
	maxFileSize  int64
	maxTotalSize int64
}

// NewReader creates a new file reader
func NewReader() *Reader {
	return &Reader{
		maxFileSize:  maxFileSize,
		maxTotalSize: maxTotalSize,
	}
}

// ReadFiles reads multiple files and returns their content
func (r *Reader) ReadFiles(filePaths []string) ([]FileContent, error) {
	var files []FileContent
	var totalSize int64

	for _, path := range filePaths {
		// Clean and validate path
		cleanPath := filepath.Clean(path)
		
		// Check if file exists
		info, err := os.Stat(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("file %s: %w", cleanPath, err)
		}

		// Check if it's a directory
		if info.IsDir() {
			return nil, fmt.Errorf("path %s is a directory, not a file", cleanPath)
		}

		// Check file size
		if info.Size() > r.maxFileSize {
			return nil, fmt.Errorf("file %s is too large (%.2f MB), max allowed: %.2f MB", 
				cleanPath, float64(info.Size())/1024/1024, float64(r.maxFileSize)/1024/1024)
		}

		// Check total size limit
		if totalSize + info.Size() > r.maxTotalSize {
			return nil, fmt.Errorf("total files size exceeds limit (%.2f MB)", 
				float64(r.maxTotalSize)/1024/1024)
		}

		// Read file content
		content, err := r.readFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", cleanPath, err)
		}

		files = append(files, FileContent{
			Path:    cleanPath,
			Content: content,
			Size:    info.Size(),
		})

		totalSize += info.Size()
	}

	return files, nil
}

// readFile reads a single file and returns its content
func (r *Reader) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read content with size limit
	limitedReader := io.LimitReader(file, r.maxFileSize)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	// Check if file is binary
	if r.isBinary(content) {
		return fmt.Sprintf("[Binary file: %s, size: %d bytes]", path, len(content)), nil
	}

	return string(content), nil
}

// isBinary checks if content appears to be binary
func (r *Reader) isBinary(content []byte) bool {
	// Check for null bytes in first 512 bytes
	checkSize := len(content)
	if checkSize > 512 {
		checkSize = 512
	}

	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}

// FormatFilesForAI formats file contents for AI consumption
func (r *Reader) FormatFilesForAI(files []FileContent) string {
	if len(files) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\n\n--- ФАЙЛЫ ДЛЯ АНАЛИЗА ---\n\n")

	for i, file := range files {
		if i > 0 {
			builder.WriteString("\n" + strings.Repeat("-", 50) + "\n\n")
		}

		builder.WriteString(fmt.Sprintf("📁 **Файл: %s**\n", file.Path))
		builder.WriteString(fmt.Sprintf("📊 Размер: %.2f KB\n\n", float64(file.Size)/1024))
		
		// Detect file type for syntax highlighting
		ext := strings.ToLower(filepath.Ext(file.Path))
		lang := r.detectLanguage(ext)
		
		if strings.HasPrefix(file.Content, "[Binary file:") {
			builder.WriteString(file.Content)
		} else {
			builder.WriteString(fmt.Sprintf("```%s\n%s\n```", lang, file.Content))
		}
		
		builder.WriteString("\n")
	}

	builder.WriteString("\n--- КОНЕЦ ФАЙЛОВ ---\n")
	return builder.String()
}

// detectLanguage detects programming language by file extension
func (r *Reader) detectLanguage(ext string) string {
	langMap := map[string]string{
		".go":     "go",
		".py":     "python",
		".js":     "javascript",
		".ts":     "typescript",
		".java":   "java",
		".cpp":    "cpp",
		".c":      "c",
		".cs":     "csharp",
		".php":    "php",
		".rb":     "ruby",
		".rs":     "rust",
		".swift":  "swift",
		".kt":     "kotlin",
		".scala":  "scala",
		".sh":     "bash",
		".bash":   "bash",
		".zsh":    "bash",
		".fish":   "fish",
		".ps1":    "powershell",
		".bat":    "batch",
		".cmd":    "batch",
		".html":   "html",
		".htm":    "html",
		".css":    "css",
		".scss":   "scss",
		".sass":   "sass",
		".less":   "less",
		".xml":    "xml",
		".json":   "json",
		".yaml":   "yaml",
		".yml":    "yaml",
		".toml":   "toml",
		".ini":    "ini",
		".cfg":    "ini",
		".conf":   "ini",
		".sql":    "sql",
		".md":     "markdown",
		".txt":    "text",
		".log":    "text",
		".env":    "bash",
		".gitignore": "text",
		".dockerfile": "dockerfile",
		".makefile":  "makefile",
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return "text"
}

// ValidateFiles validates file paths before reading
func (r *Reader) ValidateFiles(filePaths []string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no files specified")
	}

	if len(filePaths) > 10 {
		return fmt.Errorf("too many files specified (max: 10)")
	}

	for _, path := range filePaths {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("empty file path")
		}

		// Security check - prevent path traversal
		cleanPath := filepath.Clean(path)
		if strings.Contains(cleanPath, "..") {
			return fmt.Errorf("path traversal not allowed: %s", path)
		}
	}

	return nil
}
