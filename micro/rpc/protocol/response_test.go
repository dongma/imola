package protocol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResponseEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestId:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Error:      []byte("this is error"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				RequestId:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no data",
			resp: &Response{
				RequestId:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Error:      []byte("this is error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalculateHeaderLength()
			tc.resp.CalculateBodyLength()
			data := EncodeResp(tc.resp)
			resp := DecodeResp(data)
			assert.Equal(t, tc.resp, resp)
		})
	}
}
