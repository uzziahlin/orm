package orm

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/uzziahlin/orm/internal/errs"
	"testing"
)

func TestSelector_Build(t *testing.T) {

	db := memoryDB(t)

	testCases := []struct {
		name string

		sb SQLBuilder

		wantStat *Stat
		wantErr  error
	}{
		{
			name: "Model",

			sb: NewSelector[TestModel](db),

			wantStat: &Stat{
				Sql: "SELECT * FROM `test_model`",
			},
		},
		{
			name: "given from",

			sb: NewSelector[TestModel](db).From(TableOf(&TestModel1{})),

			wantStat: &Stat{
				Sql: "SELECT * FROM `test_model_1`",
			},
		},
		{
			name: "where AND",

			sb: NewSelector[TestModel](db).Where(C("Name").EQ("Jack").AND(C("Age").EQ(18))),

			wantStat: &Stat{
				Sql: "SELECT * FROM `test_model` WHERE (`name` =  ? ) AND (`age` =  ? )",
				Args: []any{
					"Jack",
					18,
				},
			},
		},
		{
			name: "where NOT",

			sb: NewSelector[TestModel](db).Where(NOT(C("Name").EQ("Jack"))),

			wantStat: &Stat{
				Sql: "SELECT * FROM `test_model` WHERE  NOT (`name` =  ? )",
				Args: []any{
					"Jack",
				},
			},
		},
		{
			name: "unknown Field",

			sb: NewSelector[TestModel](db).Where(C("Name").EQ("Jack").AND(C("Gender").EQ(18))),

			wantErr: errs.NewErrUnknownField("Gender"),
		},
		{
			name: "test camel Field",

			sb: NewSelector[TestModel](db).Where(C("Name").EQ("Jack").AND(C("TestField").EQ("test"))),

			wantStat: &Stat{
				Sql: "SELECT * FROM `test_model` WHERE (`name` =  ? ) AND (`test_field` =  ? )",
				Args: []any{
					"Jack",
					"test",
				},
			},
		},
		{
			name: "test select",

			sb: NewSelector[TestModel](db).Select(C("Name"), Avg("Age")),

			wantStat: &Stat{
				Sql: "SELECT `name`,AVG(`age`) FROM `test_model`",
			},
		},
		{
			name: "test select alias",

			sb: NewSelector[TestModel](db).Select(C("Name").AS("t_name"), Avg("Age").AS("age_avg")),

			wantStat: &Stat{
				Sql: "SELECT `name` AS `t_name`,AVG(`age`) AS `age_avg` FROM `test_model`",
			},
		},
		{
			name: "test Join",

			sb: func() SQLBuilder {
				orderInfoTab := TableOf(&OrderInfo{}).AS("oi")
				orderDetailTab := TableOf(&OrderDetail{}).AS("od")

				return NewSelector[TestModel](db).Select(orderInfoTab.C("Total"), orderDetailTab.C("ItemId").AS("skuId")).
					From(orderInfoTab.LeftJoin(orderDetailTab).On(orderInfoTab.C("Id").EQ(orderDetailTab.C("OrderId"))))
			}(),

			wantStat: &Stat{
				Sql: "SELECT `oi`.`total`,`od`.`item_id` AS `skuId` FROM `order_info` AS `oi` LEFT JOIN `order_detail` AS `od` ON `oi`.`id` = `od`.`order_id`",
			},
		},
		{
			name: "left multiple join",
			sb: func() SQLBuilder {
				t1 := TableOf(&Order{}).AS("t1")
				t2 := TableOf(&OrderDetail{}).AS("t2")
				t3 := TableOf(&Item{}).AS("t3")
				return NewSelector[Order](db).
					From(t1.LeftJoin(t2).
						On(t1.C("Id").EQ(t2.C("OrderId"))).
						LeftJoin(t3).On(t2.C("ItemId").EQ(t3.C("Id"))))
			}(),
			wantStat: &Stat{
				Sql: "SELECT * FROM `order` AS `t1` LEFT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id` LEFT JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`",
			},
		},
		{
			name: "subquery from",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")

				return NewSelector[Order](db).From(sub)
			}(),
			wantStat: &Stat{
				Sql: "SELECT * FROM (SELECT * FROM `order_detail`) AS `sub`",
			},
		},
		{
			name: "subquery select",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")

				return NewSelector[Order](db).Select(sub.C("OrderId")).From(sub)
			}(),
			wantStat: &Stat{
				Sql: "SELECT `sub`.`order_id` FROM (SELECT * FROM `order_detail`) AS `sub`",
			},
		},
		{
			name: "subquery select alias",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId").AS("id")).AsSubQuery("sub")

				return NewSelector[Order](db).Select(sub.C("id")).From(sub)
			}(),
			wantStat: &Stat{
				Sql: "SELECT `sub`.`id` FROM (SELECT `order_id` AS `id` FROM `order_detail`) AS `sub`",
			},
		},
		{
			name: "subQuery in",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubQuery("sub")
				return NewSelector[Order](db).Where(C("Id").InQuery(sub))
			}(),
			wantStat: &Stat{
				Sql: "SELECT * FROM `order` WHERE `id` IN (SELECT `order_id` FROM `order_detail`)",
			},
		},
		{
			name: "subQuery exist",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubQuery("sub")
				return NewSelector[Order](db).Where(Exist(sub))
			}(),
			wantStat: &Stat{
				Sql: "SELECT * FROM `order` WHERE  EXIST (SELECT `order_id` FROM `order_detail`)",
			},
		},
		{
			name: "subQuery not exist",
			sb: func() SQLBuilder {
				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubQuery("sub")
				return NewSelector[Order](db).Where(NOT(Exist(sub)))
			}(),
			wantStat: &Stat{
				Sql: "SELECT * FROM `order` WHERE  NOT ( EXIST (SELECT `order_id` FROM `order_detail`))",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stat, err := tc.sb.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantStat, stat)
		})
	}
}

func memoryDB(t *testing.T, opts ...DBOption) *DB {
	orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

type OrderInfo struct {
	Id    string
	Total float64
}

type OrderDetail struct {
	DetailId string
	ItemId   string
	OrderId  string

	UsingCol1 string
	UsingCol2 string
}

type Order struct {
	Id        int
	UsingCol1 string
	UsingCol2 string
}

type Item struct {
	Id int
}

type TestModel1 struct {
	Name string

	Age int

	TestField string
}

type TestModel struct {
	Name string

	Age int

	TestField string
}

func TestSelector_Get(t *testing.T) {

	mockDB, mock, err := sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)

	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  *TestModel
	}{
		{
			name:    "query error",
			mockErr: errors.New("invalid query"),
			wantErr: errors.New("invalid query"),
			query:   "select .*",
		},
		{
			name:     "no row",
			wantErr:  errs.ErrEmptyResult,
			query:    "select .*",
			mockRows: sqlmock.NewRows([]string{"name"}),
		},
		{
			name:  "get data",
			query: "select .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"name", "age", "test_field"})
				res.AddRow([]byte("Jack"), []byte("18"), []byte("test"))
				return res
			}(),
			wantVal: &TestModel{
				Name:      "Jack",
				Age:       18,
				TestField: "test",
			},
		},
	}

	for _, tc := range testCases {
		exp := mock.ExpectQuery(tc.query)
		if tc.mockErr != nil {
			exp.WillReturnError(tc.mockErr)
		} else {
			exp.WillReturnRows(tc.mockRows)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

func TestSelector_GetMulti(t *testing.T) {
	mockDB, mock, err := sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)

	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVals []*TestModel
	}{
		{
			name:  "test multi",
			query: "select .*",
			mockRows: func() *sqlmock.Rows {
				res := sqlmock.NewRows([]string{"name", "age", "test_field"})
				res.AddRow([]byte("Jack"), []byte("18"), []byte("test"))
				res.AddRow([]byte("Tom"), []byte("17"), []byte("test1"))
				res.AddRow([]byte("Jetty"), []byte("19"), []byte("test11"))
				res.AddRow([]byte("Jerry"), []byte("25"), []byte("test111"))
				res.AddRow([]byte("Ken"), []byte("29"), []byte("test1111"))
				return res
			}(),
			wantVals: []*TestModel{
				{
					Name:      "Jack",
					Age:       18,
					TestField: "test",
				},
				{
					Name:      "Tom",
					Age:       17,
					TestField: "test1",
				},
				{
					Name:      "Jetty",
					Age:       19,
					TestField: "test11",
				},
				{
					Name:      "Jerry",
					Age:       25,
					TestField: "test111",
				},
				{
					Name:      "Ken",
					Age:       29,
					TestField: "test1111",
				},
			},
		},
	}

	for _, tc := range testCases {
		exp := mock.ExpectQuery(tc.query)
		if tc.mockErr != nil {
			exp.WillReturnError(tc.mockErr)
		} else {
			exp.WillReturnRows(tc.mockRows)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).GetMulti(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, res)
		})
	}

}
