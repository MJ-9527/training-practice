package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	srcCopyPath string // 源文件路径
	dstCopyPath string // 目标文件路径
	overwrite   bool   // 是否覆盖已存在的目标文件
)

// copyCmd 复制文件
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "复制文件到指定路径",
	Long:  `将源文件复制到目标路径，支持覆盖已存在的文件（需显式指定--overwrite）`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := copyFile(); err != nil {
			fmt.Fprintf(os.Stderr, "复制文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("文件已成功复制：%s -> %s\n", srcCopyPath, dstCopyPath)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// 添加参数
	copyCmd.Flags().StringVarP(&srcCopyPath, "source", "s", "", "源文件路径（必填）")
	copyCmd.Flags().StringVarP(&dstCopyPath, "dest", "d", "", "目标文件路径（必填）")
	copyCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "覆盖已存在的目标文件")
	_ = copyCmd.MarkFlagRequired("source")
	_ = copyCmd.MarkFlagRequired("dest")
}

// copyFile 核心复制逻辑
func copyFile() error {
	// 检查源文件是否存在
	srcStat, err := os.Stat(srcCopyPath)
	if err != nil {
		return fmt.Errorf("源文件不存在: %w", err)
	}
	if srcStat.IsDir() {
		return fmt.Errorf("源路径是目录，仅支持文件复制")
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(dstCopyPath); err == nil && !overwrite {
		return fmt.Errorf("目标文件已存在，使用--overwrite强制覆盖")
	}

	// 创建目标目录（如果不存在）
	dstDir := filepath.Dir(dstCopyPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 打开源文件
	srcFile, err := os.Open(srcCopyPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstCopyPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 同步文件到磁盘
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}

	// 保留源文件权限
	return os.Chmod(dstCopyPath, srcStat.Mode())
}
