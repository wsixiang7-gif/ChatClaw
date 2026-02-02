package library

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Library 知识库 DTO（暴露给前端）
type Library struct {
	ID int64 `json:"id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name string `json:"name"`

	RerankProviderID string `json:"rerank_provider_id"`
	RerankModelID    string `json:"rerank_model_id"`

	TopK           int     `json:"top_k"`
	ChunkSize      int     `json:"chunk_size"`
	ChunkOverlap   int     `json:"chunk_overlap"`
	MatchThreshold float64 `json:"match_threshold"`
	SortOrder      int     `json:"sort_order"`
}

// CreateLibraryInput 创建知识库的输入参数
// 说明：带默认值的字段前端可不填（后端会用默认值兜底）。
type CreateLibraryInput struct {
	Name string `json:"name"`

	RerankProviderID string `json:"rerank_provider_id"`
	RerankModelID    string `json:"rerank_model_id"`

	TopK           *int     `json:"top_k"`
	ChunkSize      *int     `json:"chunk_size"`
	ChunkOverlap   *int     `json:"chunk_overlap"`
	MatchThreshold *float64 `json:"match_threshold"`
}

// UpdateLibraryInput 更新知识库的输入参数
type UpdateLibraryInput struct {
	Name *string `json:"name"`

	RerankProviderID *string `json:"rerank_provider_id"`
	RerankModelID    *string `json:"rerank_model_id"`

	TopK           *int     `json:"top_k"`
	ChunkSize      *int     `json:"chunk_size"`
	ChunkOverlap   *int     `json:"chunk_overlap"`
	MatchThreshold *float64 `json:"match_threshold"`
}

// libraryModel 数据库模型
type libraryModel struct {
	bun.BaseModel `bun:"table:library,alias:l"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`

	Name string `bun:"name,notnull"`

	RerankProviderID string `bun:"rerank_provider_id,notnull"`
	RerankModelID    string `bun:"rerank_model_id,notnull"`

	TopK           int     `bun:"top_k,notnull"`
	ChunkSize      int     `bun:"chunk_size,notnull"`
	ChunkOverlap   int     `bun:"chunk_overlap,notnull"`
	MatchThreshold float64 `bun:"match_threshold,notnull"`
	SortOrder      int     `bun:"sort_order,notnull"`
}

func (m *libraryModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	_ = ctx
	now := time.Now().UTC()
	switch query.(type) {
	case *bun.InsertQuery:
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now
		}
		m.UpdatedAt = now
	case *bun.UpdateQuery:
		m.UpdatedAt = now
	}
	return nil
}

func (m *libraryModel) toDTO() Library {
	return Library{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,

		Name: m.Name,

		RerankProviderID: m.RerankProviderID,
		RerankModelID:    m.RerankModelID,

		TopK:           m.TopK,
		ChunkSize:      m.ChunkSize,
		ChunkOverlap:   m.ChunkOverlap,
		MatchThreshold: m.MatchThreshold,
		SortOrder:      m.SortOrder,
	}
}
