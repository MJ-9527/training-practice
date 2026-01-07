// internal/fileops/checksum.go

package fileops

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
)

// Algorithm 定义了支持的哈希算法
type Algorithm string

const (
	MD5    Algorithm = "md5"
	SHA1   Algorithm = "sha1"
	SHA256 Algorithm = "sha256"
	SHA512 Algorithm = "sha512"
)

// CalculateChecksum 计算单个文件的哈希值
func CalculateChecksum(filePath string, algo Algorithm) (string, error) {
	var h hash.Hash
	switch algo {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return "", fmt.Errorf("不支持的哈希算法: %s", algo)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件 %s: %w", filePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("读取文件 %s 时出错: %w", filePath, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CalculateChecksums 批量计算多个文件的哈希值
func CalculateChecksums(filePaths []string, algo Algorithm) (map[string]string, error) {
	results := make(map[string]string)
	for _, path := range filePaths {
		checksum, err := CalculateChecksum(path, algo)
		if err != nil {
			return nil, err
		}
		results[path] = checksum
	}
	return results, nil
}
