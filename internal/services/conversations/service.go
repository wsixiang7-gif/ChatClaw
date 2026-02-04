package conversations

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"willchat/internal/errs"
	"willchat/internal/sqlite"

	"github.com/uptrace/bun"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ConversationsService 会话服务（暴露给前端调用）
type ConversationsService struct {
	app *application.App
}

func NewConversationsService(app *application.App) *ConversationsService {
	return &ConversationsService{app: app}
}

func (s *ConversationsService) db() (*bun.DB, error) {
	db := sqlite.DB()
	if db == nil {
		return nil, errs.New("error.sqlite_not_initialized")
	}
	return db, nil
}

// ListConversations 获取指定助手的会话列表（排除已软删除的）
func (s *ConversationsService) ListConversations(agentID int64) ([]Conversation, error) {
	if agentID <= 0 {
		return nil, errs.New("error.agent_id_required")
	}

	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	models := make([]conversationModel, 0)
	if err := db.NewSelect().
		Model(&models).
		Where("agent_id = ?", agentID).
		Where("is_deleted = ?", false).
		OrderExpr("updated_at DESC, id DESC").
		Scan(ctx); err != nil {
		return nil, errs.Wrap("error.conversation_list_failed", err)
	}

	out := make([]Conversation, 0, len(models))
	for i := range models {
		out = append(out, models[i].toDTO())
	}
	return out, nil
}

// GetConversation 获取单个会话
func (s *ConversationsService) GetConversation(id int64) (*Conversation, error) {
	if id <= 0 {
		return nil, errs.New("error.conversation_id_required")
	}

	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var m conversationModel
	if err := db.NewSelect().
		Model(&m).
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.Newf("error.conversation_not_found", map[string]any{"ID": id})
		}
		return nil, errs.Wrap("error.conversation_read_failed", err)
	}

	dto := m.toDTO()
	return &dto, nil
}

// CreateConversation 创建会话
func (s *ConversationsService) CreateConversation(input CreateConversationInput) (*Conversation, error) {
	if input.AgentID <= 0 {
		return nil, errs.New("error.agent_id_required")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errs.New("error.conversation_name_required")
	}
	// 截取前 100 个字符作为会话名称
	nameRunes := []rune(name)
	if len(nameRunes) > 100 {
		name = string(nameRunes[:100])
	}

	lastMessage := strings.TrimSpace(input.LastMessage)

	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 验证助手是否存在
	var agentCount int
	if err := db.NewSelect().
		Table("agents").
		ColumnExpr("COUNT(1)").
		Where("id = ?", input.AgentID).
		Scan(ctx, &agentCount); err != nil {
		return nil, errs.Wrap("error.conversation_create_failed", err)
	}
	if agentCount == 0 {
		return nil, errs.Newf("error.agent_not_found", map[string]any{"ID": input.AgentID})
	}

	m := &conversationModel{
		AgentID:     input.AgentID,
		Name:        name,
		LastMessage: lastMessage,
		IsDeleted:   false,
	}

	if _, err := db.NewInsert().Model(m).Exec(ctx); err != nil {
		return nil, errs.Wrap("error.conversation_create_failed", err)
	}

	dto := m.toDTO()
	return &dto, nil
}

// UpdateConversation 更新会话（重命名或更新最后一条消息）
func (s *ConversationsService) UpdateConversation(id int64, input UpdateConversationInput) (*Conversation, error) {
	if id <= 0 {
		return nil, errs.New("error.conversation_id_required")
	}

	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := db.NewUpdate().
		Model((*conversationModel)(nil)).
		Where("id = ?", id).
		Where("is_deleted = ?", false)

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, errs.New("error.conversation_name_required")
		}
		// 截取前 100 个字符
		nameRunes := []rune(name)
		if len(nameRunes) > 100 {
			name = string(nameRunes[:100])
		}
		q = q.Set("name = ?", name)
	}

	if input.LastMessage != nil {
		q = q.Set("last_message = ?", strings.TrimSpace(*input.LastMessage))
	}

	result, err := q.Exec(ctx)
	if err != nil {
		return nil, errs.Wrap("error.conversation_update_failed", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errs.Newf("error.conversation_not_found", map[string]any{"ID": id})
	}

	return s.GetConversation(id)
}

// DeleteConversation 软删除会话
func (s *ConversationsService) DeleteConversation(id int64) error {
	if id <= 0 {
		return errs.New("error.conversation_id_required")
	}

	db, err := s.db()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := db.NewUpdate().
		Model((*conversationModel)(nil)).
		Where("id = ?", id).
		Where("is_deleted = ?", false).
		Set("is_deleted = ?", true).
		Set("updated_at = ?", sqlite.NowUTC()).
		Exec(ctx)
	if err != nil {
		return errs.Wrap("error.conversation_delete_failed", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Newf("error.conversation_not_found", map[string]any{"ID": id})
	}
	return nil
}

// DeleteConversationsByAgentID 删除指定助手的所有会话（用于删除助手时清理）
func (s *ConversationsService) DeleteConversationsByAgentID(agentID int64) error {
	if agentID <= 0 {
		return errs.New("error.agent_id_required")
	}

	db, err := s.db()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 软删除所有该助手的会话
	_, err = db.NewUpdate().
		Model((*conversationModel)(nil)).
		Where("agent_id = ?", agentID).
		Where("is_deleted = ?", false).
		Set("is_deleted = ?", true).
		Set("updated_at = ?", sqlite.NowUTC()).
		Exec(ctx)
	if err != nil {
		return errs.Wrap("error.conversation_delete_failed", err)
	}

	return nil
}
