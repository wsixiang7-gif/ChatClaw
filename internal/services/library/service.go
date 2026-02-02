package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"willchat/internal/errs"
	"willchat/internal/services/settings"
	"willchat/internal/sqlite"

	"github.com/uptrace/bun"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// LibraryService 知识库服务（暴露给前端调用）
type LibraryService struct {
	app *application.App
}

func NewLibraryService(app *application.App) *LibraryService {
	return &LibraryService{app: app}
}

func (s *LibraryService) db() (*bun.DB, error) {
	db := sqlite.DB()
	if db == nil {
		return nil, errs.New("error.sqlite_not_initialized")
	}
	return db, nil
}

// ListLibraries 获取知识库列表（个人知识库）
func (s *LibraryService) ListLibraries() ([]Library, error) {
	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	models := make([]libraryModel, 0)
	if err := db.NewSelect().
		Model(&models).
		OrderExpr("sort_order DESC, id DESC").
		Scan(ctx); err != nil {
		return nil, errs.Wrap("error.library_list_failed", err)
	}

	out := make([]Library, 0, len(models))
	for i := range models {
		out = append(out, models[i].toDTO())
	}
	return out, nil
}

// CreateLibrary 创建知识库
func (s *LibraryService) CreateLibrary(input CreateLibraryInput) (*Library, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errs.New("error.library_name_required")
	}
	if len([]rune(name)) > 128 {
		return nil, errs.New("error.library_name_too_long")
	}

	// 全局嵌入配置（来自 settings 缓存）
	embeddingProviderID, ok := settings.GetValue("embedding_provider_id")
	if !ok || strings.TrimSpace(embeddingProviderID) == "" {
		return nil, errs.New("error.library_embedding_global_not_set")
	}
	embeddingModelID, ok := settings.GetValue("embedding_model_id")
	if !ok || strings.TrimSpace(embeddingModelID) == "" {
		return nil, errs.New("error.library_embedding_global_not_set")
	}
	embeddingProviderID = strings.TrimSpace(embeddingProviderID)
	embeddingModelID = strings.TrimSpace(embeddingModelID)

	embeddingDimension := 1536
	if dimStr, ok := settings.GetValue("embedding_dimension"); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(dimStr)); err == nil && n > 0 {
			embeddingDimension = n
		}
	}

	rerankProviderID := strings.TrimSpace(input.RerankProviderID)
	rerankModelID := strings.TrimSpace(input.RerankModelID)

	// 默认值（与 migrations 中的 DEFAULT 保持一致）
	topK := 10
	chunkSize := 1024
	chunkOverlap := 100
	matchThreshold := 0.5

	if input.TopK != nil && *input.TopK > 0 {
		topK = *input.TopK
	}
	if input.ChunkSize != nil && *input.ChunkSize > 0 {
		chunkSize = *input.ChunkSize
	}
	if input.ChunkOverlap != nil && *input.ChunkOverlap >= 0 {
		chunkOverlap = *input.ChunkOverlap
	}
	if input.MatchThreshold != nil && *input.MatchThreshold >= 0 && *input.MatchThreshold <= 1 {
		matchThreshold = *input.MatchThreshold
	}

	db, err := s.db()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 如果前端未指定重排模型，则选择一个默认值（按 models 表的 sort_order/id）
	if rerankProviderID == "" || rerankModelID == "" {
		type row struct {
			ProviderID string `bun:"provider_id"`
			ModelID    string `bun:"model_id"`
		}
		var r row
		err := db.NewSelect().
			Table("models").
			Column("provider_id", "model_id").
			Where("type = ?", "rerank").
			OrderExpr("sort_order ASC, id ASC").
			Limit(1).
			Scan(ctx, &r)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errs.New("error.library_rerank_required")
			}
			return nil, errs.Wrap("error.library_create_failed", err)
		}
		rerankProviderID = r.ProviderID
		rerankModelID = r.ModelID
	}

	// sort_order 自动 +1（越新越大）
	var maxSort sql.NullInt64
	if err := db.NewSelect().
		Table("library").
		ColumnExpr("MAX(sort_order)").
		Scan(ctx, &maxSort); err != nil {
		return nil, errs.Wrap("error.library_create_failed", err)
	}
	sortOrder := 1
	if maxSort.Valid {
		sortOrder = int(maxSort.Int64) + 1
	}

	m := &libraryModel{
		Name: name,

		EmbeddingProviderID: embeddingProviderID,
		EmbeddingModelID:    embeddingModelID,
		EmbeddingDimension:  embeddingDimension,

		RerankProviderID: rerankProviderID,
		RerankModelID:    rerankModelID,

		TopK:           topK,
		ChunkSize:      chunkSize,
		ChunkOverlap:   chunkOverlap,
		MatchThreshold: matchThreshold,
		SortOrder:      sortOrder,
	}

	if _, err := db.NewInsert().Model(m).Exec(ctx); err != nil {
		return nil, errs.Wrap("error.library_create_failed", fmt.Errorf("insert: %w", err))
	}

	dto := m.toDTO()
	return &dto, nil
}
