package windows

import "github.com/wailsapp/wails/v3/pkg/application"

// NewMainWindow 创建主窗口（主窗口是特殊窗口，不纳入子窗口 WindowService 管理）。
func NewMainWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:  "main",
		Title: "WillChat",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})
}

