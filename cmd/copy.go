// cmd/copy.go

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cheggaaa/pb/v3" // 导入进度条库
	"github.com/spf13/cobra"

	// 导入我们自己的 fileops 包
	"training-practice/internal/fileops"
)

var copyCmd = &cobra.Command{
	Use:   "copy <source-path> <destination-path>",
	Short: "并发地复制文件或目录",
	Long:  `...`, // 省略
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destPath := args[1]

		// 从本地标志中获取 overwrite 的值
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		// 1. 收集所有待复制的文件
		cmd.Println("正在扫描文件...")
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("扫描源目录时出错: %w", err)
		}

		if len(sourceFiles) == 0 {
			cmd.Println("没有找到任何文件需要复制。")
			return nil
		}

		// 2. 初始化进度条
		bar := pb.StartNew(len(sourceFiles))
		bar.SetTemplate(pb.Full) // 使用完整的进度条模板
		bar.Set("prefix", "复制中: ")

		// 3. 遍历文件列表，逐个复制
		for _, srcFile := range sourceFiles {
			// 计算目标文件的相对路径，以保持目录结构
			relPath, err := filepath.Rel(sourcePath, srcFile)
			if err != nil {
				bar.Increment()
				if skipErrors {
					cmd.Printf("警告: 无法计算相对路径 %s -> %s, 跳过。\n", srcFile, destPath)
					continue
				}
				return fmt.Errorf("无法计算相对路径 %s -> %s: %w", srcFile, destPath, err)
			}

			// 构建目标文件的完整路径
			dstFile := filepath.Join(destPath, relPath)

			// 调用核心复制函数
			err = fileops.CopyFile(srcFile, dstFile, overwrite)
			if err != nil {
				bar.Increment()
				if skipErrors {
					cmd.Printf("警告: %v, 跳过。\n", err)
					continue
				}
				return err // 如果不跳过错误，则直接返回并终止程序
			}

			// 更新进度条
			bar.Increment()
		}

		// 4. 完成并结束进度条
		bar.Finish()
		cmd.Println("复制完成！")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolP("overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
