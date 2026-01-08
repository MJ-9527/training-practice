package cmd

import (
	"os"

	"training-practice/internal/ui"

	"github.com/spf13/cobra"
)

// rootCmd 定义根命令
var rootCmd = &cobra.Command{
	Use:   "filetool",
	Short: "高并发文件处理工具",
	Long:  `支持文件MD5计算、重命名、复制、复制+重命名的高并发处理工具，带可视化界面`,
	Run: func(cmd *cobra.Command, args []string) {
		// 启动UI界面
		ui.Run()
	},
}

// Execute 执行根命令（供main.go调用）
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 可添加全局flag（如调试模式等），此处暂不添加
}
