package orm

type Column struct {
	table TableReference
	name  string
	alias string
	sort  string
}

// C 用于构造列信息
// C("fieldName").EQ(val)
func C(field string) Column {
	return Column{
		name: field,
	}
}

// expr 表示Column是表达式
func (c Column) expr() {}

func (c Column) assign() {}

func (c Column) selectedName() string {
	if c.alias != "" {
		return c.alias
	}
	return c.name
}

func (c Column) AS(alias string) Column {
	return Column{
		table: c.table,
		name:  c.name,
		alias: alias,
	}
}

func (c Column) ASC() Column {
	return Column{
		table: c.table,
		name:  c.name,
		sort:  "ASC",
	}
}

func (c Column) DESC() Column {
	return Column{
		table: c.table,
		name:  c.name,
		sort:  "DESC",
	}
}

func (c Column) InQuery(sub SubQuery) Predicate {
	sub.useAlias = false
	return Predicate{
		left:  c,
		op:    opIN,
		right: sub,
	}
}

func (c Column) EQ(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: ValueOf(val),
	}
}

func (c Column) LT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: ValueOf(val),
	}
}

func (c Column) LE(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLE,
		right: ValueOf(val),
	}
}

func (c Column) GT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: ValueOf(val),
	}
}

func (c Column) GE(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGE,
		right: ValueOf(val),
	}
}

func ValueOf(val any) Expression {
	switch v := val.(type) {
	case Expression:
		return v
	default:
		return Value{
			val: val,
		}
	}
}

type Value struct {
	val any
}

func (v Value) expr() {}
