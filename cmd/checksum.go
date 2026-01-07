// cmd/checksum.go

package cmd

import (
	"training-practice/internal/fileops"
	"training-practice/internal/workerpool"

	"github.com/spf13/cobra"
)

var (
	checksumAlgo string
)

var checksumCmd = &cobra.Command{
	Use:     "checksum <path>",
	Aliases: []string{"sum"}, // 添加别名
	Short:   "并发计算文件的校验和",
	Long: `
'checksum' 命令用于并发地计算一个或多个文件的哈希值。
支持的算法包括 md5, sha1, sha256, sha512。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourcePath := args[0]
		algo := fileops.Algorithm(checksumAlgo)

		// ... (此处逻辑与 copy.go 类似) ...
		// 1. 收集文件
		sourceFiles, err := fileops.CollectFiles(sourcePath)
		if err != nil {
			return err
		}

		// 2. 创建 WorkerPool
		pool := workerpool.NewWorkerPool(workerCount, len(sourceFiles))

		// 3. 启动结果处理器
		// 注意：对于 checksum，结果需要打印出来，格式类似 `hash  filepath`
		go func() {
			for res := range pool.ResultChan {
				if res.Success {
					cmd.Printf("%s  %s\n", res.Message, res.Task.SourcePath)
				} else if verbose {
					cmd.Printf("ERROR: %s: %v\n", res.Task.SourcePath, res.Error)
				}
			}
		}()

		// 4. 分发任务
		for _, srcFile := range sourceFiles {
			task := workerpool.Task{
				Type:       workerpool.TaskTypeChecksum,
				SourcePath: srcFile,
				Algorithm:  algo,
			}
			pool.TaskChan <- task
		}

		pool.Close()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checksumCmd)
	checksumCmd.Flags().StringVarP(&checksumAlgo, "algorithm", "a", "sha256", "指定哈希算法 (md5, sha1, sha256, sha512)")
}
