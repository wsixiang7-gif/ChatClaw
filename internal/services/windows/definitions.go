package windows

import "github.com/wailsapp/wails/v3/pkg/application"

// DefaultDefinitions 提供应用内置窗口的默认定义。
// main.go 里只需要引用这个即可，避免堆一大段窗口 options。
func DefaultDefinitions() []WindowDefinition {
	return []WindowDefinition{
		{
			Name: WindowMain,
			CreateOptions: func() application.WebviewWindowOptions {
				return application.WebviewWindowOptions{
					Name:  WindowMain,
					Title: "WillChat",
					Mac: application.MacWindow{
						InvisibleTitleBarHeight: 50,
						Backdrop:                application.MacBackdropTranslucent,
						TitleBar:                application.MacTitleBarHiddenInset,
					},
					BackgroundColour: application.NewRGB(27, 38, 54),
					URL:              "/",
				}
			},
			FocusOnShow: true,
		},
		{
			Name: WindowSettings,
			CreateOptions: func() application.WebviewWindowOptions {
				return application.WebviewWindowOptions{
					Name:   WindowSettings,
					Title:  "Settings",
					Width:  600,
					Height: 400,
					Hidden: true,
					// 多页面入口：对应 frontend/settings.html
					URL: "/settings.html",
				}
			},
			FocusOnShow: true,
		},
	}
}

