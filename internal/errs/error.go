package errs

import (
	"errors"
	"fmt"
	"github.com/uzziahlin/orm"
)

var (
	ErrUnsupportedType = errors.New("不支持的类型")
	ErrUnknownField    = errors.New("未知列")
	ErrTagInvalid      = errors.New("orm: tag不可用")
	ErrEmptyResult     = errors.New("orm：结果集为空")
	ErrUnknownColumn   = errors.New("orm: 未知字段")

	ErrUnsupportedTableType = errors.New("orm:不支持的表类型")
)

func NewErrUnsupportedType(typ string) error {
	return fmt.Errorf("%w, %s", ErrUnsupportedType, typ)
}

func NewErrUnsupportedTableType(tab any) error {
	return fmt.Errorf("%w, %v", ErrUnsupportedTableType, tab)
}

func NewErrUnknownField(f string) error {
	return fmt.Errorf("%w, %s", ErrUnknownField, f)
}

func NewErrTagInvalid(tag string) error {
	return fmt.Errorf("%w, %s", ErrTagInvalid, tag)
}

func NewErrEmptyResult() error {
	return ErrEmptyResult
}

func NewErrUnknownColumn(col string) error {
	return fmt.Errorf("%w, %s", ErrUnknownColumn, col)
}

func NewErrUnsupportedAssignableType(a orm.Assignable) error {
	return fmt.Errorf("test")
}
