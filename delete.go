package orm

import (
	"strings"
)

type Deleter[T any] struct {
	Builder
	where []Predicate
}

func NewDeleter[T any](sess Session) *Deleter[T] {
	builder := Builder{
		sess: sess,
	}
	return &Deleter[T]{
		Builder: builder,
	}
}

func (d *Deleter[T]) From(table TableReference) *Deleter[T] {
	d.table = table
	return d
}

func (d *Deleter[T]) Where(conds ...Predicate) *Deleter[T] {
	d.where = conds
	return d
}

func (d *Deleter[T]) Build() (*Stat, error) {
	var (
		t   T
		err error
	)
	d.meta, err = d.registry.Get(&t)

	if err != nil {
		return nil, err
	}

	d.builder = &strings.Builder{}

	d.builder.WriteString("delete from ")

	d.buildFrom()

	if len(d.where) > 0 {
		err := d.BuildPredicates()

		if err != nil {
			return nil, err
		}
	}

	return &Stat{
		Sql:  d.builder.String(),
		Args: d.args,
	}, nil

}

func (d *Deleter[T]) Exec() (Result, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Deleter[T]) buildFrom() {
	switch tab := d.table.(type) {
	case Table:
		meta, err := d.registry.Get(tab.entity)
		if err != nil {
			return
		}
		d.builder.WriteString(meta.TabName)
		if tab.alias != "" {
			d.builder.WriteString(" AS ")
			d.builder.WriteString(tab.alias)
		}
	case nil:
		d.builder.WriteString(d.meta.TabName)
	}
}
