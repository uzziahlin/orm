package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleter_Build(t *testing.T) {

	db, err := Open("", "")

	if err != nil {
		return
	}

	testCases := []struct {
		name     string
		builder  SQLBuilder
		wantStat *Stat
		wantErr  error
	}{
		{
			name:    "normal",
			builder: NewDeleter[TestModel](db),
			wantStat: &Stat{
				Sql: "delete from test_model",
			},
		},
		{
			name:    "where",
			builder: NewDeleter[TestModel](db).Where(C("Name").EQ("Jack").AND(C("Age").LT(18))),
			wantStat: &Stat{
				Sql: "delete from test_model where (`name` =  ? ) AND (`age` <  ? )",
				Args: []any{
					"Jack",
					18,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stat, err := tc.builder.Build()

			assert.Equal(t, tc.wantErr, err)

			if err != nil {
				return
			}

			assert.Equal(t, tc.wantStat, stat)
		})
	}
}
