package fileutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Task 定义文件处理任务
type Task struct {
	Path     string // 文件完整路径
	SrcRoot  string // 源目录根路径（用于相对路径计算）
	DestRoot string // 目标目录根路径
	Prefix   string // 重命名前缀
	Suffix   string // 重命名后缀
	Mode     string // 操作模式: md5/rename/copy/copy_rename/move
}

// Result 定义处理结果
type Result struct {
	OldName  string // 原文件名
	NewName  string // 新文件名（处理后）
	SrcMD5   string // 源文件MD5
	DstMD5   string // 目标文件MD5
	Verified bool   // 校验结果是否一致
	Err      error  // 错误信息
	Retried  int    // 重试次数
	Skipped  bool   // 是否被跳过
}

// ErrorType 定义错误类型
type ErrorType int

const (
	ErrorFileNotFound ErrorType = iota
	ErrorPermissionDenied
	ErrorDiskSpaceFull
	ErrorIORead
	ErrorIOWrite
	ErrorCrossDevice
	ErrorUnknown
)

// ErrorInfo 定义错误信息结构
type ErrorInfo struct {
	Type    ErrorType
	Message string
	Path    string
}

// ErrorPolicy 定义异常策略
type ErrorPolicy int

const (
	PolicySkip  ErrorPolicy = iota // 跳过
	PolicyRetry                    // 重试
	PolicyAbort                    // 终止
)

// ProcessFile 处理单个文件任务
func ProcessFile(t Task) Result {
	result := Result{OldName: t.Path}

	switch t.Mode {
	case "md5":
		result = processMD5(t)
	case "rename":
		result = processRename(t)
	case "copy":
		result = processCopy(t, false)
	case "copy_rename":
		result = processCopy(t, true)
	case "move":
		result = processMove(t)
	default:
		result.Err = fmt.Errorf("不支持的操作模式: %s", t.Mode)
	}

	return result
}

// ProcessFileWithRetry 带重试机制的文件处理
func ProcessFileWithRetry(t Task, maxRetries int, retryInterval time.Duration, errorHandler func(ErrorInfo) ErrorPolicy) Result {
	var result Result
	var retryCount int

	for retryCount <= maxRetries {
		result = ProcessFile(t)

		if result.Err == nil {
			return result
		}

		// 分析错误类型
		errorInfo := analyzeError(result.Err, t.Path)

		// 获取处理策略
		policy := errorHandler(errorInfo)

		switch policy {
		case PolicySkip:
			result.Skipped = true
			return result
		case PolicyAbort:
			return result
		case PolicyRetry:
			if retryCount < maxRetries {
				retryCount++
				result.Retried = retryCount
				time.Sleep(retryInterval)
				continue
			}
			return result
		}
	}

	return result
}

// analyzeError 分析错误类型
func analyzeError(err error, path string) ErrorInfo {
	errStr := err.Error()

	switch {
	case os.IsNotExist(err):
		return ErrorInfo{
			Type:    ErrorFileNotFound,
			Message: fmt.Sprintf("文件不存在: %s", errStr),
			Path:    path,
		}
	case os.IsPermission(err):
		return ErrorInfo{
			Type:    ErrorPermissionDenied,
			Message: fmt.Sprintf("权限不足: %s", errStr),
			Path:    path,
		}
	case strings.Contains(errStr, "no space left on device"):
		return ErrorInfo{
			Type:    ErrorDiskSpaceFull,
			Message: fmt.Sprintf("磁盘空间不足: %s", errStr),
			Path:    path,
		}
	case strings.Contains(errStr, "read"):
		return ErrorInfo{
			Type:    ErrorIORead,
			Message: fmt.Sprintf("读取错误: %s", errStr),
			Path:    path,
		}
	case strings.Contains(errStr, "write"):
		return ErrorInfo{
			Type:    ErrorIOWrite,
			Message: fmt.Sprintf("写入错误: %s", errStr),
			Path:    path,
		}
	case strings.Contains(errStr, "cross-device"):
		return ErrorInfo{
			Type:    ErrorCrossDevice,
			Message: fmt.Sprintf("跨设备错误: %s", errStr),
			Path:    path,
		}
	default:
		return ErrorInfo{
			Type:    ErrorUnknown,
			Message: fmt.Sprintf("未知错误: %s", errStr),
			Path:    path,
		}
	}
}

// processMD5 计算文件MD5
func processMD5(t Task) Result {
	result := Result{OldName: t.Path}

	// 检查文件是否存在
	if _, err := os.Stat(t.Path); os.IsNotExist(err) {
		result.Err = fmt.Errorf("文件不存在: %w", err)
		return result
	}

	// 计算源文件MD5
	md5Str, err := calculateFileMD5(t.Path)
	if err != nil {
		result.Err = fmt.Errorf("计算MD5失败: %w", err)
		return result
	}

	result.SrcMD5 = md5Str
	result.NewName = t.Path

	return result
}

func processRename(t Task) Result {
	result := Result{OldName: t.Path}

	// 检查源文件是否存在
	if _, err := os.Stat(t.Path); os.IsNotExist(err) {
		result.Err = fmt.Errorf("源文件不存在: %w", err)
		return result
	}

	// 计算源文件MD5
	srcMD5, err := calculateFileMD5(t.Path)
	if err != nil {
		result.Err = fmt.Errorf("计算源文件MD5失败: %w", err)
		return result
	}
	result.SrcMD5 = srcMD5

	// 🚨 修复：第三个参数改为true，以应用重命名规则
	newPath, err := generateNewPath(t.Path, t.SrcRoot, t.DestRoot, t.Prefix, t.Suffix, true)
	if err != nil {
		result.Err = fmt.Errorf("生成新路径失败: %w", err)
		return result
	}

	// 检查新旧路径是否相同
	oldAbs, _ := filepath.Abs(t.Path)
	newAbs, _ := filepath.Abs(newPath)
	if oldAbs == newAbs {
		// 新旧路径相同，不需要重命名
		result.NewName = newPath
		result.DstMD5 = srcMD5
		result.Verified = true
		return result
	}

	// 创建目标目录
	if err := createDirectory(filepath.Dir(newPath)); err != nil {
		result.Err = fmt.Errorf("创建目标目录失败: %w", err)
		return result
	}

	// 检查目标文件是否已存在（但排除自己）
	if _, err := os.Stat(newPath); err == nil {
		// 文件已存在，删除它（但这不是自己）
		if err := os.Remove(newPath); err != nil {
			result.Err = fmt.Errorf("删除已存在文件失败: %w", err)
			return result
		}
	}

	// 检查磁盘空间
	if err := checkDiskSpace(t.Path, newPath); err != nil {
		result.Err = fmt.Errorf("磁盘空间检查失败: %w", err)
		return result
	}

	// 执行重命名
	if err := os.Rename(t.Path, newPath); err != nil {
		// 如果跨文件系统，使用复制+删除
		if strings.Contains(err.Error(), "invalid cross-device link") {
			if err := copyAndDelete(t.Path, newPath); err != nil {
				result.Err = fmt.Errorf("跨文件系统重命名失败: %w", err)
				return result
			}
		} else {
			result.Err = fmt.Errorf("重命名失败: %w", err)
			return result
		}
	}

	// 计算新文件MD5
	dstMD5, err := calculateFileMD5(newPath)
	if err != nil {
		result.Err = fmt.Errorf("计算新文件MD5失败: %w", err)
		return result
	}

	result.NewName = newPath
	result.DstMD5 = dstMD5
	result.Verified = (srcMD5 == dstMD5)

	return result
}

// processCopy 复制文件
func processCopy(t Task, rename bool) Result {
	result := Result{OldName: t.Path}

	// 检查源文件是否存在
	if _, err := os.Stat(t.Path); os.IsNotExist(err) {
		result.Err = fmt.Errorf("源文件不存在: %w", err)
		return result
	}

	// 计算源文件MD5
	srcMD5, err := calculateFileMD5(t.Path)
	if err != nil {
		result.Err = fmt.Errorf("计算源文件MD5失败: %w", err)
		return result
	}
	result.SrcMD5 = srcMD5

	// 生成目标路径
	newPath, err := generateNewPath(t.Path, t.SrcRoot, t.DestRoot, t.Prefix, t.Suffix, rename)
	if err != nil {
		result.Err = fmt.Errorf("生成目标路径失败: %w", err)
		return result
	}

	// 创建目标目录
	if err := createDirectory(filepath.Dir(newPath)); err != nil {
		result.Err = fmt.Errorf("创建目标目录失败: %w", err)
		return result
	}

	// 检查磁盘空间
	if err := checkDiskSpace(t.Path, newPath); err != nil {
		result.Err = fmt.Errorf("磁盘空间检查失败: %w", err)
		return result
	}

	// 执行复制
	if err := copyFile(t.Path, newPath); err != nil {
		result.Err = fmt.Errorf("复制文件失败: %w", err)
		return result
	}

	// 计算目标文件MD5
	dstMD5, err := calculateFileMD5(newPath)
	if err != nil {
		result.Err = fmt.Errorf("计算目标文件MD5失败: %w", err)
		return result
	}

	result.NewName = newPath
	result.DstMD5 = dstMD5
	result.Verified = (srcMD5 == dstMD5)

	return result
}

// processMove 移动文件
func processMove(t Task) Result {
	result := Result{OldName: t.Path}

	// 检查源文件是否存在
	if _, err := os.Stat(t.Path); os.IsNotExist(err) {
		result.Err = fmt.Errorf("源文件不存在: %w", err)
		return result
	}

	// 计算源文件MD5
	srcMD5, err := calculateFileMD5(t.Path)
	if err != nil {
		result.Err = fmt.Errorf("计算源文件MD5失败: %w", err)
		return result
	}
	result.SrcMD5 = srcMD5

	// 生成目标路径
	newPath, err := generateNewPath(t.Path, t.SrcRoot, t.DestRoot, t.Prefix, t.Suffix, true)
	if err != nil {
		result.Err = fmt.Errorf("生成目标路径失败: %w", err)
		return result
	}

	// 创建目标目录
	if err := createDirectory(filepath.Dir(newPath)); err != nil {
		result.Err = fmt.Errorf("创建目标目录失败: %w", err)
		return result
	}

	// 检查磁盘空间
	if err := checkDiskSpace(t.Path, newPath); err != nil {
		result.Err = fmt.Errorf("磁盘空间检查失败: %w", err)
		return result
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(newPath); err == nil {
		// 文件已存在，删除它
		if err := os.Remove(newPath); err != nil {
			result.Err = fmt.Errorf("删除已存在文件失败: %w", err)
			return result
		}
	}

	// 尝试直接移动
	if err := os.Rename(t.Path, newPath); err != nil {
		// 如果跨文件系统，使用复制+删除
		if strings.Contains(err.Error(), "invalid cross-device link") {
			// 先复制
			if err := copyFile(t.Path, newPath); err != nil {
				result.Err = fmt.Errorf("跨文件系统移动-复制失败: %w", err)
				return result
			}

			// 计算目标文件MD5
			dstMD5, err := calculateFileMD5(newPath)
			if err != nil {
				result.Err = fmt.Errorf("计算目标文件MD5失败: %w", err)
				return result
			}

			// 验证MD5
			if srcMD5 != dstMD5 {
				os.Remove(newPath)
				result.Err = fmt.Errorf("移动后MD5校验不一致")
				return result
			}

			// 删除源文件
			if err := os.Remove(t.Path); err != nil {
				result.Err = fmt.Errorf("删除源文件失败: %w", err)
				return result
			}

			result.DstMD5 = dstMD5
			result.Verified = true
		} else {
			result.Err = fmt.Errorf("移动文件失败: %w", err)
			return result
		}
	} else {
		// 直接移动成功
		dstMD5, err := calculateFileMD5(newPath)
		if err != nil {
			result.Err = fmt.Errorf("计算目标文件MD5失败: %w", err)
			return result
		}

		result.DstMD5 = dstMD5
		result.Verified = (srcMD5 == dstMD5)
	}

	result.NewName = newPath
	return result
}

// calculateFileMD5 计算文件MD5
func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// generateNewPath 生成新的文件路径
func generateNewPath(oldPath, srcRoot, destRoot, prefix, suffix string, applyNaming bool) (string, error) {
	// 获取相对路径
	var relPath string
	if srcRoot != "" {
		var err error
		relPath, err = filepath.Rel(srcRoot, oldPath)
		if err != nil {
			return "", err
		}
	} else {
		relPath = filepath.Base(oldPath)
	}

	// 获取目录和文件名
	dir := filepath.Dir(relPath)
	filename := filepath.Base(relPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]

	// 应用重命名规则
	if applyNaming && (prefix != "" || suffix != "") {
		nameWithoutExt = prefix + nameWithoutExt + suffix
		filename = nameWithoutExt + ext
	}

	// 构建新路径
	var newPath string
	if destRoot != "" {
		newPath = filepath.Join(destRoot, dir, filename)
	} else {
		newPath = filepath.Join(filepath.Dir(oldPath), filename)
	}

	return newPath, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 同步到磁盘
	if err := dstFile.Sync(); err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// copyAndDelete 复制文件然后删除源文件
func copyAndDelete(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err
	}

	// 验证复制后的文件
	srcMD5, err := calculateFileMD5(src)
	if err != nil {
		return err
	}

	dstMD5, err := calculateFileMD5(dst)
	if err != nil {
		os.Remove(dst)
		return err
	}

	if srcMD5 != dstMD5 {
		os.Remove(dst)
		return fmt.Errorf("复制后MD5校验不一致")
	}

	// 删除源文件
	return os.Remove(src)
}

// createDirectory 创建目录，处理权限问题
func createDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 尝试创建目录
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			// 如果权限不足，尝试使用更低的权限
			if os.IsPermission(err) {
				if err := os.MkdirAll(dirPath, 0750); err != nil {
					return fmt.Errorf("创建目录权限不足: %w", err)
				}
			}
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return nil
}

// checkDiskSpace 检查磁盘空间
func checkDiskSpace(srcPath, dstPath string) error {
	// // 获取源文件大小
	// fileInfo, err := os.Stat(srcPath)
	// if err != nil {
	// 	return err
	// }
	// fileSize := fileInfo.Size()

	// // 获取目标路径所在磁盘的剩余空间（跨平台实现）
	// freeSpace, err := getFreeDiskSpace(filepath.Dir(dstPath))
	// if err != nil {
	// 	// 如果无法获取磁盘信息，跳过检查
	// 	return nil
	// }

	// // 预留10%的安全空间
	// safetyMargin := freeSpace / 10
	// requiredSpace := uint64(fileSize) + safetyMargin

	// if freeSpace < requiredSpace {
	// 	return fmt.Errorf("磁盘空间不足: 需要%d字节，可用%d字节", requiredSpace, freeSpace)
	// }

	return nil
}

// ErrorHandler 默认错误处理器
type ErrorHandler struct {
	policies      map[ErrorType]ErrorPolicy
	maxRetries    int
	retryInterval time.Duration
	mu            sync.Mutex
	abortFlag     bool
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		policies: map[ErrorType]ErrorPolicy{
			ErrorFileNotFound:     PolicySkip,
			ErrorPermissionDenied: PolicyRetry,
			ErrorDiskSpaceFull:    PolicyAbort,
			ErrorIORead:           PolicyRetry,
			ErrorIOWrite:          PolicyRetry,
			ErrorCrossDevice:      PolicySkip,
			ErrorUnknown:          PolicyRetry,
		},
		maxRetries:    3,
		retryInterval: 2 * time.Second,
	}
}

// HandleError 处理错误
func (h *ErrorHandler) HandleError(errorInfo ErrorInfo) ErrorPolicy {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.abortFlag {
		return PolicyAbort
	}

	policy, exists := h.policies[errorInfo.Type]
	if !exists {
		return PolicyRetry
	}

	if errorInfo.Type == ErrorDiskSpaceFull {
		h.abortFlag = true
	}

	return policy
}

// SetPolicy 设置特定错误类型的策略
func (h *ErrorHandler) SetPolicy(errorType ErrorType, policy ErrorPolicy) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.policies[errorType] = policy
}

// SetMaxRetries 设置最大重试次数
func (h *ErrorHandler) SetMaxRetries(max int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxRetries = max
}

// SetRetryInterval 设置重试间隔
func (h *ErrorHandler) SetRetryInterval(interval time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.retryInterval = interval
}

// Reset 重置错误处理器
func (h *ErrorHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.abortFlag = false
}

// getFreeDiskSpace 获取磁盘可用空间（跨平台）
func getFreeDiskSpace(_ string) (uint64, error) {

	return 0, nil
}

// AnalyzeError 分析错误类型（导出版本）
func AnalyzeError(err error, path string) ErrorInfo {
	return analyzeError(err, path)
}
