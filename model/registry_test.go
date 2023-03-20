package model

import (
	"go-train/orm/internal/errs"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Get(t *testing.T) {

	testCases := []struct {
		name string

		registry Registry
		m        any

		wantFields []*Field

		wantModel *Model
		wantErr   error
	}{
		{
			name:     "normal entity",
			registry: NewRegistry(),
			m: func() any {

				type TestModel struct {
					Name      string
					TestField string
				}

				return &TestModel{}
			}(),
			wantFields: []*Field{
				{
					GoName:  "Name",
					ColName: "name",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(0),
				},
				{
					GoName:  "TestField",
					ColName: "test_field",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(16),
				},
			},
			wantModel: &Model{
				TabName: "test_model",
			},
		},
		{
			name:     "entity with tag",
			registry: NewRegistry(),
			m: func() any {

				type TestModel struct {
					Name      string `orm:"column=name11"`
					TestField string `orm:"column=test_field_tt,index=test"`
				}

				return &TestModel{}
			}(),
			wantFields: []*Field{
				{
					GoName:  "Name",
					ColName: "name11",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(0),
				},
				{
					GoName:  "TestField",
					ColName: "test_field_tt",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(16),
				},
			},
			wantModel: &Model{
				TabName: "test_model",
			},
		},
		{
			name:     "entity without tag",
			registry: NewRegistry(),
			m: func() any {

				type TestModel struct {
					Name      string
					TestField string `orm:"column=test_field_tt,index=test"`
				}

				return &TestModel{}
			}(),
			wantFields: []*Field{
				{
					GoName:  "Name",
					ColName: "name",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(0),
				},
				{
					GoName:  "TestField",
					ColName: "test_field_tt",
					GoType:  reflect.TypeOf(""),
					Offset:  uintptr(16),
				},
			},
			wantModel: &Model{
				TabName: "test_model",
			},
		},
		{
			name:     "entity with invalid tag",
			registry: NewRegistry(),
			m: func() any {

				type TestModel struct {
					Name      string `orm:"abc"`
					TestField string `orm:"column=test_field_tt,index=test"`
				}

				return &TestModel{}
			}(),
			wantErr: errs.NewErrTagInvalid("abc"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.wantErr == nil {
				fieldMap := make(map[string]*Field)
				columnMap := make(map[string]*Field)
				for _, field := range tc.wantFields {
					fieldMap[field.GoName] = field
					columnMap[field.ColName] = field
				}
				tc.wantModel.FieldMap = fieldMap
				tc.wantModel.ColumnMap = columnMap
			}

			model, err := tc.registry.Get(tc.m)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, model)
		})
	}

}
