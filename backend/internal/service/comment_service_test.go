package service

import "testing"

func TestNormalizePageSize(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 0, want: 10},
		{input: -1, want: 10},
		{input: 1, want: 1},
		{input: 50, want: 50},
		{input: 51, want: 50},
		{input: 200, want: 50},
	}

	for _, tt := range tests {
		if got := normalizePageSize(tt.input); got != tt.want {
			t.Errorf("normalizePageSize(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
