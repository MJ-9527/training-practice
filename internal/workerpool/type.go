// internal/workerpool/types.go

package workerpool

// Task 定义了一个工作任务。
// 对于复制操作，它包含源路径和目标路径。
type Task struct {
	SourcePath string
	DestPath   string
	// 可以扩展 OpType 字段，以支持未来的 move, rename 等操作
	// OpType     string
}

// Result 定义了一个任务执行后的结果。
type Result struct {
	Task       Task  // 关联的任务
	Success    bool  // 是否成功
	Error      error // 如果失败，记录错误信息
	Type       TaskType
	SourcePath string
	DestPath   string
}

type TaskType string

const (
	TaskTypeCopy   TaskType = "copy"
	TaskTypeMove   TaskType = "move"
	TaskTypeRename TaskType = "rename"
)
