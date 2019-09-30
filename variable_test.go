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

func TestVariableGet(t *testing.T) {
	type fields struct {
		name        string
		self        *variableItem
		ttl         int64
		ttlVersion  int64
		remoteItems map[string]*variableItem
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "get value with only self data",
			fields: fields{
				name:        "var1",
				self:        &variableItem{val: 100},
				ttl:         0,
				ttlVersion:  0,
				remoteItems: map[string]*variableItem{},
			},
			want: 100,
		},
		{
			name: "get value with self and remote data",
			fields: fields{
				name:       "var1",
				self:       &variableItem{val: 100},
				ttl:        0,
				ttlVersion: 0,
				remoteItems: map[string]*variableItem{
					"node2": {val: 200},
					"node3": {val: 300},
				},
			},
			want: 600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &variable{
				name:        tt.fields.name,
				self:        tt.fields.self,
				ttl:         tt.fields.ttl,
				ttlVersion:  tt.fields.ttlVersion,
				remoteItems: tt.fields.remoteItems,
			}
			if got := v.get(); got != tt.want {
				t.Errorf("get() = %v, want %v", got, tt.want)
			}
		})
	}
}
