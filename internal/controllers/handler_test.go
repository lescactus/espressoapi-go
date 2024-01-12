package controllers

import (
	"reflect"
	"testing"

	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		sheetService         sheet.Service
		serverMaxRequestSize int64
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, 0},
			want: &Handler{nil, 0},
		},
		{
			name: "non nil args",
			args: args{sheet.New(nil), 10},
			want: &Handler{sheet.New(nil), 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.sheetService, tt.args.serverMaxRequestSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
