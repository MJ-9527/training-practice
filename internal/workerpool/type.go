// internal/workerpool/types.go

package workerpool

import (
	"training-practice/internal/fileops"
)

// Task 定义了一个工作任务。
// 对于复制操作，它包含源路径和目标路径。
type Task struct {
	SourcePath string
	DestPath   string
	Type       TaskType
	Overwrite  bool
	Algorithm  fileops.Algorithm
}

// Result 定义了一个任务执行后的结果。
type Result struct {
	Task    Task  // 关联的任务
	Success bool  // 是否成功
	Error   error // 如果失败，记录错误信息
	Message string // 额外的信息，如校验和结果
}

type TaskType string

const (
	TaskTypeCopy   TaskType = "copy"
	TaskTypeMove   TaskType = "move"
	TaskTypeRename TaskType = "rename"
	TaskTypeChecksum TaskType = "checksum"
)
