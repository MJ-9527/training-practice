// internal/workerpool/workerpool.go

package workerpool

import (
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

// worker 是一个工作goroutine，它从TaskChan接收任务并执行。
func (wp *WorkerPool) worker() {
	defer wp.wg.Done() // 当worker退出时，通知WaitGroup

	// 无限循环，从TaskChan接收任务
	for task := range wp.TaskChan {
		var result Result
		result.Task = task

		// 调用核心的文件复制函数
		err := fileops.CopyFile(task.SourcePath, task.DestPath, true) // 假设overwrite逻辑在分发任务时已确定
		if err != nil {
			result.Success = false
			result.Error = err
		} else {
			result.Success = true
		}

		// 将执行结果发送到ResultChan
		wp.ResultChan <- result
	}
}

// Close 关闭任务通道，并等待所有worker完成剩余任务。
func (wp *WorkerPool) Close() {
	close(wp.TaskChan)   // 关闭任务通道，告诉所有worker没有新任务了
	wp.wg.Wait()         // 等待所有worker处理完通道中剩余的任务
	close(wp.ResultChan) // 所有结果都已发送完毕，可以关闭结果通道
}
