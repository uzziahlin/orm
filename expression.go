package orm

type Expression interface {
	expr()
}

type RawExpr struct {
	exp  string
	args []any
}

func (r RawExpr) expr() {}

func (r RawExpr) selectedName() string {
	return ""
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		exp:  expr,
		args: args,
	}
}
