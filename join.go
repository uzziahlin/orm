package orm

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

func (b *JoinBuilder) On(pres ...Predicate) Join {
	return Join{
		left:  b.left,
		right: b.right,
		typ:   b.typ,
		on:    pres,
	}
}

func (b *JoinBuilder) Using(cs ...string) Join {
	return Join{
		left:  b.left,
		right: b.right,
		typ:   b.typ,
		using: cs,
	}
}

type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (j Join) tableAlias() string {
	return ""
}

func (j Join) Join(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: table,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: table,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(table TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: table,
		typ:   "RIGHT JOIN",
	}
}

type SubQuery struct {
	b        SQLBuilder
	table    TableReference
	cols     []Selectable
	alias    string
	useAlias bool
}

func (s SubQuery) expr() {

}

func (s SubQuery) tableAlias() string {
	return s.alias
}

func (s SubQuery) C(name string) Column {
	return Column{
		name:  name,
		table: s,
	}
}
