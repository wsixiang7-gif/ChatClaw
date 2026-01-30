package migrations

import (
	"context"
	"fmt"
	"strings"

	"willchat/internal/define"

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

			// 初始化内置供应商（使用共享配置）
			if len(define.BuiltinProviders) > 0 {
				var values []string
				for _, p := range define.BuiltinProviders {
					values = append(values, fmt.Sprintf(
						"('%s', '%s', '%s', '%s', true, %d, '%s')",
						p.ProviderID, p.Name, p.Type, p.Icon, p.SortOrder, p.APIEndpoint,
					))
				}
				insertProviders := `insert into providers (provider_id, name, type, icon, is_builtin, sort_order, api_endpoint) values ` + strings.Join(values, ",\n")
				if _, err := db.ExecContext(ctx, insertProviders); err != nil {
					return err
				}
			}

			// 初始化内置模型（使用共享配置）
			if len(define.BuiltinModels) > 0 {
				var values []string
				for _, m := range define.BuiltinModels {
					values = append(values, fmt.Sprintf(
						"('%s', '%s', '%s', '%s', true, %d)",
						m.ProviderID, m.ModelID, m.Name, m.Type, m.SortOrder,
					))
				}
				insertModels := `insert into models (provider_id, model_id, name, type, is_builtin, sort_order) values ` + strings.Join(values, ",\n")
				if _, err := db.ExecContext(ctx, insertModels); err != nil {
					return err
				}
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
