// internal/fileops/move_test.go
package fileops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试 MoveFile（单个文件移动）
func TestMoveFile(t *testing.T) {
	srcDir, dstDir, cleanup := setupTest(t)
	defer cleanup()

	srcFile := filepath.Join(srcDir, "test1.txt")
	dstFile := filepath.Join(dstDir, "moved.txt")

	// 执行移动
	err := MoveFile(srcFile, dstFile, false)
	assert.NoError(t, err)

	// 验证源文件不存在
	_, err = os.Stat(srcFile)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// 验证目标文件存在且内容一致
	//srcContent, _ := os.ReadFile(filepath.Join(srcDir, "subdir", "test2.txt")) // 用另一个文件的内容对比
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), dstContent)
}
