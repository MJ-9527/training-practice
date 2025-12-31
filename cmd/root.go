package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	workerCount int
	verbose     bool
	skipErrors  bool
)

var rootCmd = &cobra.Command{
	Use:   "training-practice",
	Short: "一个高性能的并发文件批量处理工具",
	Long: `FileBatch 是一个用Go编写的命令行工具，旨在利用goroutine和channel的强大能力，
对大量文件进行快速、高效的批量操作，如复制、移动、重命名和校验。`,
}

// Execute 将所有子命令添加到根命令并设置标志。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	rootCmd.PersistentFlags().IntVarP(&workerCount, "workers", "w", 10, "设置并发工作的goroutine数量")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细的执行过程")
	rootCmd.PersistentFlags().BoolVarP(&skipErrors, "skip-errors", "s", false, "遇到错误时跳过并继续处理")
}
