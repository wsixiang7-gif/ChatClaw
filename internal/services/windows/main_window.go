package windows

import "github.com/wailsapp/wails/v3/pkg/application"

// NewMainWindow 创建主窗口
func NewMainWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:  "main",
		Title: "WillChat",
		// 最小尺寸限制（所有平台生效）
		MinWidth:  1064,
		MinHeight: 628,
		// 默认启动尺寸：不小于最小尺寸
		Width:  1064,
		Height: 628,
		Mac: application.MacWindow{
			// 需要与前端自绘标题栏高度一致，用于让红黄绿与前端元素在同一视觉高度。
			// 与前端 TitleBar 高度一致（当前为 40px / h-10）
			InvisibleTitleBarHeight: 40,
			Backdrop:                application.MacBackdropTranslucent,
			// HiddenInset 会把 WebView 内容整体下移（inset），导致红黄绿与前端标题栏背景/按钮难对齐；
			// 这里改为 Hidden，让 WebView 延伸到标题栏区域，由前端自绘背景与布局。
			TitleBar: application.MacTitleBarHidden,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
}
