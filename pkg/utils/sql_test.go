package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddFieldsToQuery(t *testing.T) {
	type args struct {
		field string
		value interface{}
	}
	tests := []struct {
		name       string
		args       args
		add        bool
		wantFields string
		wantArgs   interface{}
	}{
		{
			name: "success with int",
			args: args{
				field: "field1",
				value: 1,
			},
			add:        true,
			wantFields: "field1 = $1",
			wantArgs:   1,
		},
		{
			name: "success with string",
			args: args{
				field: "field1",
				value: "test",
			},
			add:        true,
			wantFields: "field1 = $1",
			wantArgs:   "test",
		},
		{
			name: "success with bool",
			args: args{
				field: "field1",
				value: true,
			},
			add:        true,
			wantFields: "field1 = $1",
			wantArgs:   true,
		},
		{
			name: "success with time",
			args: args{
				field: "field1",
				value: time.Now().Add(5 * time.Second).Day(),
			},
			add:        true,
			wantFields: "field1 = $1",
			wantArgs:   time.Now().Add(5 * time.Second).Day(),
		},
		{
			name: "error",
			args: args{
				field: "field1",
				value: nil,
			},
			add: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				fields []string
				args   []interface{}
				argIdx int
			)
			AddFieldsToQuery(tt.args.field, tt.args.value, &fields, &args, &argIdx)
			if !tt.add {
				return
			}
			require.Equal(t, tt.wantFields, fields[0])
			require.Equal(t, argIdx, 1)
			require.Equal(t, tt.wantArgs, args[0])
		})
	}
}
