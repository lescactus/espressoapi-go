package controllers

import (
	"reflect"
	"testing"

	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		sheetService sheet.Service
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil},
			want: &Handler{nil},
		},
		{
			name: "non nil args",
			args: args{sheet.New(nil)},
			want: &Handler{sheet.New(nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.sheetService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
