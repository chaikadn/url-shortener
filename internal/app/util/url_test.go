package util

import "testing"

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   bool
	}{
		{
			name:   "positive test",
			rawURL: "https://practicum.yandex.ru/",
			want:   true,
		},
		{
			name:   "invalid url test",
			rawURL: "bad-url",
			want:   false,
		},
		{
			name:   "empty url test",
			rawURL: "",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidURL(tt.rawURL); got != tt.want {
				t.Errorf("IsValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
