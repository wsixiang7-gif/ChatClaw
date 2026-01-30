package tray

import (
	"context"
	"sync"
	"time"

	"willchat/internal/sqlite"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// TrayService 托盘服务（暴露给前端调用）
type TrayService struct {
	app     *application.App
	systray *application.SystemTray

	mu                    sync.RWMutex
	trayIconEnabled       bool
	minimizeToTrayEnabled bool
}

func NewTrayService(app *application.App, systray *application.SystemTray) *TrayService {
	return &TrayService{
		app:     app,
		systray: systray,
		// 默认值与 migration 保持一致：开启托盘 + 关闭最小化到托盘
		trayIconEnabled:       true,
		minimizeToTrayEnabled: true,
	}
}

// SetSystemTray 注入系统托盘实例（用于在应用启动后再创建托盘的场景）
// 注意：此方法也会暴露给前端 bindings，但前端无需调用。
func (s *TrayService) SetSystemTray(systray *application.SystemTray) {
	s.mu.Lock()
	s.systray = systray
	s.mu.Unlock()
}

// SetVisible 设置托盘图标是否可见
func (s *TrayService) SetVisible(visible bool) {
	s.mu.Lock()
	s.trayIconEnabled = visible
	systray := s.systray
	s.mu.Unlock()

	if systray == nil {
		return
	}
	if visible {
		systray.Show()
	} else {
		systray.Hide()
	}
}

// SetMinimizeToTrayEnabled 设置是否启用"关闭时最小化到托盘"
func (s *TrayService) SetMinimizeToTrayEnabled(enabled bool) {
	s.mu.Lock()
	s.minimizeToTrayEnabled = enabled
	s.mu.Unlock()
}

// IsMinimizeToTrayEnabled 检查是否启用了"关闭时最小化到托盘"
func (s *TrayService) IsMinimizeToTrayEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.minimizeToTrayEnabled
}

// IsTrayIconEnabled 检查是否启用了"显示托盘图标"
func (s *TrayService) IsTrayIconEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trayIconEnabled
}

// InitFromSettings 根据设置初始化托盘状态
func (s *TrayService) InitFromSettings() {
	// 启动时从 DB 读一次，然后后续只走内存缓存（避免在窗口关闭时查库造成卡顿）
	trayVisible := s.readSettingBool("show_tray_icon", true)
	minimizeEnabled := s.readSettingBool("minimize_to_tray_on_close", true)

	s.mu.Lock()
	s.trayIconEnabled = trayVisible
	s.minimizeToTrayEnabled = minimizeEnabled
	systray := s.systray
	s.mu.Unlock()

	if systray == nil {
		return
	}
	if trayVisible {
		systray.Show()
	} else {
		systray.Hide()
	}
}

func (s *TrayService) readSettingBool(key string, defaultValue bool) bool {
	db := sqlite.DB()
	if db == nil {
		return defaultValue
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value string
	err := db.NewSelect().
		Table("settings").
		Column("value").
		Where("key = ?", key).
		Scan(ctx, &value)
	if err != nil {
		return defaultValue
	}
	return value == "true"
}
