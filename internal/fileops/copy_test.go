// internal/fileops/copy_test.go
package fileops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试环境初始化：创建临时文件和目录
func setupTest(t *testing.T) (srcDir, dstDir string, cleanup func()) {
	// 创建临时源目录
	srcDir = t.TempDir()
	// 创建临时测试文件
	testFile1 := filepath.Join(srcDir, "test1.txt")
	assert.NoError(t, os.WriteFile(testFile1, []byte("hello world"), 0644))
	testFile2 := filepath.Join(srcDir, "subdir", "test2.txt")
	assert.NoError(t, os.MkdirAll(filepath.Dir(testFile2), 0755))
	assert.NoError(t, os.WriteFile(testFile2, []byte("filebatch test"), 0644))

	// 创建临时目标目录
	dstDir = t.TempDir()

	// 清理函数（测试结束后执行）
	cleanup = func() {
		os.RemoveAll(srcDir)
		os.RemoveAll(dstDir)
	}

	return srcDir, dstDir, cleanup
}

// 测试 CopyFile（单个文件复制）
func TestCopyFile(t *testing.T) {
	_, dstDir, cleanup := setupTest(t)
	defer cleanup()

	srcFile := filepath.Join(t.TempDir(), "source.txt")
	dstFile := filepath.Join(dstDir, "dest.txt")
	content := []byte("test copy")

	// 准备源文件
	assert.NoError(t, os.WriteFile(srcFile, content, 0644))

	// 执行复制
	err := CopyFile(srcFile, dstFile, false)
	assert.NoError(t, err)

	// 验证目标文件存在且内容一致
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

// 测试 CopyFile 覆盖模式
func TestCopyFile_Overwrite(t *testing.T) {
	_, dstDir, cleanup := setupTest(t)
	defer cleanup()

	srcFile := filepath.Join(t.TempDir(), "source.txt")
	dstFile := filepath.Join(dstDir, "dest.txt")
	oldContent := []byte("old content")
	newContent := []byte("new content")

	// 先创建已存在的目标文件
	assert.NoError(t, os.WriteFile(dstFile, oldContent, 0644))

	// 不开启覆盖模式，复制应失败
	assert.Error(t, CopyFile(srcFile, dstFile, false))

	// 开启覆盖模式，复制应成功
	assert.NoError(t, os.WriteFile(srcFile, newContent, 0644))
	assert.NoError(t, CopyFile(srcFile, dstFile, true))

	// 验证内容被覆盖
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, newContent, dstContent)
}
