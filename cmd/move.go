package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [源文件或目录] [目标目录]",
	Short: "移动文件或目录到指定位置",
	Long: `
'move' 命令用于从源路径（文件或目录）并发地移动文件到目标路径。
它会充分利用 '--workers' 标志来指定的goroutine数量，以提高移动大量小文件时的效率。
`,

	// Args 用于验证命令行参数的数量
	Args: cobra.ExactArgs(2),
	// RunE 是命令的执行入口，'E' 表示它可以返回一个错误
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. 从 args 中获取源路径和目标路径
		sourcePath := args[0]
		destPath := args[1]

		// 2. 打印当前配置，验证命令是否正确接收了参数和标志
		fmt.Println("--- 开始执行移动操作 ---")
		fmt.Printf("源路径: %s\n", sourcePath)
		fmt.Printf("目标路径: %s\n", destPath)
		fmt.Printf("并发Workers数量: %d\n", workerCount)
		fmt.Printf("是否显示详细日志: %t\n", verbose)
		fmt.Printf("是否跳过错误继续: %t\n", skipErrors)
		fmt.Println("-----------------------")

		// 3. 这里是未来放置并发移动逻辑的地方
		fmt.Println("移动操作将在这实现。")

		// 4. 返回 nil 表示命令成功执行
		return nil
	},
}

// init 函数会在 main 函数执行前被自动调用
func init() {
	// 将 moveCmd 添加为 rootCmd 的子命令
	rootCmd.AddCommand(moveCmd)

	moveCmd.Flags().BoolP("overwrite", "o", false, "如果目标文件已存在，则覆盖它")
}
