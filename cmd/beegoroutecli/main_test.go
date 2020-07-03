package main

import (
	"testing"
)

func Test_autoFunc(t *testing.T) {
	type args struct {
		obj      string
		funcName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				obj:      "DefaultVersionController",
				funcName: "Get",
			},
			want: "GetVersion",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := autoFunc(tt.args.obj, tt.args.funcName); got != tt.want {
				t.Errorf("autoFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
