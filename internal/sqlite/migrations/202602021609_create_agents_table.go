package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			sql := `
create table if not exists agents (
    id integer primary key autoincrement,
	created_at datetime not null,
    updated_at datetime not null,
	
    name varchar(100) not null,
    description varchar(1000) not null,
	icon text not null
);			
`
			if _, err := db.ExecContext(ctx, sql); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, db *bun.DB) error {
			_ = ctx
			_ = db
			return nil
		},
	)
}
