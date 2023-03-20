package orm

import (
	"context"
	"github.com/uzziahlin/orm/model"
)

type QueryContext struct {
	Type    string
	builder SQLBuilder
	model   *model.Model
	stat    *Stat
}

func (qc *QueryContext) Query() (*Stat, error) {
	if qc.stat != nil {
		return qc.stat, nil
	}
	var err error
	qc.stat, err = qc.builder.Build()
	return qc.stat, err
}

type QueryResult struct {
	Result any
	err    error
}

type HandleFunc func(ctx context.Context, qc *QueryContext) *QueryResult

type MiddleWare func(handler HandleFunc) HandleFunc
