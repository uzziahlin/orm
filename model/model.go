package model

import (
	"reflect"
)

type Model struct {
	TabName   string
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
	Fields    []*Field
}

type Option func(m *Model) error

type Field struct {
	GoName  string
	ColName string
	GoType  reflect.Type
	Offset  uintptr
}
