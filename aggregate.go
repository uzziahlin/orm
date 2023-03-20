package orm

import "fmt"

type Aggregate struct {
	fn    string
	arg   string
	alias string
}

func (a Aggregate) selectedName() string {
	if a.alias != "" {
		return a.alias
	}
	return fmt.Sprintf("%s(%s)", a.fn, a.arg)
}

// expr Aggregate本身也是一个Expression，用在having
func (a Aggregate) expr() {}

func (a Aggregate) AS(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) EQ(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: ValueOf(val),
	}
}

func (a Aggregate) LT(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: ValueOf(val),
	}
}

func (a Aggregate) LE(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opLE,
		right: ValueOf(val),
	}
}

func (a Aggregate) GT(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: ValueOf(val),
	}
}

func (a Aggregate) GE(val any) Predicate {
	return Predicate{
		left:  a,
		op:    opGE,
		right: ValueOf(val),
	}
}

func Count(arg string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: arg,
	}
}

func Sum(arg string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: arg,
	}
}

func Max(arg string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: arg,
	}
}

func Min(arg string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: arg,
	}
}

func Avg(arg string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: arg,
	}
}
