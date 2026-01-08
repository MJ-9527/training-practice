package cmd

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash" // 新增：导入hash包（Hash类型属于这个包）
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	algorithm string // 校验和算法（md5/sha256）
	filePath  string // 目标文件路径
)

// checksumCmd 计算文件校验和
var checksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "计算文件的校验和（MD5/SHA256）",
	Long:  `指定文件路径和算法，计算并输出文件的MD5或SHA256校验和`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := calculateChecksum(); err != nil {
			fmt.Fprintf(os.Stderr, "计算校验和失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// 注册到根命令
	rootCmd.AddCommand(checksumCmd)

	// 添加命令行参数
	checksumCmd.Flags().StringVarP(&algorithm, "algorithm", "a", "md5", "校验和算法（可选：md5/sha256）")
	checksumCmd.Flags().StringVarP(&filePath, "file", "f", "", "目标文件路径（必填）")
	_ = checksumCmd.MarkFlagRequired("file") // 标记file为必填参数
}

// calculateChecksum 核心计算逻辑
func calculateChecksum() error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 选择算法（修复：将io.Hash改为hash.Hash）
	var hashFunc hash.Hash
	switch algorithm {
	case "md5":
		hashFunc = md5.New()
	case "sha256":
		hashFunc = sha256.New()
	default:
		return fmt.Errorf("不支持的算法: %s（仅支持md5/sha256）", algorithm)
	}

	// 读取文件并计算哈希
	if _, err := io.Copy(hashFunc, file); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 输出结果
	hashBytes := hashFunc.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)
	fmt.Printf("%s(%s) = %s\n", algorithm, filePath, hashStr)
	return nil
}
