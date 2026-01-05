// cmd/move.go

package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"training-practice/internal/fileops"
	"training-practice/internal/workerpool"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <source-path> <destination-path>",
	Short: "并发地移动文件或目录",
	Long: `
'move' 命令用于从源路径（可以是文件或目录）并发地移动内容到目标路径。
它会先尝试快速重命名，如果跨设备则采用“复制+删除”的方式。`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 此处的逻辑与 copy.go 几乎完全相同，
		// 只是任务类型和提示信息不同。
		sourcePath := args[0]
		destPath := args[1]
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		cmd.Println("正在扫描文件...")
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("扫描源目录时出错: %w", err)
		}

		if len(sourceFiles) == 0 {
			cmd.Println("没有找到任何文件需要移动。")
			return nil
		}

		pool := workerpool.NewWorkerPool(workerCount, len(sourceFiles))

		// 启动结果消费者 (与 copy.go 中的逻辑完全相同)
		// ... (此处省略，可直接从 copy.go 复制) ...
		// 注意：需要将进度条的前缀改为 "并发移动中: "
		// 3. 初始化进度条和结果统计
		bar := pb.StartNew(len(sourceFiles))
		bar.SetTemplate(pb.Full)
		bar.Set("prefix", "并发移动中: ")

		var successCount, failCount int
		var mu sync.Mutex // 用于保护对计数器和进度条的并发访问

		// 4. 启动一个goroutine来消费结果
		var resultWg sync.WaitGroup
		resultWg.Add(1)
		go func() {
			defer resultWg.Done()
			for result := range pool.ResultChan {
				mu.Lock() // 加锁，防止多个goroutine同时修改共享数据
				bar.Increment()
				if !result.Success {
					failCount++
					if verbose {
						cmd.Printf("错误: %s\n", result.Error)
					}
				} else {
					successCount++
				}
				mu.Unlock() // 解锁
			}
		}()

		// 作为生产者，分发 MOVE 任务
		for _, srcFile := range sourceFiles {
			relPath, _ := filepath.Rel(sourcePath, srcFile)
			dstFile := filepath.Join(destPath, relPath)

			task := workerpool.Task{
				Type:       workerpool.TaskTypeMove, // 指定任务类型为 Move
				SourcePath: srcFile,
				DestPath:   dstFile,
				Overwrite:  overwrite,
			}
			pool.TaskChan <- task
		}

		pool.Close()
		// ... (等待结果处理和打印总结的逻辑，与 copy.go 相同) ...
		// 注意：将总结信息中的“复制”改为“移动”

		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().BoolP("overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
