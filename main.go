package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"

	"willchat/internal/sqlite"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/sysicon.png
var icon []byte

func init() {
	// application.RegisterEvent[string]("time")
}

func main() {
	app := application.New(application.Options{
		Name:        "WillChat",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "main",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// 创建系统托盘
	systrayMenu := app.NewMenu()
	systrayMenu.Add("Show").OnClick(func(ctx *application.Context) {
		mainWindow.Show()
	})
	systrayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	app.SystemTray.New().SetIcon(icon).SetMenu(systrayMenu)

	// ========== 数据库测试代码 开始 ==========
	if err := sqlite.Init(app); err != nil {
		log.Fatal("sqlite init failed:", err)
	}
	defer sqlite.Close(app)

	// 测试基本查询
	var sqliteVersion, vecVersion string
	err := sqlite.DB().QueryRow("SELECT sqlite_version(), vec_version()").Scan(&sqliteVersion, &vecVersion)
	if err != nil {
		log.Fatal("query failed:", err)
	}
	fmt.Printf("SQLite version: %s\n", sqliteVersion)
	fmt.Printf("sqlite-vec version: %s\n", vecVersion)

	// 测试向量操作
	var vecResult string
	err = sqlite.DB().QueryRow("SELECT vec_to_json(vec_f32('[1.0, 2.0, 3.0]'))").Scan(&vecResult)
	if err != nil {
		log.Fatal("vec test failed:", err)
	}
	fmt.Printf("Vector test: %s\n", vecResult)

	// 测试 FTS5
	_, err = sqlite.DB().Exec(`
		DROP TABLE IF EXISTS test_fts;
		CREATE VIRTUAL TABLE test_fts USING fts5(content);
		INSERT INTO test_fts(content) VALUES ('hello world test');
		INSERT INTO test_fts(content) VALUES ('中文测试内容');
	`)
	if err != nil {
		log.Fatal("FTS5 create failed:", err)
	}
	var ftsResult string
	err = sqlite.DB().QueryRow("SELECT content FROM test_fts WHERE test_fts MATCH 'hello'").Scan(&ftsResult)
	if err != nil {
		log.Fatal("FTS5 query failed:", err)
	}
	fmt.Printf("FTS5 test (English): %s\n", ftsResult)
	
	// 中文需要用字符级分词，这里验证 FTS5 模块可用即可
	var ftsCount int
	err = sqlite.DB().QueryRow("SELECT COUNT(*) FROM test_fts").Scan(&ftsCount)
	if err != nil {
		log.Fatal("FTS5 count failed:", err)
	}
	fmt.Printf("FTS5 test (row count): %d\n", ftsCount)

	// 清理测试表
	sqlite.DB().Exec("DROP TABLE IF EXISTS test_fts")

	fmt.Println("\n✅ 数据库测试全部通过！")
	fmt.Printf("📁 数据库路径: %s\n", sqlite.Path())
	// ========== 数据库测试代码 结束 ==========

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
