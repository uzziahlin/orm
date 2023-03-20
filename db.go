package orm

import (
	"context"
	"database/sql"
	"github.com/uzziahlin/orm/internal/valuer"
	"github.com/uzziahlin/orm/model"
)

type DBOption func(db *DB)

// DB Database抽象，用来描述一个数据库实例
type DB struct {
	core
	*sql.DB
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

func Open(driverName, dsn string, opts ...DBOption) (*DB, error) {

	db, err := sql.Open(driverName, dsn)

	if err != nil {
		return nil, err
	}

	return OpenDB(db, opts...)

}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {

	res := &DB{
		core: core{
			registry: model.NewRegistry(),
			creator:  valuer.NewUnsafeValuer,
			dialect:  &mysqlDialect{},
		},
		DB: db,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func DBWithRegistry(reg model.Registry) DBOption {
	return func(db *DB) {
		db.registry = reg
	}
}

func DBWithCreator(creator valuer.Creator) DBOption {
	return func(db *DB) {
		db.creator = creator
	}
}

func DBWithMiddlewares(mdls ...MiddleWare) DBOption {
	return func(db *DB) {
		db.mdls = mdls
	}
}
