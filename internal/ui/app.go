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
	myWindow.Resize(fyne.NewSize(900, 700))

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
		"移动",
	}, nil)
	modeRadio.SetSelected("计算MD5")

	// 异常策略设置
	errorPolicySelect := widget.NewSelect([]string{
		"跳过错误文件",
		"重试错误文件",
		"终止整个任务",
	}, nil)
	errorPolicySelect.SetSelected("跳过错误文件")

	// 重试设置
	maxRetriesLabel := widget.NewLabel("3")
	maxRetriesLabel.Alignment = fyne.TextAlignCenter
	maxRetriesSlider := widget.NewSlider(0, 10)
	maxRetriesSlider.SetValue(3)
	maxRetriesSlider.OnChanged = func(v float64) {
		maxRetriesLabel.SetText(fmt.Sprintf("%d", int(v)))
	}

	retryIntervalLabel := widget.NewLabel("2秒")
	retryIntervalLabel.Alignment = fyne.TextAlignCenter
	retryIntervalSlider := widget.NewSlider(1, 30)
	retryIntervalSlider.SetValue(2)
	retryIntervalSlider.OnChanged = func(v float64) {
		retryIntervalLabel.SetText(fmt.Sprintf("%.0f秒", v))
	}

	// 错误类型策略表格
	errorTypes := []string{
		"文件不存在",
		"权限不足",
		"磁盘空间不足",
		"读取错误",
		"写入错误",
		"跨设备错误",
		"未知错误",
	}

	errorPolicyOptions := []string{"跳过", "重试", "终止"}

	var errorPolicyWidgets []*widget.Select
	errorPolicyGrid := container.NewGridWithColumns(2)
	for _, errorType := range errorTypes {
		label := widget.NewLabel(errorType)
		policySelect := widget.NewSelect(errorPolicyOptions, nil)
		policySelect.SetSelected("重试")
		if errorType == "文件不存在" {
			policySelect.SetSelected("跳过")
		} else if errorType == "磁盘空间不足" {
			policySelect.SetSelected("终止")
		}
		errorPolicyWidgets = append(errorPolicyWidgets, policySelect)
		errorPolicyGrid.Add(label)
		errorPolicyGrid.Add(policySelect)
	}

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

	// 统计信息
	statsLabel := widget.NewLabel("就绪")
	statsLabel.Alignment = fyne.TextAlignCenter

	var selectedSrcDir string
	var selectedDestDir string
	var errorHandler *fileutil.ErrorHandler
	var abortFlag bool
	var abortMutex sync.Mutex

	// 更新日志辅助函数
	updateLog := func(text string) {
		logEntry.SetText(logEntry.Text + text)
		logEntry.CursorRow = len(logEntry.Text)
	}

	// 更新统计信息
	updateStats := func(success, skipped, failed, total int) {
		statsLabel.SetText(fmt.Sprintf("进度: %d/%d | 成功: %d | 跳过: %d | 失败: %d",
			success+skipped+failed, total, success, skipped, failed))
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
		if mode == "复制" || mode == "复制+重命名" || mode == "移动" {
			if selectedDestDir == "" {
				dialog.ShowError(fmt.Errorf("请先选择目标文件夹"), myWindow)
				return
			}
		}

		// 初始化错误处理器
		errorHandler = fileutil.NewErrorHandler()
		errorHandler.SetMaxRetries(int(maxRetriesSlider.Value))
		errorHandler.SetRetryInterval(time.Duration(retryIntervalSlider.Value) * time.Second)

		// 设置错误策略
		policyMap := map[string]fileutil.ErrorPolicy{
			"跳过": fileutil.PolicySkip,
			"重试": fileutil.PolicyRetry,
			"终止": fileutil.PolicyAbort,
		}

		for i, errorType := range errorTypes {
			policy := errorPolicyWidgets[i].Selected
			if policy != "" {
				var et fileutil.ErrorType
				switch errorType {
				case "文件不存在":
					et = fileutil.ErrorFileNotFound
				case "权限不足":
					et = fileutil.ErrorPermissionDenied
				case "磁盘空间不足":
					et = fileutil.ErrorDiskSpaceFull
				case "读取错误":
					et = fileutil.ErrorIORead
				case "写入错误":
					et = fileutil.ErrorIOWrite
				case "跨设备错误":
					et = fileutil.ErrorCrossDevice
				case "未知错误":
					et = fileutil.ErrorUnknown
				}
				errorHandler.SetPolicy(et, policyMap[policy])
			}
		}

		// 重置中止标志
		abortMutex.Lock()
		abortFlag = false
		abortMutex.Unlock()

		logEntry.SetText("")
		updateLog("开始扫描并处理...\n")
		updateLog(fmt.Sprintf("异常策略: %s\n", errorPolicySelect.Selected))
		updateLog(fmt.Sprintf("最大重试次数: %d\n", int(maxRetriesSlider.Value)))
		updateLog(fmt.Sprintf("重试间隔: %.0f秒\n", retryIntervalSlider.Value))

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

		// 统计信息
		successCount := 0
		skippedCount := 0
		failedCount := 0

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
		case "移动":
			modeCode = "move"
		}

		// 启动Worker Pool
		numWorkers := int(workerSlider.Value)
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() { // 移除参数i
				defer wg.Done()
				for t := range tasks {
					// 检查是否中止
					abortMutex.Lock()
					if abortFlag {
						abortMutex.Unlock()
						return
					}
					abortMutex.Unlock()

					// 使用带重试的处理
					res := fileutil.ProcessFileWithRetry(t,
						int(maxRetriesSlider.Value),
						time.Duration(retryIntervalSlider.Value)*time.Second,
						errorHandler.HandleError)

					results <- res
				}
			}() // 移除参数i
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
			for res := range results {
				abortMutex.Lock()
				shouldAbort := abortFlag
				abortMutex.Unlock()

				if shouldAbort {
					continue
				}

				// 更新统计
				if res.Err == nil {
					successCount++
				} else if res.Skipped {
					skippedCount++
				} else {
					failedCount++
				}

				currentProgress := successCount + skippedCount + failedCount
				progressBar.SetValue(float64(currentProgress))
				updateStats(successCount, skippedCount, failedCount, total)

				// 更新日志
				status := "成功"
				if res.Err != nil {
					if res.Skipped {
						status = "跳过"
					} else {
						status = fmt.Sprintf("失败: %v", res.Err)
					}
				}

				retryInfo := ""
				if res.Retried > 0 {
					retryInfo = fmt.Sprintf(" (重试%d次)", res.Retried)
				}

				verifyStr := ""
				if res.DstMD5 != "" {
					if res.Verified {
						verifyStr = " | 校验: 一致"
					} else {
						verifyStr = " | 校验: 不一致"
					}
				}

				updateLog(fmt.Sprintf("[%d/%d] %s -> %s | 源MD5: %s",
					currentProgress, total, res.OldName, res.NewName, res.SrcMD5))
				if res.DstMD5 != "" {
					updateLog(fmt.Sprintf(" 目标MD5: %s%s", res.DstMD5, verifyStr))
				}
				updateLog(fmt.Sprintf("%s | %s\n", retryInfo, status))

				// 检查是否需要中止
				if res.Err != nil && !res.Skipped {
					// 检查错误类型
					errorInfo := fileutil.AnalyzeError(res.Err, res.OldName)
					policy := errorHandler.HandleError(errorInfo)

					if policy == fileutil.PolicyAbort {
						abortMutex.Lock()
						abortFlag = true
						abortMutex.Unlock()
						updateLog("\n⚠️ 检测到严重错误，任务已中止！\n")
						break
					}
				}
			}

			duration := time.Since(startTime)
			updateLog(fmt.Sprintf("\n任务结束！耗时: %v\n", duration))

			// 显示最终统计
			finalStats := fmt.Sprintf("\n最终统计: 成功 %d, 跳过 %d, 失败 %d / 总计 %d",
				successCount, skippedCount, failedCount, total)
			updateLog(finalStats)
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

	// 异常策略设置区域
	errorPolicyGroup := container.NewVBox(
		widget.NewLabelWithStyle("异常处理策略", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(
			widget.NewLabelWithStyle("全局策略", fyne.TextAlignLeading, fyne.TextStyle{}),
			nil, nil, nil,
			errorPolicySelect,
		),
		container.NewBorder(
			widget.NewLabelWithStyle("最大重试次数", fyne.TextAlignLeading, fyne.TextStyle{}),
			maxRetriesLabel, nil, nil,
			maxRetriesSlider,
		),
		container.NewBorder(
			widget.NewLabelWithStyle("重试间隔", fyne.TextAlignLeading, fyne.TextStyle{}),
			retryIntervalLabel, nil, nil,
			retryIntervalSlider,
		),
		widget.NewLabelWithStyle("按错误类型设置策略:", fyne.TextAlignLeading, fyne.TextStyle{}),
		container.NewScroll(errorPolicyGrid),
	)

	// 动态更新UI
	updateUI := func(mode string) {
		destGroupContainer.RemoveAll()
		renameGroup.Hide()

		if mode == "复制" || mode == "复制+重命名" || mode == "移动" {
			destGroupContainer.Add(destGroupContent)
		}
		if mode == "重命名" || mode == "复制+重命名" || mode == "移动" {
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

		// 异常策略设置
		errorPolicyGroup,

		widget.NewSeparator(),

		// 执行按钮
		func() *widget.Button {
			btn := widget.NewButton("开始执行", startProcess)
			btn.Importance = widget.HighImportance
			return btn
		}(),

		// 进度条
		progressBar,

		// 统计信息
		statsLabel,
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
	content.SetOffset(0.5)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
