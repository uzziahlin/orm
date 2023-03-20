package orm

import (
	"errors"
	"github.com/uzziahlin/orm/internal/errs"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

// Update 也可以看做是一个终结方法，重新回到 Inserter 里面
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.upsert = &Upsert{
		conflictColumns: o.conflictColumns,
		assigns:         assigns,
	}
	return o.i
}

type Upsert struct {
	conflictColumns []string
	assigns         []Assignable
}

type Inserter[T any] struct {
	Builder
	cols   []string
	values []*T
	upsert *Upsert
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.cols = cols
	return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Stat, error) {

	if len(i.values) == 0 {
		return nil, errors.New("插入零行")
	}

	if i.meta == nil {
		meta, err := i.registry.Get(i.values[0])
		if err != nil {
			return nil, err
		}
		i.meta = meta
	}

	i.builder.WriteString("INSERT INTO ")
	i.quote(i.meta.TabName)
	i.builder.WriteByte('(')

	fds := i.meta.Fields

	if len(i.cols) > 0 {
		fds = nil
		for _, col := range i.cols {
			fd, ok := i.meta.FieldMap[col]
			if !ok {
				return nil, errs.NewErrUnknownField(col)
			}
			fds = append(fds, fd)
		}
	}

	for idx, fd := range fds {
		if idx > 0 {
			i.builder.WriteByte(',')
		}
		i.builder.WriteString(fd.ColName)
	}

	i.builder.WriteString(") VALUES ")

	for _, val := range i.values {
		i.builder.WriteByte('(')
		valuer := i.creator(val, i.meta)
		for idx, fd := range fds {
			if idx > 0 {
				i.builder.WriteByte(',')
			}
			i.builder.WriteByte('?')

			f, err := valuer.GetField(fd.GoName)

			if err != nil {
				return nil, err
			}

			i.addArgs(f)
		}
		i.builder.WriteByte(')')
	}

	if i.upsert != nil {
		err := i.core.dialect.buildUpsert(&i.Builder, i.upsert)
		if err != nil {
			return nil, err
		}
	}

	return &Stat{
		Sql:  i.builder.String(),
		Args: i.args,
	}, nil
}
