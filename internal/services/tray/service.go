package tray

import (
	"context"
	"log/slog"
	"time"

	"willchat/internal/sqlite"

	"github.com/uptrace/bun"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TrayService 托盘服务（暴露给前端调用）
type TrayService struct {
	app     *application.App
	systray *application.SystemTray
}

func NewTrayService(app *application.App, systray *application.SystemTray) *TrayService {
	return &TrayService{
		app:     app,
		systray: systray,
	}
}

// SetVisible 设置托盘图标是否可见
func (s *TrayService) SetVisible(visible bool) {
	if visible {
		s.systray.Show()
	} else {
		s.systray.Hide()
	}
}

// IsMinimizeToTrayEnabled 检查是否启用了"关闭时最小化到托盘"
func (s *TrayService) IsMinimizeToTrayEnabled() bool {
	db := sqlite.DB()
	if db == nil {
		slog.Warn("IsMinimizeToTrayEnabled: db is nil, returning default true")
		return true // 默认启用
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value string
	err := db.NewSelect().
		Table("settings").
		Column("value").
		Where("key = ?", "minimize_to_tray_on_close").
		Scan(ctx, &value)

	if err != nil {
		slog.Warn("IsMinimizeToTrayEnabled: query error, returning default true", "error", err)
		return true // 查询失败时默认启用
	}

	result := value == "true"
	slog.Info("IsMinimizeToTrayEnabled", "value", value, "result", result)
	return result
}

// IsTrayIconEnabled 检查是否启用了"显示托盘图标"
func (s *TrayService) IsTrayIconEnabled() bool {
	db := sqlite.DB()
	if db == nil {
		return true // 默认启用
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value string
	err := db.NewSelect().
		Table("settings").
		Column("value").
		Where("key = ?", "show_tray_icon").
		Scan(ctx, &value)

	if err != nil {
		return true // 查询失败时默认启用
	}
	return value == "true"
}

// DB 返回数据库实例（用于内部查询）
func (s *TrayService) db() *bun.DB {
	return sqlite.DB()
}

// InitFromSettings 根据设置初始化托盘状态
func (s *TrayService) InitFromSettings() {
	if s.IsTrayIconEnabled() {
		s.systray.Show()
	} else {
		s.systray.Hide()
	}
}
