// cmd/rename.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"training-practice/internal/fileops"
	"training-practice/internal/workerpool"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <source-path> <new-name-or-pattern>",
	Short: "并发地重命名文件或目录",
	Long: `
'rename' 命令用于对源路径下的所有文件进行并发重命名。
<new-name-or-pattern> 可以是一个新的文件名（用于单个文件）或一个简单的模式。
例如，要给所有 .txt 文件加上 .bak 后缀:
training-practice rename ./docs/*.txt '{{.Name}}.bak'
注意：模式功能在本阶段暂未实现，当前版本只支持简单的批量重命名。`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		newNameOrPattern := args[1]
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		cmd.Println("正在扫描文件...")
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return fmt.Errorf("扫描源目录时出错: %w", err)
		}

		if len(sourceFiles) == 0 {
			cmd.Println("没有找到任何文件需要重命名。")
			return nil
		}

		// 简化版逻辑：假设 newNameOrPattern 是一个新的目录或文件名
		// 这是一个简化的实现，更复杂的模式匹配可以在未来添加
		isDestDir := true // 假设目标是一个目录
		if len(sourceFiles) == 1 {
			// 如果只有一个文件，可以重命名为一个具体的文件名
			if !IsDir(newNameOrPattern) { // 需要一个辅助函数来判断路径是否为目录
				isDestDir = false
			}
		}

		pool := workerpool.NewWorkerPool(workerCount, len(sourceFiles))

		// 启动结果消费者 (与 copy.go 中的逻辑完全相同)
		// ... (此处省略，可直接从 copy.go 复制) ...
		// 注意：需要将进度条的前缀改为 "并发重命名中: "
		// 3. 初始化进度条和结果统计
		bar := pb.StartNew(len(sourceFiles))
		bar.SetTemplate(pb.Full)
		bar.Set("prefix", "并发重命名中: ")

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

		// 作为生产者，分发 RENAME 任务
		for _, srcFile := range sourceFiles {
			var dstFile string
			if isDestDir {
				fileName := filepath.Base(srcFile)
				dstFile = filepath.Join(newNameOrPattern, fileName)
			} else {
				// 对于单个文件，直接使用新名称
				dstFile = newNameOrPattern
			}

			task := workerpool.Task{
				Type:       workerpool.TaskTypeRename, // 指定任务类型为 Rename
				SourcePath: srcFile,
				DestPath:   dstFile,
				Overwrite:  overwrite,
			}
			pool.TaskChan <- task
		}

		pool.Close()
		// ... (等待结果处理和打印总结的逻辑，与 copy.go 相同) ...
		// 注意：将总结信息中的“复制”改为“重命名”

		return nil
	},
}

// 辅助函数，用于判断路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.Flags().BoolP("overwrite", "o", false, "如果目标文件已存在，则强制覆盖")
}
