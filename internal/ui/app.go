package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"training-practice/internal/fileutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Run 启动文件处理工具的UI界面
func Run() {
	myApp := app.New()
	myApp.Settings().SetTheme(NewCustomTheme())
	myWindow := myApp.NewWindow("高并发文件处理工具")
	myWindow.Resize(fyne.NewSize(800, 650))

	// --- UI控件定义 ---
	// 源目录选择
	srcPathLabel := widget.NewLabel("未选择")
	srcPathLabel.Wrapping = fyne.TextWrapWord
	srcPathLabel.Truncation = fyne.TextTruncateEllipsis

	// 重命名选项
	prefixEntry := widget.NewEntry()
	prefixEntry.SetPlaceHolder("重命名加前缀（可选）...")
	suffixEntry := widget.NewEntry()
	suffixEntry.SetPlaceHolder("重命名加后缀（可选）...")

	// 操作模式选择
	modeRadio := widget.NewRadioGroup([]string{
		"计算MD5",
		"重命名",
		"复制",
		"复制+重命名",
	}, nil)
	modeRadio.SetSelected("计算MD5")

	// 重命名选项区域
	renameGroup := container.NewBorder(
		widget.NewLabelWithStyle("重命名设置", fyne.TextAlignLeading, fyne.TextStyle{}),
		nil, nil, nil,
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("前缀:"), prefixEntry),
			container.NewVBox(widget.NewLabel("后缀:"), suffixEntry),
		),
	)

	// 目标目录选择区域
	destPathLabel := widget.NewLabel("未选择")
	destPathLabel.Wrapping = fyne.TextWrapWord
	destPathLabel.Truncation = fyne.TextTruncateEllipsis
	destGroupContainer := container.NewVBox()

	// 并发设置
	workerCountLabel := widget.NewLabel("4")
	workerCountLabel.Alignment = fyne.TextAlignCenter
	workerSlider := widget.NewSlider(1, 20)
	workerSlider.SetValue(4)
	workerSlider.OnChanged = func(v float64) {
		workerCountLabel.SetText(fmt.Sprintf("%d", int(v)))
	}

	// 日志显示
	logEntry := widget.NewMultiLineEntry()
	logEntry.Disable()
	logEntry.Wrapping = fyne.TextWrapWord

	// 进度条
	progressBar := widget.NewProgressBar()

	var selectedSrcDir string
	var selectedDestDir string

	// 更新日志辅助函数
	updateLog := func(text string) {
		logEntry.SetText(logEntry.Text + text)
		logEntry.CursorRow = len(logEntry.Text)
	}

	// --- 核心处理逻辑 ---
	startProcess := func() {
		if selectedSrcDir == "" {
			dialog.ShowError(fmt.Errorf("请先选择源文件夹"), myWindow)
			return
		}

		mode := modeRadio.Selected
		if mode == "" {
			dialog.ShowError(fmt.Errorf("请选择操作模式"), myWindow)
			return
		}

		// 检查目标目录
		if mode == "复制" || mode == "复制+重命名" {
			if selectedDestDir == "" {
				dialog.ShowError(fmt.Errorf("请先选择目标文件夹"), myWindow)
				return
			}
		}

		logEntry.SetText("")
		updateLog("开始扫描并处理...\n")
		startTime := time.Now()

		// 扫描所有文件
		var files []string
		filepath.Walk(selectedSrcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			files = append(files, path)
			return nil
		})

		total := len(files)
		if total == 0 {
			dialog.ShowError(fmt.Errorf("源目录中没有找到文件"), myWindow)
			return
		}

		progressBar.Max = float64(total)
		progressBar.SetValue(0)

		// 任务/结果通道
		tasks := make(chan fileutil.Task, total)
		results := make(chan fileutil.Result, total)
		var wg sync.WaitGroup

		// 转换操作模式
		var modeCode string
		switch mode {
		case "计算MD5":
			modeCode = "md5"
		case "重命名":
			modeCode = "rename"
		case "复制":
			modeCode = "copy"
		case "复制+重命名":
			modeCode = "copy_rename"
		}

		// 启动Worker Pool
		numWorkers := int(workerSlider.Value)
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for t := range tasks {
					res := fileutil.ProcessFile(t)
					results <- res
				}
			}()
		}

		// 分发任务
		go func() {
			for _, f := range files {
				tasks <- fileutil.Task{
					Path:     f,
					SrcRoot:  selectedSrcDir,
					DestRoot: selectedDestDir,
					Prefix:   prefixEntry.Text,
					Suffix:   suffixEntry.Text,
					Mode:     modeCode,
				}
			}
			close(tasks)
		}()

		// 等待Worker完成并关闭结果通道
		go func() {
			wg.Wait()
			close(results)
		}()

		// 更新UI进度和日志
		go func() {
			count := 0
			for res := range results {
				count++
				progressBar.SetValue(float64(count))
				status := "成功"
				if res.Err != nil {
					status = fmt.Sprintf("失败: %v", res.Err)
				}
				verifyStr := ""
				if res.DstMD5 != "" {
					if res.Verified {
						verifyStr = " | 校验: 一致"
					} else {
						verifyStr = " | 校验: 不一致"
					}
				}
				updateLog(fmt.Sprintf("[%d/%d] %s -> %s | 源MD5: %s", count, total, res.OldName, res.NewName, res.SrcMD5))
				if res.DstMD5 != "" {
					updateLog(fmt.Sprintf(" 目标MD5: %s%s", res.DstMD5, verifyStr))
				}
				updateLog(fmt.Sprintf(" | %s\n", status))
			}
			duration := time.Since(startTime)
			updateLog(fmt.Sprintf("\n任务结束！耗时: %v\n", duration))
		}()
	}

	// --- 按钮与布局逻辑 ---
	// 源目录选择按钮
	selectSrcBtn := widget.NewButton("选择源文件夹", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err == nil && list != nil {
				selectedSrcDir = list.Path()
				srcPathLabel.SetText(selectedSrcDir)
			}
		}, myWindow)
	})

	// 目标目录选择按钮
	selectDestBtn := widget.NewButton("选择目标文件夹", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err == nil && list != nil {
				selectedDestDir = list.Path()
				destPathLabel.SetText(selectedDestDir)
			}
		}, myWindow)
	})

	// 目标目录布局
	destGroupContent := container.NewBorder(
		widget.NewLabelWithStyle("目标目录", fyne.TextAlignLeading, fyne.TextStyle{}),
		nil, nil, nil,
		container.NewHBox(selectDestBtn, destPathLabel),
	)

	// 动态更新UI（根据操作模式显示/隐藏控件）
	updateUI := func(mode string) {
		destGroupContainer.RemoveAll()
		renameGroup.Hide()

		if mode == "复制" || mode == "复制+重命名" {
			destGroupContainer.Add(destGroupContent)
		}
		if mode == "重命名" || mode == "复制+重命名" {
			renameGroup.Show()
		}
	}
	modeRadio.OnChanged = updateUI
	updateUI(modeRadio.Selected)

	// 配置区域布局
	configArea := container.NewVBox(
		widget.NewLabelWithStyle("高并发文件处理工具", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		// 源目录
		container.NewBorder(
			widget.NewLabelWithStyle("源目录", fyne.TextAlignLeading, fyne.TextStyle{}),
			nil, nil, nil,
			container.NewHBox(selectSrcBtn, srcPathLabel),
		),
		widget.NewSeparator(),

		// 操作模式
		container.NewBorder(
			widget.NewLabelWithStyle("操作模式", fyne.TextAlignLeading, fyne.TextStyle{}),
			nil, nil, nil,
			modeRadio,
		),

		// 目标目录（动态）
		destGroupContainer,

		// 重命名设置（动态）
		renameGroup,

		widget.NewSeparator(),

		// 并发设置
		container.NewBorder(
			widget.NewLabelWithStyle("并发Worker数", fyne.TextAlignLeading, fyne.TextStyle{}),
			workerCountLabel, nil, nil,
			workerSlider,
		),

		widget.NewSeparator(),

		// 执行按钮
		func() *widget.Button {
			btn := widget.NewButton("开始执行", startProcess)
			btn.Importance = widget.HighImportance
			return btn
		}(),

		// 进度条
		progressBar,
	)

	// 日志区域布局
	logArea := container.NewBorder(
		widget.NewLabelWithStyle("处理日志", fyne.TextAlignLeading, fyne.TextStyle{}),
		nil, nil, nil,
		container.NewScroll(logEntry),
	)

	// 主布局（左右分栏）
	content := container.NewHSplit(
		container.NewScroll(configArea),
		logArea,
	)
	content.SetOffset(0.45)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
