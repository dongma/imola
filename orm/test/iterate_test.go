package test

import (
	"github.com/stretchr/testify/assert"
	"imola/orm/reflect"
	"testing"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCases := []struct {
		Name     string
		entity   any
		wantVals []any
		wantErr  error
	}{
		{
			Name:     "[]int",
			entity:   [3]int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
		{
			Name:     "slice",
			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			vals, err := reflect.IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantVals, vals)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, vals)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		Name     string
		entity   any
		wantKeys []any
		wantVals []any
		wantErr  error
	}{
		{
			Name: "map",
			entity: map[string]string{
				"A": "a",
				"B": "b",
			},
			wantKeys: []any{"A", "B"},
			wantVals: []any{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			keys, vals, err := reflect.IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantKeys, keys)
			assert.EqualValues(t, tc.wantVals, vals)
		})
	}
}
