package controllers

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRoastDateMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		date RoastDate
		want string
	}{
		{
			name: "UTC date with time",
			date: RoastDate(time.Date(2022, time.December, 31, 23, 59, 58, 0, time.UTC)),
			want: "2022-12-31",
		},
		{
			name: "date with non-UTC offset",
			date: RoastDate(time.Date(2023, time.January, 1, 1, 2, 3, 0, time.FixedZone("UTC+2", 2*60*60))),
			want: "2023-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.MarshalJSON()
			if err != nil {
				t.Fatalf("RoastDate.MarshalJSON error: %v", err)
			}

			if string(got) != `"`+tt.want+`"` {
				t.Errorf("RoastDate.MarshalJSON() = %s, want %q", got, tt.want)
			}
			if got := tt.date.Format(); got != tt.want {
				t.Errorf("RoastDate.Format() = %q, want %q", got, tt.want)
			}
		})
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
			name:    "valid date",
			input:   []byte("\"2022-12-31\""),
			want:    RoastDate(time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "invalid date",
			input:   []byte("\"invalid-date\""),
			want:    RoastDate(time.Time{}),
			wantErr: true,
		},
		{
			name:    "date and time",
			input:   []byte("\"2022-12-31T00:00:00Z\""),
			want:    RoastDate(time.Time{}),
			wantErr: true,
		},
		{
			name:    "unquoted date",
			input:   []byte("2022-12-31"),
			want:    RoastDate(time.Time{}),
			wantErr: true,
		},
		{
			name:    "null",
			input:   []byte("null"),
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

			if !time.Time(rd).Equal(time.Time(tt.want)) {
				t.Errorf("RoastDate.UnmarshalJSON() = %v, want %v", rd, tt.want)
			}
		})
	}
}

func TestRoastDateJSONRoundTrip(t *testing.T) {
	tests := []string{
		`"2022-12-31"`,
		`"2024-02-29"`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var date RoastDate
			if err := json.Unmarshal([]byte(input), &date); err != nil {
				t.Fatalf("json.Unmarshal(%s) error: %v", input, err)
			}

			got, err := json.Marshal(date)
			if err != nil {
				t.Fatalf("json.Marshal error: %v", err)
			}
			if string(got) != input {
				t.Errorf("round trip = %s, want %s", got, input)
			}
		})
	}
}
