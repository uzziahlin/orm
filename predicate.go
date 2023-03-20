package orm

type op string

func (o op) String() string {
	return string(o)
}

const (
	opEQ     op = "="
	opNOT    op = "NOT"
	opAND    op = "AND"
	opLT     op = "<"
	opLE     op = "<="
	opGT     op = ">"
	opGE     op = ">="
	opIN     op = "IN"
	opExists op = "EXIST"
)

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (p Predicate) expr() {}

func (p Predicate) AND(cond Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAND,
		right: cond,
	}
}

func (p Predicate) OR(cond Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAND,
		right: cond,
	}
}

func NOT(cond Predicate) Predicate {
	return Predicate{
		op:    opNOT,
		right: cond,
	}
}

func Exist(sub SubQuery) Predicate {
	sub.useAlias = false
	return Predicate{
		op:    opExists,
		right: sub,
	}
}
