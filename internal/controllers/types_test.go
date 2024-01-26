package controllers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestRoastDateMarshalJSON(t *testing.T) {
	rd := RoastDate(time.Now())

	want, err := json.Marshal(time.Time(rd))
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	got, err := rd.MarshalJSON()
	if err != nil {
		t.Fatalf("RoastDate.MarshalJSON error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RoastDate.MarshalJSON() = %s, want %s", got, want)
	}
}
func TestRoastDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    RoastDate
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			input:   []byte("\"2022-12-31\""),
			want:    RoastDate(time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   []byte("\"invalid-date\""),
			want:    RoastDate(time.Time{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rd RoastDate
			err := rd.UnmarshalJSON(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RoastDate.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(rd, tt.want) {
				t.Errorf("RoastDate.UnmarshalJSON() = %v, want %v", rd, tt.want)
			}
		})
	}
}
