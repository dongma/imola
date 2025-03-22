package test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"imola/micro/rpc"
	"imola/micro/rpc/compresser"
	"imola/micro/rpc/serialize/json"
	"testing"
)

func Test_setFuncFields(t *testing.T) {
	testCases := []struct {
		name    string
		service rpc.Service
		mock    func(ctrl *gomock.Controller) rpc.Proxy
		wantErr error
	}{
		{
			name:    "nil",
			service: nil,
			mock: func(ctrl *gomock.Controller) rpc.Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 不支持nil"),
		},
		{
			name:    "no pointer",
			service: rpc.UserService{},
			mock: func(ctrl *gomock.Controller) rpc.Proxy {
				return NewMockProxy(ctrl)
			},
			wantErr: errors.New("rpc: 只支持指向结构体的一级指针"),
		},
		/*{
			name: "user service",
			mock: func(ctrl *gomock.Controller) rpc.Proxy {
				proxy := NewMockProxy(ctrl)
				proxy.EXPECT().Invoke(gomock.Any(), &protocol.Request{
					ServiceName: "user-service",
					MethodName:  "GetById",
					Data:        []byte(`{"Id":123}`),
				}).Return(&protocol.Response{}, nil)
				return proxy
			},
			service: &rpc.UserService{},
		},*/
	}

	s := &json.Serializer{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := rpc.SetFuncField(tc.service, tc.mock(ctrl), s, compresser.DoNothing{})
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			resp, err := tc.service.(*rpc.UserService).GetById(context.Background(), &rpc.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			t.Log(resp)
		})
	}

}
