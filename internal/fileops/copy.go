// internal/fileops/copy.go

package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile 将单个文件从源路径复制到目标路径。
// overwrite: 如果目标文件已存在，是否强制覆盖。
func CopyFile(src, dst string, overwrite bool) error {
	// 1. 检查源文件是否存在
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法访问源文件 %s: %w", src, err)
	}

	// 2. 检查目标文件是否存在，如果存在且不允许覆盖，则返回错误
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("目标文件 %s 已存在，且未设置覆盖标志 (-o/--overwrite)", dst)
		}
	}

	// 3. 创建目标文件所在的目录（如果不存在）
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %w", dstDir, err)
	}

	// 4. 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件 %s: %w", src, err)
	}
	defer srcFile.Close() // 确保函数退出时源文件被关闭

	// 5. 创建目标文件
	// 使用 os.Create 会截断（清空）已存在的文件，这符合我们的覆盖逻辑。
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件 %s: %w", dst, err)
	}
	defer dstFile.Close() // 确保函数退出时目标文件被关闭

	// 6. 执行复制操作
	// io.Copy 是一个高效的复制函数，它会自动处理缓冲区，适合各种大小的文件。
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容时出错 %s -> %s: %w", src, dst, err)
	}

	// 7. 确保数据完全写入磁盘
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("同步文件到磁盘时出错 %s: %w", dst, err)
	}

	// 8. 复制源文件的权限
	if err := os.Chmod(dst, srcFileInfo.Mode()); err != nil {
		return fmt.Errorf("复制文件权限时出错 %s: %w", dst, err)
	}

	return nil // 复制成功
}
