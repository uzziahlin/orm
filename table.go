package orm

type TableReference interface {
	tableAlias() string
}

type Table struct {
	entity any
	alias  string
}

func (t Table) AS(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) tableAlias() string {
	return t.alias
}

func (t Table) Join(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: table,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: table,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: table,
		typ:   "RIGHT JOIN",
	}
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}
