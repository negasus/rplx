package rplx

import (
	"reflect"
	"testing"
)

func TestVariableNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *variable
	}{
		{
			name: "create new variable",
			args: args{
				name: "var1",
			},
			want: &variable{
				name:        "var1",
				self:        &variableItem{},
				ttl:         0,
				ttlVersion:  0,
				remoteItems: map[string]*variableItem{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newVariable(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}
