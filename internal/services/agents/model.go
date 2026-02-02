package agents

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Agent 助手 DTO（暴露给前端）
type Agent struct {
	ID int64 `json:"id"`

	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	Icon   string `json:"icon"`

	DefaultLLMProviderID string  `json:"default_llm_provider_id"`
	DefaultLLMModelID    string  `json:"default_llm_model_id"`
	LLMTemperature       float64 `json:"llm_temperature"`
	LLMTopP              float64 `json:"llm_top_p"`
	ContextCount         int     `json:"context_count"`
	LLMMaxTokens         int     `json:"llm_max_tokens"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAgentInput struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	Icon   string `json:"icon"`
}

type UpdateAgentInput struct {
	Name   *string `json:"name"`
	Prompt *string `json:"prompt"`
	Icon   *string `json:"icon"`

	DefaultLLMProviderID *string `json:"default_llm_provider_id"`
	DefaultLLMModelID    *string `json:"default_llm_model_id"`

	LLMTemperature *float64 `json:"llm_temperature"`
	LLMTopP        *float64 `json:"llm_top_p"`
	ContextCount   *int     `json:"context_count"`
	LLMMaxTokens   *int     `json:"llm_max_tokens"`
}

type agentModel struct {
	bun.BaseModel `bun:"table:agents,alias:a"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`

	Name   string `bun:"name,notnull"`
	Prompt string `bun:"prompt,notnull"`
	Icon   string `bun:"icon,notnull"`

	DefaultLLMProviderID string  `bun:"default_llm_provider_id,notnull"`
	DefaultLLMModelID    string  `bun:"default_llm_model_id,notnull"`
	LLMTemperature       float64 `bun:"llm_temperature,notnull"`
	LLMTopP              float64 `bun:"llm_top_p,notnull"`
	ContextCount         int     `bun:"context_count,notnull"`
	LLMMaxTokens         int     `bun:"llm_max_tokens,notnull"`
}

func (m *agentModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
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

func (m *agentModel) toDTO() Agent {
	return Agent{
		ID: m.ID,

		Name:   m.Name,
		Prompt: m.Prompt,
		Icon:   m.Icon,

		DefaultLLMProviderID: m.DefaultLLMProviderID,
		DefaultLLMModelID:    m.DefaultLLMModelID,
		LLMTemperature:       m.LLMTemperature,
		LLMTopP:              m.LLMTopP,
		ContextCount:         m.ContextCount,
		LLMMaxTokens:         m.LLMMaxTokens,

		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
