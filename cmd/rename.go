// cmd/rename.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"training-practice/internal/fileops"
	"training-practice/internal/workerpool"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:     "rename <source-path> <new-name-or-directory>",
	Aliases: []string{"rn"}, // 添加别名
	Short:   "并发地重命名文件或移动到新目录",
	Long: `
'rename' 命令用于对源路径下的所有文件进行并发处理。
如果目标是一个已存在的目录，文件将被移动到该目录。
如果目标是一个新名称且源路径是单个文件，则该文件将被重命名。`,
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
			cmd.Println("没有找到任何文件需要处理。")
			return nil
		}

		// 检查目标路径是否为已存在的目录
		destInfo, err := os.Stat(destPath)
		isDestDir := err == nil && destInfo.IsDir()

		pool := workerpool.NewWorkerPool(workerCount, totalFiles)

		bar := progressbar.Default(int64(totalFiles), "正在处理...")

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
			var newPath string
			if isDestDir {
				// 如果目标是目录，则移动文件到该目录
				fileName := filepath.Base(srcFile)
				newPath = filepath.Join(destPath, fileName)
			} else if totalFiles == 1 {
				// 如果只有一个源文件且目标不是目录，则重命名该文件
				newPath = destPath
			} else {
				// 多个源文件但目标不是目录，这是一个错误
				return fmt.Errorf("无法将多个文件重命名为单个文件名 '%s'。请确保目标是一个目录。", destPath)
			}

			task := workerpool.Task{
				Type:       workerpool.TaskTypeRename,
				SourcePath: srcFile,
				DestPath:   newPath,
			}
			pool.TaskChan <- task
		}

		pool.Close()
		wg.Wait()
		bar.Finish()

		cmd.Println("\n--- 处理完成 ---")
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
	rootCmd.AddCommand(renameCmd)
}
