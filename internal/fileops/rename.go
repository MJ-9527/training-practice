// internal/fileops/rename.go

package fileops

import (
	"fmt"
	"os"
)

// RenameFile 重命名文件。
func RenameFile(oldPath, newPath string, overwrite bool) error {
	// 1. 检查目标文件是否存在，如果存在且不允许覆盖，则返回错误
	if !overwrite {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("目标文件 %s 已存在，且未设置覆盖标志", newPath)
		}
	}

	// 2. 直接调用 os.Rename
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("重命名文件失败 %s -> %s: %w", oldPath, newPath, err)
	}

	return nil
}
