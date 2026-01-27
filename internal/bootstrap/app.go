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
	// 创建实例的时候立即注册的服务
	services := []application.Service{
		application.NewService(greet.NewGreetService("Hello, ")),
	}

	app := application.New(application.Options{
		Name:        "WillChat",
		Description: "WillChat Desktop App",
		Services:    services,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(opts.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 创建实例后才能注册的服务
	windowService, err := windows.NewWindowService(app, windows.WindowServiceOptions{
		Definitions: windows.DefaultDefinitions(),
		Precreate: []string{
			windows.WindowMain,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("init window service: %w", err)
	}
	app.RegisterService(application.NewService(windowService))

	// 系统托盘
	systrayMenu := app.NewMenu()
	systrayMenu.Add("Show").OnClick(func(ctx *application.Context) {
		_ = windowService.Show(windows.WindowMain)
	})
	systrayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	app.SystemTray.New().SetIcon(opts.Icon).SetMenu(systrayMenu)

	return app, nil
}
