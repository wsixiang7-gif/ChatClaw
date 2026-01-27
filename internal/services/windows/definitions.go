package windows

import "github.com/wailsapp/wails/v3/pkg/application"

// DefaultDefinitions 提供子窗口的默认定义
func DefaultDefinitions() []WindowDefinition {
	return []WindowDefinition{
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

