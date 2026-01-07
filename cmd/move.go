// cmd/move.go

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

var moveCmd = &cobra.Command{
	Use:     "move <source-path> <destination-path>",
	Aliases: []string{"mv"}, // 添加别名
	Short:   "并发地移动文件或目录",
	Long: `
'move' 命令用于从源路径（可以是文件或目录）并发地移动内容到目标路径。
它会先尝试快速重命名，如果跨设备则采用“复制+删除”的方式。`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destPath := args[1]

		cmd.Println("正在扫描源文件...")
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("扫描源目录时出错: %w", err)
		}

		totalFiles := len(sourceFiles)
		if totalFiles == 0 {
			cmd.Println("没有找到任何文件需要移动。")
			return nil
		}

		pool := workerpool.NewWorkerPool(workerCount, totalFiles)

		bar := progressbar.Default(int64(totalFiles), "正在移动...")

		var wg sync.WaitGroup
		var successCount, failCount int
		var failedFiles []string
		var mu sync.Mutex

		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range pool.ResultChan {
				bar.Add(1)
				mu.Lock()
				if res.Success {
					successCount++
				} else {
					failCount++
					failedFiles = append(failedFiles, res.Task.SourcePath)
					if verbose {
						cmd.Printf("ERROR: %s: %v\n", res.Task.SourcePath, res.Error)
					}
				}
				mu.Unlock()
			}
		}()

		for _, srcFile := range sourceFiles {
			relPath, _ := filepath.Rel(sourcePath, srcFile)
			dstFile := filepath.Join(destPath, relPath)

			task := workerpool.Task{
				Type:       workerpool.TaskTypeMove,
				SourcePath: srcFile,
				DestPath:   dstFile,
			}
			pool.TaskChan <- task
		}

		pool.Close()
		wg.Wait()
		bar.Finish()

		cmd.Println("\n--- 移动完成 ---")
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
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
