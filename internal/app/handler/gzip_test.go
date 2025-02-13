package handler

import "testing"

func Test_parseHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		value  string
		want   bool
	}{
		{
			name:   "Test 1",
			header: "gzip, deflate, br, zstd",
			value:  "gzip",
			want:   true,
		},
		{
			name:   "Test 2",
			header: "deflate, br, zstd",
			value:  "gzip",
			want:   false,
		},
		{
			name:   "Test 3",
			header: "",
			value:  "gzip",
			want:   false,
		},
		{
			name:   "Test 4",
			header: "gzip,deflate,br,zstd",
			value:  "gzip",
			want:   true,
		},
		{
			name:   "Test 5",
			header: "gzip",
			value:  "gzip",
			want:   true,
		},
		{
			name:   "Test 6",
			header: "deflate",
			value:  "gzip",
			want:   false,
		},
		{
			name:   "Test 7",
			header: "gzip,deflate,br,zstd,,,",
			value:  "gzip",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseHeader(tt.header, tt.value); got != tt.want {
				t.Errorf("parseHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
