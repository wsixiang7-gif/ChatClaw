package windows

import (
	"errors"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	WindowMain     = "main"
	WindowSettings = "settings"
)

type WindowInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Created bool   `json:"created"`
	Visible bool   `json:"visible"`
}

type WindowDefinition struct {
	// Name 必须唯一，用于前端控制/设置项映射
	Name string

	// CreateOptions 用于生成该窗口的创建参数（支持后续做更复杂的窗口配置）
	CreateOptions func() application.WebviewWindowOptions

	// FocusOnShow 为 true 时，Show 后会自动 Focus
	FocusOnShow bool
}

type WindowServiceOptions struct {
	Definitions []WindowDefinition

	// 用于应用启动时预创建的窗口 Name 列表
	Precreate []string
}

// WindowService 暴露给前端的多窗口控制 API
// 前端通过 bindings 调用（例如 WindowService.Show("settings")）
type WindowService struct {
	app *application.App

	mu      sync.RWMutex
	defs    map[string]WindowDefinition
	windows map[string]*application.WebviewWindow
}

func NewWindowService(app *application.App, opts WindowServiceOptions) (*WindowService, error) {
	if app == nil {
		return nil, errors.New("app is required")
	}
	s := &WindowService{
		app:     app,
		defs:    make(map[string]WindowDefinition),
		windows: make(map[string]*application.WebviewWindow),
	}
	for _, def := range opts.Definitions {
		if err := s.register(def); err != nil {
			return nil, err
		}
	}
	for _, name := range opts.Precreate {
		if _, err := s.ensure(name); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *WindowService) register(def WindowDefinition) error {
	if def.Name == "" {
		return errors.New("window name is required")
	}
	if def.CreateOptions == nil {
		return fmt.Errorf("window '%s' CreateOptions is required", def.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.defs[def.Name]; exists {
		return fmt.Errorf("window '%s' already registered", def.Name)
	}
	s.defs[def.Name] = def
	return nil
}

func (s *WindowService) ensure(name string) (*application.WebviewWindow, error) {
	s.mu.RLock()
	if existing := s.windows[name]; existing != nil {
		s.mu.RUnlock()
		return existing, nil
	}
	def, ok := s.defs[name]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("window '%s' not registered", name)
	}

	options := def.CreateOptions()
	if options.Name == "" {
		options.Name = name
	}

	w := s.app.Window.NewWithOptions(options)

	// 窗口被 Close 时会从 Wails 的 WindowManager 中移除，但我们这里也需要清理缓存，
	// 否则下次 Show/ensure 会拿到“已 destroyed”的旧指针（无法复用）。
	w.OnWindowEvent(events.Common.WindowClosing, func(_ *application.WindowEvent) {
		s.mu.Lock()
		delete(s.windows, name)
		s.mu.Unlock()
	})

	s.mu.Lock()
	s.windows[name] = w
	s.mu.Unlock()

	return w, nil
}

func (s *WindowService) get(name string) (*application.WebviewWindow, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w := s.windows[name]
	return w, w != nil
}

func (s *WindowService) List() []WindowInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]WindowInfo, 0, len(s.defs))
	for name, def := range s.defs {
		info := WindowInfo{
			Name:    name,
			Created: s.windows[name] != nil,
			Visible: false,
		}

		// 从定义里拿 Title/URL（比从窗口实例拿更稳定，且未创建时也能展示）
		opts := def.CreateOptions()
		info.Title = opts.Title
		info.URL = opts.URL

		if w := s.windows[name]; w != nil {
			info.Visible = w.IsVisible()
		}
		result = append(result, info)
	}
	return result
}

func (s *WindowService) Show(name string) error {
	s.mu.RLock()
	def, ok := s.defs[name]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("window '%s' not registered", name)
	}

	w, err := s.ensure(name)
	if err != nil {
		return err
	}
	w.Show()
	if def.FocusOnShow {
		w.Focus()
	}
	return nil
}

func (s *WindowService) Hide(name string) error {
	// 不存在则直接返回 nil，便于前端“幂等隐藏”
	w, ok := s.get(name)
	if !ok {
		return nil
	}
	w.Hide()
	return nil
}

func (s *WindowService) Close(name string) error {
	w, ok := s.get(name)
	if !ok {
		return nil
	}
	// 先移除缓存，避免 Close 过程中再次被并发访问拿到旧指针
	s.mu.Lock()
	delete(s.windows, name)
	s.mu.Unlock()

	w.Close()
	return nil
}

func (s *WindowService) IsVisible(name string) (bool, error) {
	// 未创建直接返回 false
	w, ok := s.get(name)
	if !ok {
		// 若未注册，返回错误，便于前端发现拼写问题
		s.mu.RLock()
		_, registered := s.defs[name]
		s.mu.RUnlock()
		if !registered {
			return false, fmt.Errorf("window '%s' not registered", name)
		}
		return false, nil
	}
	return w.IsVisible(), nil
}

func (s *WindowService) SetVisible(name string, visible bool) (bool, error) {
	if visible {
		if err := s.Show(name); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := s.Hide(name); err != nil {
		return false, err
	}
	return false, nil
}

func (s *WindowService) Toggle(name string) (bool, error) {
	current, err := s.IsVisible(name)
	if err != nil {
		return false, err
	}
	return s.SetVisible(name, !current)
}
