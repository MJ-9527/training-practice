// cmd/copy.go

package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"training-practice/internal/fileops"
	"training-practice/internal/workerpool"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	overwrite bool
)

var copyCmd = &cobra.Command{
	Use:     "copy <source-path> <destination-path>",
	Aliases: []string{"cp"}, // 添加别名
	Short:   "并发地复制文件或目录",
	Long: `
'copy' 命令用于从源路径（可以是文件或目录）并发地复制内容到目标路径。
它会保留源目录的结构，并利用多核CPU来加速复制过程。`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destPath := args[1]

		// 1. 收集所有需要复制的源文件
		cmd.Println("正在扫描源文件...")
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("扫描源目录时出错: %w", err)
		}

		totalFiles := len(sourceFiles)
		if totalFiles == 0 {
			cmd.Println("没有找到任何文件需要复制。")
			return nil
		}

		// 2. 创建 Worker Pool
		pool := workerpool.NewWorkerPool(workerCount, totalFiles)

		// 3. 初始化进度条
		bar := progressbar.Default(int64(totalFiles), "正在复制...")

		// 4. 启动结果处理器 Goroutine
		var wg sync.WaitGroup
		var successCount, failCount int
		var failedFiles []string
		var mu sync.Mutex // 用于保护共享变量

		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range pool.ResultChan {
				// 每个结果处理完，进度条+1
				bar.Add(1)

				mu.Lock() // 锁定以安全地更新计数器和切片
				if res.Success {
					successCount++
				} else {
					failCount++
					failedFiles = append(failedFiles, res.Task.SourcePath)
					if verbose { // 如果开启了详细模式，则打印错误
						cmd.Printf("ERROR: %s: %v\n", res.Task.SourcePath, res.Error)
					}
				}
				mu.Unlock() // 解锁
			}
		}()

		// 5. 作为生产者，分发任务
		for _, srcFile := range sourceFiles {
			// 计算目标文件路径，以保持目录结构
			relPath, _ := filepath.Rel(sourcePath, srcFile)
			dstFile := filepath.Join(destPath, relPath)

			task := workerpool.Task{
				Type:       workerpool.TaskTypeCopy,
				SourcePath: srcFile,
				DestPath:   dstFile,
			}
			pool.TaskChan <- task
		}

		// 6. 关闭任务通道，等待所有任务完成
		pool.Close()

		// 7. 等待结果处理器 Goroutine 处理完所有结果
		wg.Wait()
		bar.Finish() // 完成进度条

		// 8. 打印总结报告
		cmd.Println("\n--- 复制完成 ---")
		cmd.Printf("总计: %d 个文件\n", totalFiles)
		cmd.Printf("成功: %d 个文件\n", successCount)
		cmd.Printf("失败: %d 个文件\n", failCount)

		if failCount > 0 && !verbose {
			cmd.Println("\n失败的文件列表:")
			for _, f := range failedFiles {
				cmd.Printf("  - %s\n", f)
			}
			cmd.Println("使用 -v 或 --verbose 标志查看详细错误信息。")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
