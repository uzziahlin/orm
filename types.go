package orm

import (
	"context"
	"database/sql"
)

type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

type Executor[T any] interface {
	Exec() (Result, error)
}

type Result interface {
	sql.Result
}

type SQLBuilder interface {
	Build() (*Stat, error)
}

type Stat struct {
	Sql  string
	Args []any
}

/*type SQL interface {
	Stat() (string, error)
	Args() ([]any, error)
}*/
