package sort

import (
	"testing"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name                        string
		width, height, length, mass int
		want                        Stack
	}{
		{
			name:  "standard package",
			width: 10, height: 10, length: 10, mass: 10,
			want: Standard,
		},
		{
			name:  "bulky package by volume",
			width: 120, height: 120, length: 120, mass: 10,
			want: Special,
		},
		{
			name:  "bulky package by dimension",
			width: 10, height: 10, length: 200, mass: 10,
			want: Special,
		},
		{
			name:  "heavy package",
			width: 10, height: 10, length: 10, mass: 30,
			want: Special,
		},
		{
			name:  "heavy and bulky package by volume",
			width: 120, height: 120, length: 120, mass: 30,
			want: Rejected,
		},
		{
			name:  "heavy and bulky package by dimension",
			width: 200, height: 10, length: 10, mass: 30,
			want: Rejected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sort(tt.width, tt.height, tt.length, tt.mass); got != tt.want {
				t.Errorf("Sort() = %v, want %v", got, tt.want)
			}
		})
	}
}
