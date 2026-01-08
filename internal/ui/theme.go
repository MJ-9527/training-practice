package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// customTheme 自定义主题：增大字体+黑色文本
type customTheme struct {
	fyne.Theme
}

func (m *customTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 14 // 增大字体到14
	}
	return m.Theme.Size(name)
}

func (m *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground, theme.ColorNameDisabled:
		return color.NRGBA{0, 0, 0, 255} // 强制文本黑色
	default:
		return m.Theme.Color(name, variant)
	}
}

// NewCustomTheme 创建自定义主题实例
func NewCustomTheme() fyne.Theme {
	return &customTheme{Theme: theme.DefaultTheme()}
}
