package valuer

import (
	"database/sql"
	"github.com/uzziahlin/orm/internal/errs"
	"github.com/uzziahlin/orm/model"
	"reflect"
)

type reflectValuer struct {
	tp   reflect.Value
	meta *model.Model
}

func NewReflectValuer(tp any, meta *model.Model) Valuer {
	val := reflect.ValueOf(tp).Elem()
	return &reflectValuer{
		tp:   val,
		meta: meta,
	}
}

func (r *reflectValuer) SetColumns(rows *sql.Rows) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cols))

	for _, col := range cols {
		fd, ok := r.meta.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		vals = append(vals, r.tp.FieldByName(fd.GoName).Addr().Interface())
	}

	return rows.Scan(vals...)
}

func (r *reflectValuer) GetField(name string) (any, error) {
	fd := r.tp.FieldByName(name)

	if fd == (reflect.Value{}) {
		return nil, errs.NewErrUnknownField(name)
	}

	return fd.Interface(), nil
}
