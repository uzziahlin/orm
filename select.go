package orm

import (
	"context"
	"github.com/uzziahlin/orm/internal/errs"
	"reflect"
	"strconv"
	"strings"
)

type Selectable interface {
	selectedName() string
}

type Selector[T any] struct {
	Builder
	selectable []Selectable
	where      []Predicate
	groupCols  []Column
	having     []Predicate
	orderCols  []Column
	offset     int
	limit      int
}

func NewSelector[T any](sess Session) *Selector[T] {

	c := sess.getCore()

	builder := Builder{
		sess:   sess,
		core:   c,
		quoter: c.dialect.quoter(),
	}
	return &Selector[T]{
		Builder: builder,
	}
}

func (s *Selector[T]) Select(selectable ...Selectable) *Selector[T] {
	s.selectable = selectable
	return s
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(conds ...Predicate) *Selector[T] {
	s.where = conds
	return s
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupCols = cols
	return s
}

func (s *Selector[T]) Having(predicates ...Predicate) *Selector[T] {
	s.having = predicates
	return s
}

func (s *Selector[T]) OrderBy(cols ...Column) *Selector[T] {
	s.orderCols = cols
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Build() (*Stat, error) {

	defer func() {
		s.builder.Reset()
		s.args = nil
	}()

	if err := s.init(); err != nil {
		return nil, err
	}

	if err := s.buildSelect(); err != nil {
		return nil, err
	}

	if err := s.BuildFrom(); err != nil {
		return nil, err
	}

	if len(s.where) > 0 {
		if err := s.buildWhere(); err != nil {
			return nil, err
		}
	}

	if len(s.groupCols) > 0 {
		if err := s.buildGroup(); err != nil {
			return nil, err
		}

		// having必须有group，不允许单独设置having
		if len(s.having) > 0 {
			if err := s.buildHaving(); err != nil {
				return nil, err
			}
		}
	}

	if len(s.orderCols) > 0 {
		if err := s.buildOrder(); err != nil {
			return nil, err
		}
	}

	if s.offset > 0 {
		s.buildOffset()
	}

	if s.limit > 0 {
		s.buildLimit()
	}

	return &Stat{
		Sql:  s.builder.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) init() error {
	var (
		t   T
		err error
	)

	s.meta, err = s.registry.Get(&t)

	if err != nil {
		return err
	}

	s.builder = &strings.Builder{}

	return nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	root := s.handler()

	for _, md := range s.mdls {
		root = md(root)
	}

	qc := &QueryContext{
		Type:    "SELECT",
		builder: s,
		model:   s.meta,
	}

	res := root(ctx, qc)

	if res.err != nil {
		return nil, res.err
	}

	return res.Result.(*T), nil
}

func (s *Selector[T]) handler() HandleFunc {
	return func(ctx context.Context, qc *QueryContext) *QueryResult {
		stat, err := qc.Query()

		if err != nil {
			return &QueryResult{
				Result: nil,
				err:    err,
			}
		}

		rows, err := s.sess.QueryContext(ctx, stat.Sql, stat.Args...)

		if err != nil {
			return &QueryResult{
				Result: nil,
				err:    err,
			}
		}

		if !rows.Next() {
			return &QueryResult{
				Result: nil,
				err:    err,
			}
		}

		tp := new(T)
		meta, err := s.registry.Get(tp)
		if err != nil {
			return &QueryResult{
				Result: nil,
				err:    err,
			}
		}

		val := s.creator(tp, meta)
		err = val.SetColumns(rows)

		return &QueryResult{
			Result: tp,
			err:    err,
		}
	}
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	stat, err := s.Build()

	if err != nil {
		return nil, err
	}

	rows, err := s.sess.QueryContext(ctx, stat.Sql, stat.Args...)

	if err != nil {
		return nil, err
	}

	tp := new(T)

	meta, err := s.registry.Get(tp)

	val := s.creator(tp, meta)

	res := make([]*T, 0)

	for rows.Next() {
		err = val.SetColumns(rows)

		if err != nil {
			return nil, err
		}

		t := new(T)
		*t = *tp
		res = append(res, t)
	}

	return res, nil
}

func (s *Selector[T]) bindResult(tp *T, cols []string) ([]any, error) {

	meta, err := s.registry.Get(tp)

	if err != nil {
		return nil, err
	}

	tVal := reflect.ValueOf(tp).Elem()

	vals := make([]any, 0, len(cols))

	for _, col := range cols {
		field, ok := meta.ColumnMap[col]
		if !ok {
			return nil, errs.NewErrUnknownField(col)
		}
		fd := tVal.FieldByName(field.GoName)

		vals = append(vals, fd.Addr().Interface())
	}

	return vals, nil
}

func (s *Selector[T]) BuildFrom() error {
	s.builder.WriteString(" FROM ")
	return s.BuildTable(s.table)
}

func (s *Selector[T]) BuildTable(table TableReference) error {
	switch tab := table.(type) {
	case Table:
		meta, err := s.registry.Get(tab.entity)
		if err != nil {
			return err
		}
		s.quote(meta.TabName)
		if tab.alias != "" {
			s.builder.WriteString(" AS ")
			s.quote(tab.alias)
		}
	case nil:
		s.quote(s.meta.TabName)
	case Join:
		err := s.BuildTable(tab.left)
		if err != nil {
			return err
		}

		s.builder.WriteString(" ")
		s.builder.WriteString(tab.typ)
		s.builder.WriteString(" ")

		err = s.BuildTable(tab.right)
		if err != nil {
			return err
		}

		if len(tab.on) > 0 {
			s.builder.WriteString(" ON ")
			err := s.BuildPredicates(tab.on...)
			if err != nil {
				return err
			}
		}

		if len(tab.using) > 0 {
			s.builder.WriteString(" USING (")
			for idx, col := range tab.using {
				if idx > 0 {
					s.builder.WriteByte(',')
				}
				s.quote(col)
			}
			s.builder.WriteString(")")
		}
	case SubQuery:
		err := s.buildSubQuery(tab)

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector[T]) buildSelect() error {
	s.builder.WriteString("SELECT ")

	if len(s.selectable) == 0 {
		s.builder.WriteString("*")
		return nil
	}

	for idx, sel := range s.selectable {
		if idx > 0 {
			s.builder.WriteByte(',')
		}
		switch elem := sel.(type) {
		case Column:
			err := s.buildColumn(elem)

			if err != nil {
				return err
			}
		case Aggregate:
			err := s.buildAggregate(elem)

			if err != nil {
				return err
			}
		case RawExpr:
			s.builder.WriteString(elem.exp)
		}
	}

	return nil
}

func (s *Selector[T]) buildGroup() error {
	s.builder.WriteString(" GROUP BY ")

	for idx, col := range s.groupCols {
		if idx > 0 {
			s.builder.WriteString(",")
		}
		if err := s.buildColumn(col); err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector[T]) buildWhere() error {
	s.builder.WriteString(" WHERE ")
	return s.BuildPredicates(s.where...)
}

func (s *Selector[T]) buildHaving() error {
	s.builder.WriteString(" HAVING ")
	return s.BuildPredicates(s.having...)
}

func (s *Selector[T]) buildOrder() error {
	s.builder.WriteString(" ORDER BY ")
	for idx, col := range s.orderCols {
		if idx > 0 {
			s.builder.WriteByte(',')
		}
		if err := s.buildColumn(col); err != nil {
			return err
		}
		if col.sort != "" {
			s.builder.WriteString(" ")
			s.builder.WriteString(col.sort)
		}
	}
	return nil
}

func (s *Selector[T]) buildOffset() {
	s.builder.WriteString(" OFFSET ")
	s.builder.WriteString(strconv.Itoa(s.offset))
}

func (s *Selector[T]) buildLimit() {
	s.builder.WriteString(" LIMIT ")
	s.builder.WriteString(strconv.Itoa(s.limit))
}

func (s *Selector[T]) AsSubQuery(alias string) SubQuery {
	var table TableReference

	if s.table == nil {
		table = TableOf(new(T))
	}
	return SubQuery{
		cols:     s.selectable,
		table:    table,
		b:        s,
		alias:    alias,
		useAlias: true,
	}
}
