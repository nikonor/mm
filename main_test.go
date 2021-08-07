package main

import (
	"reflect"
	"testing"
)

func Test_splitH(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []string
	}{
		{
			name: "one",
			args: "aaa: bbb",
			want: []string{"aaa", "bbb"},
		},
		{
			name: "two",
			args: "aaa:bbb",
			want: []string{"aaa", "bbb"},
		},
		{
			name: "three",
			args: "aaabbb",
			want: []string{"aaabbb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitH(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitH() = %v, want %v", got, tt.want)
			}
		})
	}
}
