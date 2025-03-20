package protocol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				RequestId:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("hello, world"),
			},
		},
		{
			name: "no meta with data",
			req: &Request{
				RequestId:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Data:        []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}
}
