package taskmanager

import (
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TaskManager 全局任务管理器，基于 ants 封装
// 提供任务防重复、状态跟踪等能力
type TaskManager struct {
	app    *application.App
	pool   *ants.Pool
	mu     sync.RWMutex
	tasks  map[string]*TaskInfo // taskKey -> TaskInfo
	closed bool
}

// TaskInfo 任务信息
type TaskInfo struct {
	Key       string // 任务唯一标识
	RunID     string // 当前运行 ID（用于防止旧任务回写）
	Cancelled bool   // 是否已取消
}

var (
	once     sync.Once
	instance *TaskManager
)

// Init 初始化全局任务管理器
// maxWorkers: 最大并发工作协程数
func Init(app *application.App, maxWorkers int) error {
	var initErr error
	once.Do(func() {
		pool, err := ants.NewPool(maxWorkers,
			ants.WithPreAlloc(false),        // 不预分配
			ants.WithNonblocking(false),     // 阻塞模式，任务满时等待
			ants.WithPanicHandler(func(i interface{}) {
				// panic 处理，防止程序崩溃
				if app != nil {
					app.Logger.Error("task panic", "error", i)
				}
			}),
		)
		if err != nil {
			initErr = err
			return
		}
		instance = &TaskManager{
			app:   app,
			pool:  pool,
			tasks: make(map[string]*TaskInfo),
		}
	})
	return initErr
}

// Get 获取全局任务管理器实例
func Get() *TaskManager {
	return instance
}

// Submit 提交任务
// taskKey: 任务唯一标识（同一 key 的任务不会重复执行）
// runID: 运行 ID（用于校验任务是否被取消/替换）
// fn: 任务函数，接收 TaskInfo 用于检查任务状态
func (tm *TaskManager) Submit(taskKey, runID string, fn func(info *TaskInfo)) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return false
	}

	// 检查是否已有同 key 的任务在运行
	if existing, ok := tm.tasks[taskKey]; ok {
		// 标记旧任务为取消
		existing.Cancelled = true
	}

	// 创建新任务信息
	info := &TaskInfo{
		Key:       taskKey,
		RunID:     runID,
		Cancelled: false,
	}
	tm.tasks[taskKey] = info

	// 提交到 ants 协程池
	err := tm.pool.Submit(func() {
		defer tm.removeTask(taskKey)
		fn(info)
	})

	if err != nil {
		delete(tm.tasks, taskKey)
		return false
	}

	return true
}

// Cancel 取消指定任务
func (tm *TaskManager) Cancel(taskKey string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if info, ok := tm.tasks[taskKey]; ok {
		info.Cancelled = true
	}
}

// IsTaskRunning 检查任务是否正在运行
func (tm *TaskManager) IsTaskRunning(taskKey string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, ok := tm.tasks[taskKey]
	return ok
}

// GetTaskInfo 获取任务信息
func (tm *TaskManager) GetTaskInfo(taskKey string) *TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.tasks[taskKey]
}

// IsCancelled 检查任务是否已被取消
func (info *TaskInfo) IsCancelled() bool {
	return info == nil || info.Cancelled
}

// Emit 发送事件到前端
func (tm *TaskManager) Emit(eventName string, data any) {
	if tm.app != nil {
		tm.app.Event.Emit(eventName, data)
	}
}

// removeTask 移除任务记录
func (tm *TaskManager) removeTask(taskKey string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tasks, taskKey)
}

// Running 返回当前正在运行的协程数
func (tm *TaskManager) Running() int {
	return tm.pool.Running()
}

// Cap 返回协程池容量
func (tm *TaskManager) Cap() int {
	return tm.pool.Cap()
}

// Free 返回空闲协程数
func (tm *TaskManager) Free() int {
	return tm.pool.Free()
}

// Tune 动态调整协程池容量
func (tm *TaskManager) Tune(size int) {
	tm.pool.Tune(size)
}

// Stop 停止任务管理器，等待所有任务完成
func (tm *TaskManager) Stop() {
	tm.mu.Lock()
	tm.closed = true
	tm.mu.Unlock()

	tm.pool.Release()
}

// StopNow 立即停止任务管理器，不等待未完成的任务
func (tm *TaskManager) StopNow() {
	tm.mu.Lock()
	tm.closed = true
	// 取消所有任务
	for _, info := range tm.tasks {
		info.Cancelled = true
	}
	tm.mu.Unlock()

	tm.pool.Release()
}
