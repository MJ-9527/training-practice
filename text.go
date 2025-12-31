package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type Task struct {
	src string
	dst string
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func worker(id int, tasks <-chan Task, wg *sync.WaitGroup, done *int32) {
	defer wg.Done()
	for task := range tasks {
		err := copyFile(task.src, task.dst)
		if err != nil {
			fmt.Printf("Worker %d error: %v\n", id, err)
		}
		atomic.AddInt32(done, 1)
	}
}

func main() {
	srcDir := flag.String("src", "", "source directory")
	dstDir := flag.String("dst", "", "destination directory")
	workers := flag.Int("workers", 4, "number of workers")
	flag.Parse()

	if *srcDir == "" || *dstDir == "" {
		fmt.Println("Usage: filetool -src <source> -dst <dest> -workers <n>")
		return
	}

	var tasks = make(chan Task, 100)
	var wg sync.WaitGroup

	var total int32
	var done int32

	// 启动 worker pool
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(i, tasks, &wg, &done)
	}

	// 扫描目录
	go func() {
		filepath.Walk(*srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			rel, _ := filepath.Rel(*srcDir, path)
			dstPath := filepath.Join(*dstDir, rel)

			atomic.AddInt32(&total, 1)
			tasks <- Task{src: path, dst: dstPath}
			return nil
		})
		close(tasks)
	}()

	// 进度监控
	go func() {
		for {
			if atomic.LoadInt32(&done) == atomic.LoadInt32(&total) {
				return
			}
			fmt.Printf("\rProgress: %d / %d", done, total)
		}
	}()

	wg.Wait()
	fmt.Printf("\nDone! Copied %d files.\n", total)
}
