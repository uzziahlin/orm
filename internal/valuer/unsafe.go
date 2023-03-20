package valuer

import (
	"database/sql"
	"github.com/uzziahlin/orm/internal/errs"
	"github.com/uzziahlin/orm/model"
	"reflect"
	"unsafe"
)

type unsafeValuer struct {
	address unsafe.Pointer
	meta    *model.Model
}

func NewUnsafeValuer(tp any, meta *model.Model) Valuer {
	ptr := reflect.ValueOf(tp).UnsafePointer()
	return &unsafeValuer{
		address: ptr,
		meta:    meta,
	}
}

func (u *unsafeValuer) SetColumns(rows *sql.Rows) error {

	cols, err := rows.Columns()

	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cols))

	for _, col := range cols {
		fd, ok := u.meta.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		ptr := unsafe.Pointer(uintptr(u.address) + fd.Offset)
		vals = append(vals, reflect.NewAt(fd.GoType, ptr).Interface())
	}

	return rows.Scan(vals...)
}

func (u *unsafeValuer) GetField(name string) (any, error) {
	fd, ok := u.meta.FieldMap[name]

	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}

	ptr := unsafe.Pointer(uintptr(u.address) + fd.Offset)

	val := reflect.NewAt(fd.GoType, ptr).Elem()

	return val.Interface(), nil
}
