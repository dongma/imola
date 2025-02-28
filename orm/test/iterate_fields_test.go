package test

import (
	"imola/orm/unsafe"
	"testing"
)

func TestPrintFieldOffset(t *testing.T) {
	testCases := []struct {
		name   string
		entity any
	}{
		{
			name:   "user",
			entity: UserV1{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unsafe.PrintFieldOffset(tc.entity)
		})
	}

}

type User struct {
	// offset-0
	Name string
	// offset-16
	Age int32
	// offset-24
	AgeV1 int32
	// offset-24
	Alias []string
	// offset-48
	Address string
}

type UserV1 struct {
	// offset-0
	Name string
	// offset-16
	Age   int32
	AgeV1 int32
	// offset-24
	Alias []string
	// offset-48
	Address string
}
