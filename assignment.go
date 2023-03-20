package orm

type Assignable interface {
	assign()
}

func Assign(column string, val any) Assignment {
	v, ok := val.(Expression)
	if !ok {
		v = Value{val: val}
	}
	return Assignment{
		column: column,
		val:    v,
	}
}

type Assignment struct {
	column string
	val    Expression
}

func (a Assignment) assign() {}
