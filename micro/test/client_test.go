package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"imola/micro/net"
	"testing"
	"time"
)

func TestClient_Send(t *testing.T) {
	server := &net.Server{}
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(3 * time.Second)
	client := &net.Client{
		Network: "tcp",
		Addr:    "localhost:8081",
	}
	resp, err := client.Send("hello")
	require.NoError(t, err)
	assert.Equal(t, "hellohello", resp)
}
