package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	srcMovePath string // 源文件路径
	dstMovePath string // 目标文件路径
	forceMove   bool   // 强制移动（覆盖目标）
)

// moveCmd 移动文件
var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "移动文件到指定路径",
	Long:  `将源文件移动到目标路径，支持跨目录移动，可强制覆盖已存在的目标文件`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := moveFile(); err != nil {
			fmt.Fprintf(os.Stderr, "移动文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("文件已成功移动：%s -> %s\n", srcMovePath, dstMovePath)
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)

	// 添加参数
	moveCmd.Flags().StringVarP(&srcMovePath, "source", "s", "", "源文件路径（必填）")
	moveCmd.Flags().StringVarP(&dstMovePath, "dest", "d", "", "目标文件路径（必填）")
	moveCmd.Flags().BoolVarP(&forceMove, "force", "f", false, "强制覆盖已存在的目标文件")
	_ = moveCmd.MarkFlagRequired("source")
	_ = moveCmd.MarkFlagRequired("dest")
}

// moveFile 核心移动逻辑
func moveFile() error {
	// 检查源文件是否存在
	srcStat, err := os.Stat(srcMovePath)
	if err != nil {
		return fmt.Errorf("源文件不存在: %w", err)
	}
	if srcStat.IsDir() {
		return fmt.Errorf("源路径是目录，仅支持文件移动")
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(dstMovePath); err == nil {
		if !forceMove {
			return fmt.Errorf("目标文件已存在，使用--force强制覆盖")
		}
		// 强制覆盖：先删除目标文件
		if err := os.Remove(dstMovePath); err != nil {
			return fmt.Errorf("删除现有目标文件失败: %w", err)
		}
	}

	// 创建目标目录（如果不存在）
	dstDir := filepath.Dir(dstMovePath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 执行移动（先尝试rename，跨文件系统则复制+删除）
	if err := os.Rename(srcMovePath, dstMovePath); err != nil {
		// rename失败，降级为复制+删除
		fmt.Println("跨文件系统移动，执行复制+删除...")
		if err := copyFileWithPath(srcMovePath, dstMovePath); err != nil {
			return fmt.Errorf("复制文件失败: %w", err)
		}
		if err := os.Remove(srcMovePath); err != nil {
			return fmt.Errorf("删除源文件失败: %w", err)
		}
	}

	return nil
}

// copyFileWithPath 复用复制逻辑（适配move的参数）
func copyFileWithPath(src, dst string) error {
	srcCopyPath = src
	dstCopyPath = dst
	overwrite = true
	return copyFile()
}
