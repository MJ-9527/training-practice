// internal/fileops/utils.go

package fileops

import (
	"os"
	"path/filepath"
)

// CollectFiles 递归遍历指定目录，并返回所有文件的绝对路径列表。
func CollectFiles(rootDir string) ([]string, error) {
	var fileList []string

	// filepath.WalkDir 是一个方便的目录遍历函数
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // 如果访问某个路径时出错，停止遍历
		}
		// 如果不是目录（即它是一个文件），则将其路径添加到列表中
		if !d.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func Dir(path string) string {
	return filepath.Dir(path)
}
