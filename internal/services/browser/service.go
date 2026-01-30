package browser

import (
	"willchat/internal/errs"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// BrowserService 浏览器服务（暴露给前端调用）
type BrowserService struct {
	app *application.App
}

func NewBrowserService(app *application.App) *BrowserService {
	return &BrowserService{app: app}
}

// OpenURL 在系统默认浏览器中打开 URL
func (s *BrowserService) OpenURL(url string) error {
	if url == "" {
		return errs.New("error.browser_url_required")
	}
	if err := s.app.Browser.OpenURL(url); err != nil {
		return errs.Wrap("error.browser_open_failed", err)
	}
	return nil
}
