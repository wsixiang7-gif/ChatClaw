package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			// 创建供应商表
			createProviders := `
create table if not exists providers (
    id integer primary key autoincrement,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    
    provider_id varchar(64) not null unique,
    name varchar(64) not null,
    type varchar(16) not null default 'openai',
    icon text not null default '',
    is_builtin boolean not null default false,
    enabled boolean not null default false,
    sort_order integer not null default 0,
    
    api_endpoint varchar(1024) not null default '',
    api_key varchar(1024) not null default '',
    extra_config text not null default '{}'
);
`
			if _, err := db.ExecContext(ctx, createProviders); err != nil {
				return err
			}

			// 创建模型表
			createModels := `
create table if not exists models (
    id integer primary key autoincrement,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    
    provider_id varchar(64) not null,
    model_id varchar(128) not null,
    name varchar(128) not null,
    type varchar(16) not null default 'llm',
    is_builtin boolean not null default false,
    enabled boolean not null default true,
    sort_order integer not null default 0,
    
    unique(provider_id, model_id)
);
`
			if _, err := db.ExecContext(ctx, createModels); err != nil {
				return err
			}

			// 初始化内置供应商
			insertProviders := `
insert into providers (provider_id, name, type, icon, is_builtin, sort_order, api_endpoint) values
('openai', 'OpenAI', 'openai', 'openai', true, 1, 'https://api.openai.com/v1'),
('azure', 'Azure OpenAI', 'azure', 'azure', true, 2, ''),
('anthropic', 'Anthropic', 'anthropic', 'anthropic', true, 3, 'https://api.anthropic.com/v1'),
('google', 'Google Gemini', 'gemini', 'google', true, 4, 'https://generativelanguage.googleapis.com/v1beta'),
('grok', 'Grok', 'openai', 'grok', true, 5, 'https://api.x.ai/v1'),
('deepseek', 'DeepSeek', 'openai', 'deepseek', true, 6, 'https://api.deepseek.com/v1'),
('zhipu', '智谱 GLM', 'openai', 'zhipu', true, 7, 'https://open.bigmodel.cn/api/paas/v4'),
('qwen', '通义千问', 'openai', 'qwen', true, 8, 'https://dashscope.aliyuncs.com/compatible-mode/v1'),
('doubao', '豆包', 'openai', 'doubao', true, 9, 'https://ark.cn-beijing.volces.com/api/v3'),
('baidu', '百度文心', 'openai', 'baidu', true, 10, 'https://qianfan.baidubce.com/v2'),
('ollama', 'Ollama', 'openai', 'ollama', true, 11, 'http://localhost:11434/v1');
`
			if _, err := db.ExecContext(ctx, insertProviders); err != nil {
				return err
			}

			// 初始化内置模型
			insertModels := `
insert into models (provider_id, model_id, name, type, is_builtin, sort_order) values
('openai', 'gpt-5.2', 'GPT-5.2', 'llm', true, 100),
('openai', 'gpt-5.1', 'GPT-5.1', 'llm', true, 102),
('openai', 'gpt-5', 'GPT-5', 'llm', true, 102),
('openai', 'gpt-5 mini', 'GPT-5 mini, 'llm', true, 103),
('openai', 'gpt-5.2 nano', 'GPT-5.2 nano', 'llm', true, 104),
('openai', 'gpt-5.2 pro', 'GPT-5.2 pro', 'llm', true, 105),

('openai', 'text-embedding-3-large', 'Text Embedding 3 Large', 'embedding', true, 100),
('openai', 'text-embedding-3-small', 'Text Embedding 3 Small', 'embedding', true, 101),

('anthropic', 'claude-sonnet-4-5-20250929', 'Claude Sonnet 4.5', 'llm', true, 100),
('anthropic', 'claude-haiku-4-5-20251001', 'Claude Haiku 4.5', 'llm', true, 101),
('anthropic', 'claude-opus-4-5-20251101', 'Claude Opus 4.5', 'llm', true, 102),

('google', 'gemini-3-pro-preview', 'Gemini 3 Pro', 'llm', true, 100),
('google', 'gemini-3-flash-preview', 'Gemini 3 Flash', 'llm', true, 101),
('google', 'gemini-2.5-flash', 'Gemini 2.5 Flash', 'llm', true, 102),
('google', 'gemini-2.5-flash-lite', 'Gemini 2.5 Flash-Lite', 'llm', true, 103),
('google', 'gemini-2.5-pro', 'Gemini 2.5 Pro', 'llm', true, 104),

('deepseek', 'deepseek-chat', 'DeepSeek V3', 'llm', true, 100),
('deepseek', 'deepseek-reasoner', 'DeepSeek R1', 'llm', true, 101),

('zhipu', 'glm-4.7', 'glm-4.7', 'llm', true, 100),
('zhipu', 'glm-4.7-flash', 'glm-4.7-flash', 'llm', true, 101),
('zhipu', 'glm-4.7-flashx', 'glm-4.7-flashx', 'llm', true, 102),
('zhipu', 'glm-4.6', 'glm-4.6', 'llm', true, 103),
('zhipu', 'glm-4.5-air', 'glm-4.5-air', 'llm', true, 104),
('zhipu', 'glm-4.5-airx', 'glm-4.5-airx', 'llm', true, 105),
('zhipu', 'glm-4.5-flash', 'glm-4.5-flash', 'llm', true, 106),
('zhipu', 'glm-4-flash-250414', 'glm-4-flash-250414', 'llm', true, 107),
('zhipu', 'glm-4-flashx-250414', 'glm-4-flashx-250414', 'llm', true, 108),
('zhipu', 'embedding-3', 'Embedding-3', 'embedding', true, 100),

('qwen', 'qwen3-max', '通义千问 Max', 'llm', true, 100),
('qwen', 'qwen-plus', '通义千问 Plus', 'llm', true, 101),
('qwen', 'qwen-flash', '通义千问 Flash', 'llm', true, 102),
('qwen', 'qwen-long', '通义千问 Long', 'llm', true, 103),
('qwen', 'text-embedding-v3', 'Text Embedding V3', 'embedding', true, 100),

('baidu', 'ernie-5.0-thinking-latest', 'ERNIE 5.0', 'llm', true, 100),
('baidu', 'ernie-4.5-turbo-latest', 'ERNIE 4.5 Turbo', 'llm', true, 101),
('baidu', 'ernie-speed-pro-128k', 'ERNIE Speed', 'llm', true, 102),
('baidu', 'ernie-lite-pro-128k', 'ERNIE Lite', 'llm', true, 102),

('grok', 'grok-4-1-fast-reasoning', 'Grok 4.1 Fast Reasoning', 'llm', true, 100),
('grok', 'grok-4-1-fast-reasoning-pro', 'Grok 4.1 Fast Reasoning Pro', 'llm', true, 101),
('grok', 'grok-4-fast-reasoning', 'Grok 4 Fast Reasoning', 'llm', true, 102);
('grok', 'grok-4-fast-non-reasoning', 'Grok 4 Fast Non-Reasoning', 'llm', true, 103);
`
			if _, err := db.ExecContext(ctx, insertModels); err != nil {
				return err
			}

			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			if _, err := db.ExecContext(ctx, `drop table if exists models`); err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx, `drop table if exists providers`); err != nil {
				return err
			}
			return nil
		},
	)
}
