package orm

import (
	"context"
	"database/sql"
	"github.com/uzziahlin/orm/internal/valuer"
	"github.com/uzziahlin/orm/model"
)

type Session interface {
	getCore() core

	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type core struct {
	registry model.Registry
	creator  valuer.Creator
	dialect  Dialect
	mdls     []MiddleWare
}
