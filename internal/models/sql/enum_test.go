package sql

import "testing"

func TestComparisonWithPreviousResultIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value ComparisonWithPreviousResult
		want  bool
	}{
		{name: "worst", value: Worst, want: true},
		{name: "same", value: Same, want: true},
		{name: "better", value: Better, want: true},
		{name: "unknown", value: Unknown, want: true},
		{name: "above range", value: 4, want: false},
		{name: "maximum uint8", value: 255, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("ComparisonWithPreviousResult.IsValid() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRoastLevelIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value RoastLevel
		want  bool
	}{
		{name: "light", value: RoastLevelLight, want: true},
		{name: "light to medium", value: RoastLevelLightToMedium, want: true},
		{name: "medium", value: RoastLevelMedium, want: true},
		{name: "medium to dark", value: RoastLevelMediumToDark, want: true},
		{name: "dark", value: RoastLevelDark, want: true},
		{name: "above range", value: 5, want: false},
		{name: "maximum uint8", value: 255, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("RoastLevel.IsValid() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRoastLevelString(t *testing.T) {
	tests := []struct {
		name  string
		value RoastLevel
		want  string
	}{
		{name: "light", value: RoastLevelLight, want: "Light"},
		{name: "light to medium", value: RoastLevelLightToMedium, want: "Light to medium"},
		{name: "medium", value: RoastLevelMedium, want: "Medium"},
		{name: "medium to dark", value: RoastLevelMediumToDark, want: "Medium to dark"},
		{name: "dark", value: RoastLevelDark, want: "Dark"},
		{name: "invalid", value: 5, want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.String(); got != tt.want {
				t.Errorf("RoastLevel.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComparisonWithPreviousResultString(t *testing.T) {
	tests := []struct {
		name  string
		value ComparisonWithPreviousResult
		want  string
	}{
		{name: "worst", value: Worst, want: "Worse"},
		{name: "same", value: Same, want: "Same"},
		{name: "better", value: Better, want: "Better"},
		{name: "unknown", value: Unknown, want: "Unknown"},
		{name: "invalid", value: 4, want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.String(); got != tt.want {
				t.Errorf("ComparisonWithPreviousResult.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
