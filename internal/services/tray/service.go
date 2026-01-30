package tray

import (
	"sync"

	"willchat/internal/services/settings"

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

// SetVisible 设置托盘图标是否可见
func (s *TrayService) SetVisible(visible bool) {
	s.mu.Lock()
	s.trayIconEnabled = visible
	s.mu.Unlock()

	if visible {
		s.systray.Show()
	} else {
		s.systray.Hide()
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
	// 从 settings 内存缓存读取（不走 DB）
	trayVisible := settings.GetBool("show_tray_icon", true)
	minimizeEnabled := settings.GetBool("minimize_to_tray_on_close", true)

	s.mu.Lock()
	s.trayIconEnabled = trayVisible
	s.minimizeToTrayEnabled = minimizeEnabled
	s.mu.Unlock()

	if trayVisible {
		s.systray.Show()
	} else {
		s.systray.Hide()
	}
}
