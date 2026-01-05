// internal/fileops/move.go

package fileops

import (
	"fmt"
	"os"
)

// MoveFile 将文件从源路径移动到目标路径。
// 它首先尝试重命名，如果跨设备则回退到“复制+删除”的方式。
func MoveFile(src, dst string, overwrite bool) error {
	// 1. 尝试直接重命名，这是最快的方式
	err := os.Rename(src, dst)
	if err == nil {
		return nil // 重命名成功
	}

	// 2. 如果重命名失败（通常是因为跨文件系统），则使用复制+删除的方式
	// 先创建目标目录
	if err := os.MkdirAll(Dir(dst), 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %w", err)
	}

	// 复制文件
	if err := CopyFile(src, dst, overwrite); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 复制成功后，删除源文件
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("移动后删除源文件失败: %w", err)
	}

	return nil
}
