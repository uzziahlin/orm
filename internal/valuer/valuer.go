package valuer

import (
	"database/sql"
	"github.com/uzziahlin/orm/model"
)

type Valuer interface {
	SetColumns(rows *sql.Rows) error
	GetField(name string) (any, error)
}

type Creator func(tp any, meta *model.Model) Valuer
