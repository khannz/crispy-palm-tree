package waddle

import (
	"context"

	"github.com/gradusp/protos/pkg/waddle"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type (
	//Node alias
	Node = waddle.Node

	//GetServicesRequest request
	GetServicesRequest struct {
		*Node
	}

	//GetServicesResponse response
	GetServicesResponse struct {
		*waddle.ServicesResponse
	}

	//NodeInfoServiceClient ...
	NodeInfoServiceClient interface {
		Close() error
		GetServices(ctx context.Context, in GetServicesRequest, opts ...grpc.CallOption) (GetServicesResponse, error)
	}
)

//NewNodeInfoServiceClient ...
func NewNodeInfoServiceClient(_ context.Context, conn *grpc.ClientConn) (NodeInfoServiceClient, error) {
	ret := &nodeInfoService{
		NodeServiceClient: waddle.NewNodeServiceClient(conn),
		closer: func() error {
			return conn.Close()
		},
	}
	return ret, nil
}

var _ NodeInfoServiceClient = (*nodeInfoService)(nil)

type nodeInfoService struct {
	waddle.NodeServiceClient
	closer func() error
}

//Close io.Closer
func (client *nodeInfoService) Close() error {
	return client.closer()
}

func (client *nodeInfoService) GetServices(ctx context.Context, req GetServicesRequest, opts ...grpc.CallOption) (GetServicesResponse, error) {
	const api = "NodeInfoClient/GetServices"
	var resp GetServicesResponse
	r, e := client.NodeServiceClient.GetServices(ctx, req.Node, opts...)
	if e == nil {
		resp.ServicesResponse = r
	}
	return resp, errors.Wrap(e, api)
}
