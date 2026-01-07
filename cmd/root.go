// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	verbose     bool
	workerCount int // 从配置文件读取默认值
)

// rootCmd 是基础命令
var rootCmd = &cobra.Command{
	Use:   "filebatch",
	Short: "高并发文件批量处理工具",
	Long: `filebatch 是一个支持并发复制、移动、重命名和校验和计算的命令行工具，
旨在提高文件操作效率，支持跨平台使用。`,
	// 初始化配置（在子命令执行前）
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		// 优先级：命令行参数 > 配置文件 > 环境变量 > 默认值
		viper.BindPFlag("workers", cmd.Flags().Lookup("workers"))
		viper.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))
		// 读取配置到全局变量
		workerCount = viper.GetInt("workers")
		verbose = viper.GetBool("verbose")
	},
}

// 初始化配置
func initConfig() {
	// 1. 配置文件路径优先级：命令行指定 > 默认路径
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// 默认配置文件路径：~/.filebatch.yaml
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)         // 搜索 ~ 目录
		viper.SetConfigName(".filebatch") // 配置文件名（无后缀）
		viper.SetConfigType("yaml")       // 配置文件类型
	}

	// 2. 支持环境变量（前缀：FILEBATCH_，如 FILEBATCH_WORKERS=8）
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FILEBATCH")

	// 3. 设置默认值
	viper.SetDefault("workers", 4)                   // 默认并发数：4
	viper.SetDefault("verbose", false)               // 默认不开启详细模式
	viper.SetDefault("checksum.algorithm", "sha256") // 默认校验和算法

	// 4. 读取配置文件（如果存在）
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Printf("使用配置文件：%s\n", viper.ConfigFileUsed())
		}
	}
}

// 初始化全局标志
func init() {
	// 配置文件相关标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "指定配置文件路径（默认：~/.filebatch.yaml）")
	// 通用标志（所有子命令可继承）
	rootCmd.PersistentFlags().IntVarP(&workerCount, "workers", "w", 0, "并发工作线程数（默认：4，可通过配置文件修改）")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "开启详细日志输出")
}

// Execute 启动命令行工具
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
