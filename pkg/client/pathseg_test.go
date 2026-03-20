package client

import "testing"

func TestJobPathSegment(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "leading zero in first block",
			in:   "dcdcbf4-3a0b-4591-96ff-f715b6582631",
			want: "0dcdcbf4-3a0b-4591-96ff-f715b6582631",
		},
		{
			name: "standard UUID unchanged",
			in:   "01234567-89ab-cdef-0123-456789abcdef",
			want: "01234567-89ab-cdef-0123-456789abcdef",
		},
		{
			name: "en dash U+2013 as first separator",
			in:   "dcdcbf4\u20133a0b-4591-96ff-f715b6582631",
			want: "0dcdcbf4-3a0b-4591-96ff-f715b6582631",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobPathSegment(tt.in); got != tt.want {
				t.Fatalf("JobPathSegment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
