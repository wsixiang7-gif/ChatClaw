package floatingball

import (
	"sync"
	"time"

	"willchat/internal/services/settings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type DockSide string

const (
	DockNone  DockSide = ""
	DockLeft  DockSide = "left"
	DockRight DockSide = "right"
)

const (
	windowName = "floatingball"

	// UI/behavior tuning (DIP pixels)
	ballSize        = 64
	defaultMargin   = 0
	edgeSnapGap     = 24
	collapsedWidth  = 32
	collapsedVisible = 18

	snapDebounce   = 180 * time.Millisecond
	rehideDebounce = 450 * time.Millisecond
	idleDockDelay  = 5 * time.Second

	// 首次 Show 后延迟定位，避免 impl 未就绪导致 SetPosition 失效
	postShowRepositionDelay = 80 * time.Millisecond
	postShowRepositionTries = 25
)

// FloatingBallService 悬浮球服务（暴露给前端调用）
//
// 职责：
// - 创建/显示一个独立的悬浮球窗口（AlwaysOnTop、无边框、透明）
// - 监听 WindowDidMove：拖动到屏幕边缘后自动贴边并半隐藏
// - 鼠标移入/移出：贴边状态下展开/回缩
// - 双击：唤起主窗口
type FloatingBallService struct {
	app        *application.App
	mainWindow *application.WebviewWindow

	mu  sync.Mutex
	win *application.WebviewWindow

	visible bool
	dock    DockSide
	hovered bool
	collapsed bool
	appActive bool
	dragging bool
	dragStartX int
	dragStartY int
	dragMoved  bool

	// remember last position/state to avoid re-centering on every Show/SetVisible call
	hasLastState bool
	lastRelX     int
	lastRelY     int
	lastDock     DockSide
	lastCollapsed bool

	ignoreMoveUntil time.Time
	snapTimer       *time.Timer
	rehideTimer     *time.Timer
	idleDockTimer   *time.Timer
	repositionTimer *time.Timer
	repositionTries int
}

func (s *FloatingBallService) debugLog(msg string, fields map[string]any) {
	if s.app == nil {
		return
	}
	// 使用 Info 级别，方便在 dev 输出中直接看到
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	s.app.Logger.Info("[floatingball] "+msg, args...)
}

func NewFloatingBallService(app *application.App, mainWindow *application.WebviewWindow) *FloatingBallService {
	return &FloatingBallService{
		app:        app,
		mainWindow: mainWindow,
		visible:    true,
		dock:       DockNone,
		appActive:  true,
	}
}

// InitFromSettings 根据 settings 内存缓存初始化悬浮球显示状态
func (s *FloatingBallService) InitFromSettings() {
	visible := settings.GetBool("show_floating_window", true)
	_ = s.SetVisible(visible)
}

// IsVisible 返回悬浮球窗口是否可见
func (s *FloatingBallService) IsVisible() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.visible && s.win != nil && s.win.IsVisible()
}

// SetVisible 设置悬浮球窗口是否可见
func (s *FloatingBallService) SetVisible(visible bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.visible = visible
	if !visible {
		// 关闭时不主动创建窗口，避免“唤醒主页面”时意外弹出悬浮球
		if s.win == nil {
			s.stopTimersLocked()
			s.dock = DockNone
			s.hovered = false
			s.collapsed = false
			return nil
		}
		s.stopTimersLocked()
		// remember current state (if window exists)
		if s.win != nil {
			x, y := s.win.RelativePosition()
			s.hasLastState = true
			s.lastRelX, s.lastRelY = x, y
			s.lastDock = s.dock
			s.lastCollapsed = s.collapsed
		}
		s.win.Hide()
		s.dock = DockNone
		s.hovered = false
		s.collapsed = false
		s.dragging = false
		s.dragMoved = false
		return nil
	}

	win := s.ensureLocked()
	if win == nil {
		return nil
	}

	s.stopTimersLocked()
	s.hovered = false
	s.dragging = false
	s.dragMoved = false
	// do NOT reset dock/collapsed on non-initial show; preserve last state if available
	if !s.hasLastState {
		s.dock = DockNone
		s.collapsed = false
	} else {
		s.dock = s.lastDock
		s.collapsed = s.lastCollapsed
	}

	win.Show()
	// 首次显示时，impl 可能还没 ready；用重试机制确保定位最终生效
	s.scheduleRepositionLocked()
	// 不抢占用户焦点：初始化/切换开启仅显示，不主动 Focus()
	s.scheduleIdleDockLocked()
	return nil
}

// Hover 通知后端鼠标是否移入悬浮球（用于贴边展开/回缩）
func (s *FloatingBallService) Hover(entered bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.win == nil {
		return
	}

	s.hovered = entered

	// Cancel any pending idle dock
	if s.idleDockTimer != nil {
		s.idleDockTimer.Stop()
		s.idleDockTimer = nil
	}

	// Cancel any pending re-hide
	if s.rehideTimer != nil {
		s.rehideTimer.Stop()
		s.rehideTimer = nil
	}

	if entered {
		s.expandLocked()
		return
	}

	// Mouse left: if not docked yet, wait idleDockDelay then dock+shrink
	if s.dock == DockNone {
		s.scheduleIdleDockLocked()
		return
	}

	// Only auto re-hide when currently docked
	s.rehideTimer = time.AfterFunc(rehideDebounce, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.rehideLocked()
	})
}

// SetDragging 通知后端当前是否处于拖拽中。
// 拖拽中不自动贴边/缩小，避免“需要重复多次移动才会移动到屏幕外/贴边行为打断拖拽”。
func (s *FloatingBallService) SetDragging(dragging bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.dragging
	s.dragging = dragging
	if s.win == nil || !s.visible {
		s.debugLog("SetDragging(no_window)", map[string]any{
			"prev": prev, "now": dragging, "visible": s.visible,
		})
		return
	}

	relX, relY := s.win.RelativePosition()
	b := s.win.Bounds()
	s.debugLog("SetDragging", map[string]any{
		"prev": prev, "now": dragging,
		"dock": s.dock, "collapsed": s.collapsed,
		"relX": relX, "relY": relY,
		"w": b.Width, "h": b.Height,
	})

	if dragging {
		// 记录拖拽起点，用于区分“点击”和“真实拖动”
		s.dragStartX, s.dragStartY = relX, relY
		s.dragMoved = false
		// 拖拽中取消自动贴边/缩小相关计时器
		if s.snapTimer != nil {
			s.snapTimer.Stop()
			s.snapTimer = nil
		}
		if s.idleDockTimer != nil {
			s.idleDockTimer.Stop()
			s.idleDockTimer = nil
		}
		return
	}

	// 拖拽结束：如果没有发生有效移动，不做 snap（避免“单击就贴边缩小”，影响双击）
	if !s.dragMoved {
		s.debugLog("drag_end_snap:skip_no_move", map[string]any{})
		return
	}

	// 拖拽结束：稍作延迟等待系统最终位置稳定，然后立刻判断贴边/对齐（不在这里缩小）
	time.AfterFunc(60*time.Millisecond, func() {
		s.dragEndSnap()
	})
}

func (s *FloatingBallService) dragEndSnap() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.win == nil || !s.visible {
		s.debugLog("drag_end_snap:skip", map[string]any{"visible": s.visible, "hasWin": s.win != nil})
		return
	}
	s.debugLog("drag_end_snap:run", map[string]any{})
	s.snapAfterMoveLocked()
}

// SetAppActive 通知后端应用是否处于激活状态（用于失焦时自动缩小贴边）
func (s *FloatingBallService) SetAppActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.appActive = active
	if s.win == nil || !s.visible {
		return
	}

	// 失去焦点：如果已经贴边，则立即缩小为把手，且关闭所有待执行的展开/回缩
	if !active {
		if s.rehideTimer != nil {
			s.rehideTimer.Stop()
			s.rehideTimer = nil
		}
		if s.idleDockTimer != nil {
			s.idleDockTimer.Stop()
			s.idleDockTimer = nil
		}
		if s.dock != DockNone {
			_, relY := s.win.RelativePosition()
			s.collapseToYLocked(relY)
		}
	}
}

// CloseFromUI 前端点击关闭按钮
func (s *FloatingBallService) CloseFromUI() {
	_ = s.SetVisible(false)
}

// OpenMainFromUI 前端双击悬浮球，唤起主窗口
func (s *FloatingBallService) OpenMainFromUI() {
	if s.mainWindow == nil {
		return
	}
	s.mainWindow.UnMinimise()
	s.mainWindow.Show()
	s.mainWindow.Focus()
}

func (s *FloatingBallService) ensureLocked() *application.WebviewWindow {
	if s.app == nil {
		return nil
	}
	if s.win != nil {
		return s.win
	}

	// 创建时就设置为屏幕最右侧贴边 + 垂直居中（避免首次显示跑到默认位置）
	x, y := s.defaultPositionLocked()

	w := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:          windowName,
		Title:         "WillChat",
		Width:         ballSize,
		Height:        ballSize,
		InitialPosition: application.WindowXY,
		X:               x,
		Y:               y,
		DisableResize: true,
		Frameless:     true,
		AlwaysOnTop:   true,
		Hidden:        true,
		URL:           "/floatingball.html",

		BackgroundType: application.BackgroundTypeTransparent,
		// 鼠标事件必须保留，否则无法交互
		IgnoreMouseEvents: false,

		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
		},
		Mac: application.MacWindow{
			Backdrop:     application.MacBackdropTransparent,
			DisableShadow: true,
			WindowLevel:  application.MacWindowLevelFloating,
			// 不依赖 titlebar drag，前端使用 --wails-draggable
			InvisibleTitleBarHeight: 0,
		},
		Linux: application.LinuxWindow{
			WindowIsTranslucent: true,
		},
	})

	// 监听移动事件（拖拽贴边隐藏）
	w.RegisterHook(events.Common.WindowDidMove, func(_ *application.WindowEvent) {
		s.onWindowDidMove()
	})
	// 显示后再次兜底定位（部分平台首次 SetPosition 可能被忽略）
	w.RegisterHook(events.Common.WindowShow, func(_ *application.WindowEvent) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.win == nil || !s.visible {
			return
		}
		s.scheduleRepositionLocked()
	})

	s.win = w
	return s.win
}

func (s *FloatingBallService) onWindowDidMove() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.win == nil {
		return
	}
	if !s.visible {
		return
	}
	if time.Now().Before(s.ignoreMoveUntil) {
		return
	}
	// 拖拽中不自动贴边/缩小
	if s.dragging {
		// 记录是否发生有效移动（阈值 2px）
		relX, relY := s.win.RelativePosition()
		if abs(relX-s.dragStartX) > 2 || abs(relY-s.dragStartY) > 2 {
			s.dragMoved = true
		}
		s.debugLog("WindowDidMove:skip_dragging", map[string]any{})
		return
	}

	if s.snapTimer != nil {
		s.snapTimer.Stop()
		s.snapTimer = nil
	}
	s.snapTimer = time.AfterFunc(snapDebounce, func() {
		s.snapAfterMove()
	})
}

func (s *FloatingBallService) snapAfterMove() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapAfterMoveLocked()
}

func (s *FloatingBallService) snapAfterMoveLocked() {
	if s.win == nil || !s.visible {
		return
	}

	// 使用 RelativePosition（相对 WorkArea），避免 macOS 绝对坐标系差异导致定位错误
	relX, relY := s.win.RelativePosition()
	bounds := s.win.Bounds()
	width := bounds.Width
	height := bounds.Height

	work, ok := s.workAreaLocked()
	if !ok {
		return
	}

	// Clamp Y into work area first (relative)
	y := clamp(relY, 0, work.Height-height)

	// Snap + collapse if near left/right edges (relative)
	if relX <= edgeSnapGap {
		s.dock = DockLeft
		s.debugLog("snap:DockLeft", map[string]any{"relX": relX, "edgeSnapGap": edgeSnapGap})
		// 仅贴边对齐（保持完整大小）；缩小交给失焦/鼠标移出/idle 逻辑
		s.expandToYLocked(y)
		s.scheduleIdleDockLocked()
		return
	}
	if relX+width >= work.Width-edgeSnapGap {
		s.dock = DockRight
		s.debugLog("snap:DockRight", map[string]any{"relX": relX, "width": width, "workW": work.Width, "edgeSnapGap": edgeSnapGap})
		// 仅贴边对齐（保持完整大小）；缩小交给失焦/鼠标移出/idle 逻辑
		s.expandToYLocked(y)
		s.scheduleIdleDockLocked()
		return
	}

	// Not docked: keep within work area and clear dock state
	s.dock = DockNone
	if s.collapsed {
		s.debugLog("snap:undock_expand", map[string]any{"relX": relX, "relY": relY})
		s.expandToYLocked(y)
		return
	}
	x := clamp(relX, 0, work.Width-width)
	s.debugLog("snap:none", map[string]any{"x": x, "y": y, "relX": relX, "relY": relY})
	s.setRelativePositionLocked(x, y)

	// 移动结束后，若鼠标未 hover，超过一段时间自动贴边缩小
	s.scheduleIdleDockLocked()
}

func (s *FloatingBallService) resetToDefaultPositionLocked() {
	if s.win == nil || s.app == nil {
		return
	}

	x, y := s.defaultPositionLocked()
	s.dock = DockNone
	s.collapsed = false
	s.setSizeLocked(ballSize, ballSize)
	s.setRelativePositionLocked(x, y)
}

func (s *FloatingBallService) defaultPositionLocked() (int, int) {
	work, ok := s.workAreaLocked()
	if !ok {
		return 0, 0
	}
	// relative to WorkArea (0,0)
	x := work.Width - ballSize - defaultMargin // 贴右边（默认无边距）
	y := (work.Height - ballSize) / 2
	return x, y
}

func (s *FloatingBallService) scheduleRepositionLocked() {
	if s.win == nil || !s.visible {
		return
	}
	// cancel previous
	if s.repositionTimer != nil {
		s.repositionTimer.Stop()
		s.repositionTimer = nil
	}
	s.repositionTries = 0
	s.repositionTimer = time.AfterFunc(postShowRepositionDelay, func() {
		s.repositionTick()
	})
}

func (s *FloatingBallService) repositionTick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.win == nil || !s.visible {
		return
	}
	s.repositionTries++

	// impl 就绪的一个可靠信号：GetScreen() 返回非 nil
	screen, _ := s.win.GetScreen()
	if screen != nil {
		s.restoreOrDefaultLocked()
		return
	}
	if s.repositionTries >= postShowRepositionTries {
		// 最后兜底：即使拿不到 screen，也尝试设置一次位置
		s.restoreOrDefaultLocked()
		return
	}

	// retry
	s.repositionTimer = time.AfterFunc(postShowRepositionDelay, func() {
		s.repositionTick()
	})
}

func (s *FloatingBallService) restoreOrDefaultLocked() {
	if s.win == nil {
		return
	}
	// If we have a last known state, restore it; otherwise use default.
	if s.hasLastState {
		s.debugLog("restore:last_state", map[string]any{
			"x": s.lastRelX, "y": s.lastRelY, "dock": s.lastDock, "collapsed": s.lastCollapsed,
		})
		s.dock = s.lastDock
		s.collapsed = s.lastCollapsed
		if s.collapsed {
			s.setSizeLocked(collapsedWidth, ballSize)
		} else {
			s.setSizeLocked(ballSize, ballSize)
		}
		s.setRelativePositionLocked(s.lastRelX, s.lastRelY)
		return
	}
	s.resetToDefaultPositionLocked()
}

func (s *FloatingBallService) expandLocked() {
	if s.win == nil || s.dock == DockNone {
		return
	}

	work, ok := s.workAreaLocked()
	if !ok {
		return
	}
	_, relY := s.win.RelativePosition()
	bounds := s.win.Bounds()
	y := clamp(relY, 0, work.Height-bounds.Height)

	s.expandToYLocked(y)
}

func (s *FloatingBallService) rehideLocked() {
	if s.win == nil || s.dock == DockNone {
		return
	}

	work, ok := s.workAreaLocked()
	if !ok {
		return
	}
	_, relY := s.win.RelativePosition()
	bounds := s.win.Bounds()
	y := clamp(relY, 0, work.Height-bounds.Height)

	s.collapseToYLocked(y)
}

func (s *FloatingBallService) scheduleIdleDockLocked() {
	if s.win == nil || !s.visible {
		return
	}
	// 未 hover 时生效（无论是否已贴边），用于“停留一段时间后自动缩小”
	if s.hovered {
		return
	}
	if s.collapsed {
		return
	}

	if s.idleDockTimer != nil {
		s.idleDockTimer.Stop()
		s.idleDockTimer = nil
	}
	s.idleDockTimer = time.AfterFunc(idleDockDelay, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.win == nil || !s.visible {
			return
		}
		if s.hovered || s.collapsed {
			return
		}
		if !s.win.IsVisible() {
			return
		}

		// 自动缩小：若已贴边则直接缩小；若未贴边则仅在靠近边缘时贴边并缩小
		work, ok := s.workAreaLocked()
		if !ok {
			return
		}
		relX, relY := s.win.RelativePosition()
		b := s.win.Bounds()
		width := b.Width
		height := b.Height
		y := clamp(relY, 0, work.Height-height)

		if s.dock == DockLeft || s.dock == DockRight {
			s.rehideLocked()
			return
		}
		// decide side by proximity
		if relX <= edgeSnapGap {
			s.dock = DockLeft
			s.collapseToYLocked(y)
			return
		}
		if relX+width >= work.Width-edgeSnapGap {
			s.dock = DockRight
			s.collapseToYLocked(y)
			return
		}
	})
}

func (s *FloatingBallService) stopTimersLocked() {
	if s.snapTimer != nil {
		s.snapTimer.Stop()
		s.snapTimer = nil
	}
	if s.rehideTimer != nil {
		s.rehideTimer.Stop()
		s.rehideTimer = nil
	}
	if s.idleDockTimer != nil {
		s.idleDockTimer.Stop()
		s.idleDockTimer = nil
	}
	if s.repositionTimer != nil {
		s.repositionTimer.Stop()
		s.repositionTimer = nil
	}
}

func (s *FloatingBallService) setPositionLocked(x, y int) {
	if s.win == nil {
		return
	}
	s.ignoreMoveUntil = time.Now().Add(250 * time.Millisecond)
	s.win.SetPosition(x, y)
}

func (s *FloatingBallService) setRelativePositionLocked(x, y int) {
	if s.win == nil {
		return
	}
	s.ignoreMoveUntil = time.Now().Add(250 * time.Millisecond)
	s.win.SetRelativePosition(x, y)
}

func (s *FloatingBallService) setSizeLocked(width, height int) {
	if s.win == nil {
		return
	}
	s.ignoreMoveUntil = time.Now().Add(250 * time.Millisecond)
	s.win.SetSize(width, height)
}

func (s *FloatingBallService) expandToYLocked(y int) {
	if s.win == nil {
		return
	}
	work, ok := s.workAreaLocked()
	if !ok {
		return
	}
	s.collapsed = false
	s.setSizeLocked(ballSize, ballSize)

	y = clamp(y, 0, work.Height-ballSize)
	x := 0
	switch s.dock {
	case DockLeft:
		x = 0
	case DockRight:
		x = work.Width - ballSize
	}
	s.setRelativePositionLocked(x, y)
}

func (s *FloatingBallService) collapseToYLocked(y int) {
	if s.win == nil {
		return
	}
	work, ok := s.workAreaLocked()
	if !ok {
		return
	}
	s.collapsed = true
	s.setSizeLocked(collapsedWidth, ballSize)

	y = clamp(y, 0, work.Height-ballSize)
	x := 0
	switch s.dock {
	case DockLeft:
		x = -(collapsedWidth - collapsedVisible)
	case DockRight:
		x = work.Width - collapsedVisible
	}
	s.debugLog("collapse", map[string]any{"dock": s.dock, "x": x, "y": y, "w": collapsedWidth, "h": ballSize})
	s.setRelativePositionLocked(x, y)
}

func clamp(v, min, max int) int {
	if max < min {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (s *FloatingBallService) workAreaLocked() (application.Rect, bool) {
	// 优先以窗口所在屏幕为准（mac 上更可靠；也支持多显示器）
	if s.win != nil {
		if screen, _ := s.win.GetScreen(); screen != nil {
			if screen.WorkArea.Width > 0 && screen.WorkArea.Height > 0 {
				return screen.WorkArea, true
			}
		}
	}
	if s.app == nil || s.app.Screen == nil {
		return application.Rect{}, false
	}
	if screen := s.app.Screen.GetPrimary(); screen != nil {
		if screen.WorkArea.Width > 0 && screen.WorkArea.Height > 0 {
			return screen.WorkArea, true
		}
	}
	return application.Rect{}, false
}

