package orm

import (
	"errors"
	"github.com/uzziahlin/orm/internal/errs"
	"github.com/uzziahlin/orm/model"
	"strings"
)

type Builder struct {
	core
	table   TableReference
	builder *strings.Builder
	meta    *model.Model
	args    []any
	sess    Session
	quoter  byte
}

func (s *Builder) quote(name string) {
	s.builder.WriteByte(s.quoter)
	s.builder.WriteString(name)
	s.builder.WriteByte(s.quoter)
}

func (s *Builder) BuildPredicates(predicates ...Predicate) error {

	expr := predicates[0]

	for i := 1; i < len(predicates); i++ {
		expr = expr.AND(predicates[i])
	}

	err := s.buildExpression(expr)

	if err != nil {
		return err
	}

	return nil
}

func (s *Builder) buildExpression(expr Expression) error {
	switch elem := expr.(type) {
	case nil:
	case Predicate:
		_, ok := elem.left.(Predicate)
		if ok {
			s.builder.WriteString("(")
		}

		err := s.buildExpression(elem.left)

		if ok {
			s.builder.WriteString(")")
		}

		if err != nil {
			return err
		}

		if elem.op != "" {
			s.builder.WriteString(" ")
			s.builder.WriteString(elem.op.String())
			s.builder.WriteString(" ")
		}

		_, ok = elem.right.(Predicate)
		if ok {
			s.builder.WriteString("(")
		}

		err = s.buildExpression(elem.right)

		if ok {
			s.builder.WriteString(")")
		}

		if err != nil {
			return err
		}
	case Column:
		err := s.buildColumn(elem)

		if err != nil {
			return err
		}
	case Value:
		s.builder.WriteString(" ? ")
		s.addArgs(elem.val)
	case RawExpr:
		s.builder.WriteString(elem.exp)
		s.addArgs(elem.args...)
	case Aggregate:
		err := s.buildAggregate(elem)

		if err != nil {
			return err
		}
	case SubQuery:
		err := s.buildSubQuery(elem)

		if err != nil {
			return err
		}
	default:
		return errors.New("不支持的类型")
	}

	return nil
}

func (s *Builder) buildSubQuery(sub SubQuery) error {
	stat, err := sub.b.Build()
	if err != nil {
		return err
	}
	s.builder.WriteByte('(')
	s.builder.WriteString(stat.Sql)
	s.builder.WriteByte(')')
	if sub.useAlias && sub.alias != "" {
		s.builder.WriteString(" AS ")
		s.quote(sub.alias)
	}
	s.addArgs(stat.Args...)
	return nil
}

func (s *Builder) addArgs(args ...any) {
	if len(args) == 0 {
		return
	}
	s.args = append(s.args, args...)
}

func (s *Builder) buildColumn(col Column) error {

	var alias string

	table := col.table

	if table != nil {
		alias = table.tableAlias()
	}

	if alias != "" {
		s.quote(alias)
		s.builder.WriteByte('.')
	}

	colName, err := s.colName(table, col.name)

	if err != nil {
		return err
	}

	s.quote(colName)

	if col.alias != "" {
		s.buildAlias(col.alias)
	}

	return nil
}

func (s *Builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	case nil:
		f, ok := s.meta.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return f.ColName, nil
	case Table:
		meta, err := s.registry.Get(tab.entity)
		if err != nil {
			return "", err
		}
		f, ok := meta.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return f.ColName, nil
	case SubQuery:
		if len(tab.cols) > 0 {
			for _, col := range tab.cols {
				if col.selectedName() == fd {
					return fd, nil
				}
			}
			return "", errs.NewErrUnknownField(fd)
		}
		return s.colName(tab.table, fd)
	default:
		return "", errs.NewErrUnsupportedTableType(table)
	}
}

func (s *Builder) buildAggregate(aggregate Aggregate) error {
	s.builder.WriteString(aggregate.fn)
	s.builder.WriteByte('(')
	err := s.buildColumn(C(aggregate.arg))
	if err != nil {
		return err
	}
	s.builder.WriteByte(')')

	if aggregate.alias != "" {
		s.buildAlias(aggregate.alias)
	}

	return nil
}

func (s *Builder) buildAlias(alias string) {
	s.builder.WriteString(" AS ")
	s.quote(alias)
}
