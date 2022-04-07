package rpcutil

import (
	"context"
	"testing"

	"github.com/pingcap/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type (
	Request  struct{}
	Response struct{}
)

var req *Request = nil

type mockRPCClient struct {
	addr string
	cnt  int
}

func (c *mockRPCClient) MockRPC(ctx context.Context, req *Request, opts ...grpc.CallOption) (*Response, error) {
	c.cnt++
	return nil, nil
}

func (c *mockRPCClient) MockFailRPC(ctx context.Context, req *Request, opts ...grpc.CallOption) (*Response, error) {
	c.cnt++
	return nil, errors.New("mock fail")
}

var testClient = &mockRPCClient{}

func mockDail(context.Context, string) (*mockRPCClient, CloseableConnIface, error) {
	return testClient, nil, nil
}

func TestFailoverRpcClients(t *testing.T) {
	ctx := context.Background()
	clients, err := NewFailoverRPCClients(ctx, []string{"url1", "url2"}, mockDail)
	require.NoError(t, err)
	_, err = DoFailoverRPC(ctx, clients, req, (*mockRPCClient).MockRPC)
	require.NoError(t, err)
	require.Equal(t, 1, testClient.cnt)
	require.Len(t, clients.Endpoints(), 2)

	// reset
	testClient.cnt = 0
	_, err = DoFailoverRPC(ctx, clients, req, (*mockRPCClient).MockFailRPC)
	require.Error(t, err)
	require.Equal(t, 2, testClient.cnt)

	clients.UpdateClients(ctx, []string{"url1", "url2", "url3"}, "")
	testClient.cnt = 0
	_, err = DoFailoverRPC(ctx, clients, req, (*mockRPCClient).MockFailRPC)
	require.Error(t, err)
	require.Equal(t, 3, testClient.cnt)
	require.Len(t, clients.Endpoints(), 3)
}

type closer struct{}

func (c *closer) Close() error {
	return nil
}

func mockDailWithAddr(_ context.Context, addr string) (*mockRPCClient, CloseableConnIface, error) {
	return &mockRPCClient{addr: addr}, &closer{}, nil
}

func TestValidLeaderAfterUpdateClients(t *testing.T) {
	ctx := context.Background()
	clients, err := NewFailoverRPCClients(ctx, []string{"url1", "url2"}, mockDailWithAddr)
	require.NoError(t, err)
	require.NotNil(t, clients.GetLeaderClient())

	clients.UpdateClients(ctx, []string{"url1", "url2", "url3"}, "url1")
	require.Equal(t, "url1", clients.GetLeaderClient().addr)

	clients.UpdateClients(ctx, []string{"url1", "url2", "url3"}, "not-in-set")
	require.NotNil(t, clients.GetLeaderClient())

	clients.UpdateClients(ctx, []string{}, "")
	require.Nil(t, clients.GetLeaderClient())
}