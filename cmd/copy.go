// cmd/copy.go

package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"

	"training-practice/internal/fileops"
	"training-practice/internal/workerpool" // 导入workerpool包
)

var copyCmd = &cobra.Command{
	Use:   "copy <source-path> <destination-path>",
	Short: "并发地复制文件或目录",
	Long:  `...`, // 省略
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		destPath := args[1]
		//overwrite, _ := cmd.Flags().GetBool("overwrite")

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

		// 2. 初始化Worker Pool
		// 使用全局标志 --workers 来设置并发数
		// 缓冲区大小设置为文件总数，避免生产者阻塞
		pool := workerpool.NewWorkerPool(workerCount, len(sourceFiles))

		// 3. 初始化进度条和结果统计
		bar := pb.StartNew(len(sourceFiles))
		bar.SetTemplate(pb.Full)
		bar.Set("prefix", "并发复制中: ")

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

		// 5. 作为生产者，分发任务
		for _, srcFile := range sourceFiles {
			relPath, err := filepath.Rel(sourcePath, srcFile)
			if err != nil {
				if !skipErrors {
					return fmt.Errorf("无法计算相对路径 %s: %w", srcFile, err)
				}
				cmd.Printf("警告: 跳过无法处理的路径 %s\n", srcFile)
				continue
			}
			dstFile := filepath.Join(destPath, relPath)

			// 创建任务并发送到Worker Pool
			task := workerpool.Task{
				SourcePath: srcFile,
				DestPath:   dstFile,
			}
			pool.TaskChan <- task
		}

		// 6. 所有任务已分发，关闭Worker Pool
		pool.Close()

		// 7. 等待所有结果被处理完毕
		resultWg.Wait()

		// 8. 完成并打印总结
		bar.Finish()
		cmd.Println("--- 复制完成 ---")
		cmd.Printf("成功: %d 个文件\n", successCount)
		cmd.Printf("失败: %d 个文件\n", failCount)

		if failCount > 0 && !skipErrors {
			return fmt.Errorf("复制过程中发生 %d 个错误", failCount)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolP("overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
