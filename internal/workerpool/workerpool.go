// internal/workerpool/workerpool.go

package workerpool

import (
	"fmt"
	"sync"

	// 导入我们的文件操作包
	"training-practice/internal/fileops"
)

// WorkerPool 管理一组工作goroutine来并发处理任务。
type WorkerPool struct {
	TaskChan   chan Task      // 用于接收任务的通道
	ResultChan chan Result    // 用于发送结果的通道
	wg         sync.WaitGroup // 用于等待所有worker完成
}

// NewWorkerPool 创建一个新的WorkerPool实例。
// workerCount: 工作goroutine的数量。
// bufferSize: 任务和结果通道的缓冲区大小。
func NewWorkerPool(workerCount, bufferSize int) *WorkerPool {
	wp := &WorkerPool{
		TaskChan:   make(chan Task, bufferSize),
		ResultChan: make(chan Result, bufferSize),
	}

	// 根据指定数量启动worker
	for i := 0; i < workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker() // 启动一个worker goroutine
	}

	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for task := range wp.TaskChan {
		var result Result
		result.Task = task
		var err error

		// 根据任务类型分发到不同的操作函数
		switch task.Type {
		case TaskTypeCopy:
			err = fileops.CopyFile(task.SourcePath, task.DestPath, result.Task.Overwrite)
		case TaskTypeMove:
			err = fileops.MoveFile(task.SourcePath, task.DestPath, result.Task.Overwrite)
		case TaskTypeRename:
			err = fileops.RenameFile(task.SourcePath, task.DestPath, result.Task.Overwrite)
		default:
			err = fmt.Errorf("未知的任务类型: %s", task.Type)
		}

		if err != nil {
			result.Success = false
			result.Error = err
		} else {
			result.Success = true
		}

		wp.ResultChan <- result
	}
}

// Close 关闭任务通道，并等待所有worker完成剩余任务。
func (wp *WorkerPool) Close() {
	close(wp.TaskChan)   // 关闭任务通道，告诉所有worker没有新任务了
	wp.wg.Wait()         // 等待所有worker处理完通道中剩余的任务
	close(wp.ResultChan) // 所有结果都已发送完毕，可以关闭结果通道
}
