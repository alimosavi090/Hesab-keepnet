package database

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type connPool struct {
	*sql.Conn
}

func WithImmediateTx(ctx context.Context, gdb *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("unwrap sql db: %w", err)
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Close()

	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("begin immediate: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
		}
	}()

	tx := gdb.Session(&gorm.Session{
		NewDB:                  true,
		SkipDefaultTransaction: true,
		Context:                ctx,
	})
	tx.Statement.ConnPool = connPool{conn}

	if err = fn(tx); err != nil {
		return err
	}

	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true

	return nil
}
