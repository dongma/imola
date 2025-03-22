package gzip

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGzipCompresser(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
	}{
		{
			name:  "hello world",
			input: []byte("hello world"),
		},
	}

	zipc := GzipCompresser{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := zipc.Compress(tc.input)
			require.NoError(t, err)
			data, err = zipc.UnCompress(data)
			assert.Equal(t, tc.input, data)
		})
	}
}
