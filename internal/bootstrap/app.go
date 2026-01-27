package bootstrap

import (
	"fmt"
	"io/fs"

	"changeme/internal/services/greet"
	"changeme/internal/services/windows"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Options struct {
	Assets fs.FS
	Icon   []byte
}

func NewApp(opts Options) (*application.App, error) {
	// 创建应用实例
	app := application.New(application.Options{
		Name:        "WillChat",
		Description: "WillChat Desktop App",
		Services: []application.Service{
			application.NewService(greet.NewGreetService("Hello, ")),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(opts.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 创建主窗口
	mainWindow := windows.NewMainWindow(app)

	// 创建子窗口服务
	windowService, err := windows.NewWindowService(app, windows.DefaultDefinitions())
	if err != nil {
		return nil, fmt.Errorf("init window service: %w", err)
	}
	app.RegisterService(application.NewService(windowService))

	// 创建系统托盘
	systrayMenu := app.NewMenu()
	systrayMenu.Add("Show").OnClick(func(ctx *application.Context) {
		mainWindow.Show()
		mainWindow.Focus()
	})
	systrayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	app.SystemTray.New().SetIcon(opts.Icon).SetMenu(systrayMenu)

	return app, nil
}
