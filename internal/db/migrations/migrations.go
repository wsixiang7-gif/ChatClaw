package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

// Migrations 是全局迁移注册表：
// - SQL migrations 通过 Discover 从 embed FS 加载
// - Go migrations 在同包内的 init() 里调用 Migrations.MustRegister 注册
var Migrations = migrate.NewMigrations()

// SQL + Go 迁移统一放在同一目录下：
// - *.up.sql / *.down.sql / *.tx.up.sql / *.tx.down.sql（SQL migrations）
// - 其它 .go 文件里的 init() 会向 Migrations 注册 Go migrations
//
//go:embed *.sql
var migrationsFS embed.FS

func init() {
	if err := Migrations.Discover(migrationsFS); err != nil {
		panic(err)
	}
}
