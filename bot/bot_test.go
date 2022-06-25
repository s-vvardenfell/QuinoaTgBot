package bot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_checkYear(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "Lower val",
			arg:  "1900",
			want: true,
		},
		{
			name: "Upper val",
			arg:  "2023",
			want: true,
		},
		{
			name: "Char / string",
			arg:  "2023abc",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkYear(tt.arg); got != tt.want {
				require.Equal(t, got, tt.want)
			}
		})
	}
}
