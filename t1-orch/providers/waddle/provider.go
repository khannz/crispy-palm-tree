package waddle

import (
	"context"

	waddleService "github.com/khannz/crispy-palm-tree/t1-orch/external-api/waddle"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

//WithCallOpts call opts
type WithCallOpts struct {
	providers.GetOption
	CallOptions []grpc.CallOption
}

//WithNodeIP with node IP
type WithNodeIP struct {
	providers.GetOption
	NodeIp string
}

//ServiceConfigData ...
type ServiceConfigData struct {
	providers.ServicesConfig `json:"-"`
	waddleService.GetServicesResponse
}

//ProviderConfig ...
type ProviderConfig struct {
	DefaultNodeIP       string
	WaddleServiceClient waddleService.NodeInfoServiceClient //owned
}

//NewServiceConfigProvider new provider
func NewServiceConfigProvider(conf ProviderConfig) (providers.ServicesConfigProvider, error) {
	return &providerImpl{ProviderConfig: conf}, nil
}

//---------------------------------------------------- IMPL ---------------------------------------------------

type providerImpl struct {
	ProviderConfig
}

//Close io.Closer
func (impl *providerImpl) Close() error {
	return impl.WaddleServiceClient.Close()
}

func (impl *providerImpl) Get(ctx context.Context, opts ...providers.GetOption) (providers.ServicesConfig, error) {
	const api = "waddle/ServiceConfigProvide/Get"

	var callOpts []grpc.CallOption
	var req waddleService.GetServicesRequest
	req.Node = &waddleService.Node{
		NodeIp: impl.DefaultNodeIP,
	}
	for _, o := range opts {
		switch t := o.(type) {
		case WithNodeIP:
			req.Node.NodeIp = t.NodeIp
		case WithCallOpts:
			callOpts = append(callOpts, t.CallOptions...)
		}
	}
	resp, err := impl.WaddleServiceClient.GetServices(ctx, req, callOpts...)
	return ServiceConfigData{GetServicesResponse: resp}, errors.Wrap(err, api)
}
