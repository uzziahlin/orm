package orm

import "github.com/uzziahlin/orm/internal/errs"

var (
	_ Dialect = &mysqlDialect{}
	_ Dialect = &sqlite3Dialect{}
)

// Dialect 对数据库方言的抽象，因为有些sql语法在不同的方言会有不同实现
type Dialect interface {
	quoter() byte
	buildUpsert(i *Builder, upsert *Upsert) error
}

type mysqlDialect struct {
}

func (m mysqlDialect) buildUpsert(b *Builder, odk *Upsert) error {
	b.builder.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, a := range odk.assigns {
		if idx > 0 {
			b.builder.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			colName, err := b.colName(assign.table, assign.name)
			if err != nil {
				return err
			}
			b.quote(colName)
			b.builder.WriteString("=VALUES(")
			b.quote(colName)
			b.builder.WriteByte(')')
		case Assignment:
			err := b.buildColumn(C(assign.column))
			if err != nil {
				return err
			}
			b.builder.WriteString("=")
			return b.buildExpression(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}

func (m mysqlDialect) quoter() byte {
	return '`'
}

type standardSQLDialect struct {
}

func (s standardSQLDialect) buildUpsert(b *Builder, odk *Upsert) error {
	b.builder.WriteString(" ON CONFLICT")
	if len(odk.conflictColumns) > 0 {
		b.builder.WriteByte('(')
		for i, col := range odk.conflictColumns {
			if i > 0 {
				b.builder.WriteByte(',')
			}
			err := b.buildColumn(C(col))
			if err != nil {
				return err
			}
		}
		b.builder.WriteByte(')')
	}
	b.builder.WriteString(" DO UPDATE SET ")

	for idx, a := range odk.assigns {
		if idx > 0 {
			b.builder.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			colName, err := b.colName(assign.table, assign.name)
			if err != nil {
				return err
			}
			b.quote(colName)
			b.builder.WriteString("=excluded.")
			b.quote(colName)
		case Assignment:
			err := b.buildColumn(C(assign.column))
			if err != nil {
				return err
			}
			b.builder.WriteString("=")
			return b.buildExpression(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}

func (s standardSQLDialect) quoter() byte {
	return '`'
}

type sqlite3Dialect struct {
	standardSQLDialect
}
