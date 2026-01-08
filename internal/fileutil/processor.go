package fileutil

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Task 定义文件处理任务
type Task struct {
	Path     string // 源文件的完整路径
	SrcRoot  string // 源目录（用于计算相对路径）
	DestRoot string // 目标根目录（为空表示就地操作）
	Prefix   string // 重命名前缀
	Suffix   string // 重命名后缀
	Mode     string // 操作模式: "md5", "rename", "copy", "copy_rename"
}

// Result 定义文件处理结果
type Result struct {
	OldName  string
	NewName  string
	SrcMD5   string // 源文件MD5
	DstMD5   string // 目标文件MD5（复制模式下）
	Verified bool   // MD5是否一致（复制校验）
	Err      error  // 错误信息
}

// ProcessFile 处理单个文件任务（修复文件句柄未释放问题）
func ProcessFile(t Task) Result {
	oldName := filepath.Base(t.Path)
	var (
		srcMD5   string
		dstMD5   string
		verified bool
		opErr    error
		newName  string
	)

	// 1. 计算源文件MD5（单独封装，确保文件句柄立即释放）
	srcMD5, opErr = calculateFileMD5(t.Path)
	if opErr != nil {
		return Result{
			OldName:  oldName,
			NewName:  newName,
			SrcMD5:   srcMD5,
			DstMD5:   dstMD5,
			Verified: verified,
			Err:      opErr,
		}
	}

	// 2. 处理不同模式（重命名前确保MD5的文件句柄已完全释放）
	switch t.Mode {
	case "rename":
		// 重命名逻辑
		ext := filepath.Ext(oldName)
		base := oldName[:len(oldName)-len(ext)]
		newName = fmt.Sprintf("%s%s%s%s", t.Prefix, base, t.Suffix, ext)
		newPath := filepath.Join(filepath.Dir(t.Path), newName)

		// Windows下增加重试机制（防止句柄延迟释放）
		opErr = renameWithRetry(t.Path, newPath, 3, 100*time.Millisecond)

	case "copy":
		// 复制逻辑（保留原文件名）
		relPath, _ := filepath.Rel(t.SrcRoot, t.Path)
		newPath := filepath.Join(t.DestRoot, relPath)
		newName = filepath.Base(newPath)

		// 创建目标目录
		if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
			opErr = err
			break
		}

		// 复制文件
		opErr = copyFile(t.Path, newPath)
		if opErr != nil {
			break
		}

		// 计算目标文件MD5
		dstMD5, opErr = calculateFileMD5(newPath)
		if opErr != nil {
			break
		}
		verified = (srcMD5 == dstMD5)

	case "copy_rename":
		// 复制+重命名逻辑
		ext := filepath.Ext(oldName)
		base := oldName[:len(oldName)-len(ext)]
		newName = fmt.Sprintf("%s%s%s%s", t.Prefix, base, t.Suffix, ext)

		relDir, _ := filepath.Rel(t.SrcRoot, filepath.Dir(t.Path))
		newPath := filepath.Join(t.DestRoot, relDir, newName)

		// 创建目标目录
		if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
			opErr = err
			break
		}

		// 复制文件
		opErr = copyFile(t.Path, newPath)
		if opErr != nil {
			break
		}

		// 计算目标文件MD5
		dstMD5, opErr = calculateFileMD5(newPath)
		if opErr != nil {
			break
		}
		verified = (srcMD5 == dstMD5)

	case "md5":
		// 仅计算MD5，无其他操作
		newName = oldName
	}

	return Result{
		OldName:  oldName,
		NewName:  newName,
		SrcMD5:   srcMD5,
		DstMD5:   dstMD5,
		Verified: verified,
		Err:      opErr,
	}
}

// calculateFileMD5 单独封装MD5计算，确保文件句柄立即释放
func calculateFileMD5(filePath string) (string, error) {
	// 打开文件（只读模式，减少锁占用）
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	// 立即关闭文件（defer会在函数结束时执行，这里确保句柄释放）
	defer func() {
		_ = f.Close()
	}()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 强制刷新并关闭文件（双重保障）
	_ = f.Close()
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// copyFile 封装文件复制，确保句柄释放
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 同步到磁盘并关闭
	_ = dstFile.Sync()
	_ = dstFile.Close()
	_ = srcFile.Close()

	return nil
}

// renameWithRetry Windows下重命名重试机制（解决句柄延迟释放）
func renameWithRetry(src, dst string, retryTimes int, delay time.Duration) error {
	var err error
	for i := 0; i < retryTimes; i++ {
		err = os.Rename(src, dst)
		if err == nil {
			return nil
		}
		// 重试前等待，让系统释放句柄
		time.Sleep(delay)
	}
	return fmt.Errorf("重命名失败（已重试%d次）: %w", retryTimes, err)
}
