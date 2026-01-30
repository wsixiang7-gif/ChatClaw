package main

import (
	"embed"
	_ "embed"
	"log"
	"runtime"

	"willchat/internal/bootstrap"
	"willchat/internal/sqlite"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/sysicon.png
var iconPNG []byte

// Windows tray icons are most reliable with .ico
//
//go:embed build/windows/icon.ico
var iconICO []byte

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)

	// application.RegisterEvent[string]("time")
}

func main() {
	appIcon := iconPNG
	if runtime.GOOS == "windows" && len(iconICO) > 0 {
		appIcon = iconICO
	}

	app, err := bootstrap.NewApp(bootstrap.Options{
		Assets: assets,
		Icon:   appIcon,
		// Locale 为空时自动检测系统语言
	})
	if err != nil {
		log.Fatal(err)
	}

	// 初始化 SQLite（跑迁移/默认设置）
	// NOTE: 必须在 app.Run() 前完成，这样后续托盘/关闭逻辑读取设置时 DB 已就绪。
	if err := sqlite.Init(app); err != nil {
		log.Fatal("sqlite init failed:", err)
	}
	defer sqlite.Close(app)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
