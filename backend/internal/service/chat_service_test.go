package service

import "testing"

func TestSafeLimit(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 0, want: 100},
		{input: -5, want: 100},
		{input: 1, want: 1},
		{input: 50, want: 50},
		{input: 200, want: 200},
		{input: 201, want: 100},
	}

	for _, tt := range tests {
		if got := safeLimit(tt.input); got != tt.want {
			t.Errorf("safeLimit(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
