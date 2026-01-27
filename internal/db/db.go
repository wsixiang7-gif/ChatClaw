package db

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"willchat/internal/db/migrations"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/migrate"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	mu     sync.Mutex
	sqlDB  *sql.DB
	bunDB  *bun.DB
	dbPath string
)

func Path() string {
	mu.Lock()
	defer mu.Unlock()
	return dbPath
}

func DB() *bun.DB {
	mu.Lock()
	defer mu.Unlock()
	return bunDB
}

// Init 打开数据库并执行迁移
// 该方法可重复调用（幂等）。
func Init(app *application.App) error {
	mu.Lock()
	defer mu.Unlock()

	if bunDB != nil {
		return nil
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	// 作为客户端软件，db 放到用户配置目录下更合理（随账号走，且可被备份/迁移）。
	dir := filepath.Join(cfgDir, "willchat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dbPath = filepath.Join(dir, "willchat.db")
	app.Logger.Info("db path", "path", dbPath)

	sqldb, err := sql.Open(sqliteshim.ShimName, dbPath)
	if err != nil {
		return err
	}

	// sqlite 设置单连接，并设置合理的超时。
	sqldb.SetMaxOpenConns(1)
	sqldb.SetMaxIdleConns(1)
	sqldb.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 基础连通性检查
	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return err
	}

	// 设置 journal_mode 和 synchronous
	_, _ = sqldb.ExecContext(ctx, `PRAGMA journal_mode = WAL;`)
	_, _ = sqldb.ExecContext(ctx, `PRAGMA synchronous = NORMAL;`)
	_, err = sqldb.ExecContext(ctx, `PRAGMA foreign_keys = ON;`)
	if err != nil {
		_ = sqldb.Close()
		return err
	}

	journalMode := ""
	_ = sqldb.QueryRowContext(ctx, `PRAGMA journal_mode;`).Scan(&journalMode)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	migrator := migrate.NewMigrator(db, migrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		_ = db.Close()
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		_ = db.Close()
		return err
	}

	sqlDB = sqldb
	bunDB = db

	if app != nil {
		if journalMode != "" {
			if journalMode == "wal" || journalMode == "WAL" {
				app.Logger.Info("sqlite journal_mode", "mode", journalMode)
			} else {
				app.Logger.Warn("sqlite journal_mode not wal", "mode", journalMode)
			}
		}
		if group != nil && !group.IsZero() {
			app.Logger.Info("db migrated", "path", dbPath, "group", group.String())
		} else {
			app.Logger.Debug("db migration up-to-date", "path", dbPath)
		}
	}

	return nil
}

func Close(app *application.App) error {
	mu.Lock()
	defer mu.Unlock()

	if bunDB == nil {
		return nil
	}
	err := bunDB.Close()
	sqlDB = nil
	bunDB = nil

	if err != nil && app != nil && !errors.Is(err, sql.ErrConnDone) {
		app.Logger.Warn("db close failed", "error", err)
	}
	return err
}
