package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	oldName string // 原文件路径/名称
	newName string // 新文件路径/名称
	force   bool   // 强制重命名（覆盖已存在的新名称）
)

// renameCmd 重命名文件
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "重命名文件（或修改文件路径）",
	Long:  `修改文件的名称或路径，支持强制覆盖已存在的同名文件`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := renameFile(); err != nil {
			fmt.Fprintf(os.Stderr, "重命名文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("文件已成功重命名：%s -> %s\n", oldName, newName)
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)

	// 添加参数
	renameCmd.Flags().StringVarP(&oldName, "old", "o", "", "原文件路径/名称（必填）")
	renameCmd.Flags().StringVarP(&newName, "new", "n", "", "新文件路径/名称（必填）")
	renameCmd.Flags().BoolVarP(&force, "force", "f", false, "强制覆盖已存在的新文件")
	_ = renameCmd.MarkFlagRequired("old")
	_ = renameCmd.MarkFlagRequired("new")
}

// renameFile 核心重命名逻辑
func renameFile() error {
	// 检查原文件是否存在
	oldStat, err := os.Stat(oldName)
	if err != nil {
		return fmt.Errorf("原文件不存在: %w", err)
	}
	if oldStat.IsDir() {
		return fmt.Errorf("原路径是目录，仅支持文件重命名")
	}

	// 修复：要么使用newStat，要么直接判断错误（这里选择简化逻辑，去掉无用变量）
	if _, err := os.Stat(newName); err == nil {
		if !force {
			return fmt.Errorf("新文件已存在，使用--force强制覆盖")
		}
		// 强制覆盖：先删除新文件
		if err := os.Remove(newName); err != nil {
			return fmt.Errorf("删除现有新文件失败: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查新文件状态失败: %w", err)
	}

	// 创建新文件所在目录（如果不存在）
	newDir := filepath.Dir(newName)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("创建新文件目录失败: %w", err)
	}

	// 执行重命名
	if err := os.Rename(oldName, newName); err != nil {
		return fmt.Errorf("重命名失败: %w", err)
	}

	return nil
}
