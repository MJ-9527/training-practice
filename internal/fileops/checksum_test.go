// internal/fileops/checksum_test.go
package fileops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试 CalculateChecksum
func TestCalculateChecksum(t *testing.T) {
	// 创建测试文件
	testFile := filepath.Join(t.TempDir(), "checksum.txt")
	content := []byte("test checksum")
	assert.NoError(t, os.WriteFile(testFile, content, 0644))

	// 预定义正确的哈希值（可通过终端命令计算：echo -n "test checksum" | md5sum）
	expectedMD5 := "d41d8cd98f00b204e9800998ecf8427e"                                    // 空字符串的 MD5（示例，需替换为实际值）
	expectedSHA256 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // 空字符串的 SHA256

	// 计算 MD5
	md5Sum, err := CalculateChecksum(testFile, MD5)
	assert.NoError(t, err)
	assert.Equal(t, expectedMD5, md5Sum)

	// 计算 SHA256
	sha256Sum, err := CalculateChecksum(testFile, SHA256)
	assert.NoError(t, err)
	assert.Equal(t, expectedSHA256, sha256Sum)

	// 测试不支持的算法
	_, err = CalculateChecksum(testFile, "invalid")
	assert.Error(t, err)
}
