package consul

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gradusp/go-platform/pkg/parallel"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
	"github.com/pkg/errors"
)

//ProviderConfig provider config
type ProviderConfig struct {
	ConsulAddress        string
	ConsulSubscribePath  string
	ConsulAppServersPath string
	ServiceManifest      string
}

//NewServiceConfigProvider new provider
func NewServiceConfigProvider(conf ProviderConfig) (providers.ServicesConfigProvider, error) {
	const api = "consul/NewServiceConfigProvider"

	consulConf := consulAPI.DefaultConfig()
	consulConf.Address = conf.ConsulAddress
	client, err := consulAPI.NewClient(consulConf)
	if err != nil {
		return nil, errors.Wrap(err, api)
	}

	return &providerImpl{
		dataAccessor:    client.KV(),
		subscribePath:   conf.ConsulSubscribePath,
		appServersPath:  conf.ConsulAppServersPath,
		serviceManifest: conf.ServiceManifest,
	}, nil
}

type WithWaitIndex struct {
	providers.GetOption
	WaitIndex uint64
}

var (
	_ providers.GetOption = WithWaitIndex{}
)

//--------------------------------------------------- Impl ---------------------------------------------------

type providerImpl struct {
	dataAccessor    *consulAPI.KV
	subscribePath   string
	appServersPath  string
	serviceManifest string
}

//Close impl io.Closer
func (impl *providerImpl) Close() error {
	return nil
}

//Get impl providers.BalancerConfigProvider
func (impl *providerImpl) Get(ctx context.Context, opts ...providers.GetOption) (providers.ServicesConfig, error) {
	const api = "consul/ServiceConfigProvide/Get"

	var queryOptions consulAPI.QueryOptions
	var index2check uint64
	for _, o := range opts {
		switch t := o.(type) {
		case WithWaitIndex:
			index2check, queryOptions.WaitIndex = t.WaitIndex, t.WaitIndex
		}
	}
	balancingServices, meta, err := impl.dataAccessor.Keys(impl.subscribePath, "/", &queryOptions)
	if err != nil {
		return nil, errors.Wrap(ErrConsul{error: err, Op: OpKeys, Path: impl.subscribePath}, api)
	}

	if index2check != 0 && index2check == meta.LastIndex {
		return ServiceTransportData{
			QueryMeta: meta,
			Payload:   NonePayload{},
		}, nil
	}

	var mx sync.Mutex
	var services []ServiceTransport
	err = parallel.ExecAbstract(len(balancingServices), 10, func(i int) error {
		bsPath := balancingServices[i]
		if bsPath == impl.subscribePath {
			return nil
		}
		service, e := impl.GetServiceTransport(ctx, bsPath)
		if e != nil {
			return e
		}
		mx.Lock()
		services = append(services, service)
		mx.Unlock()
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, api)
	}
	return ServiceTransportData{
		QueryMeta: meta,
		Payload:   ServicesPayload{Services: services},
	}, nil
}

//GetServiceTransport ...
func (impl *providerImpl) GetServiceTransport(_ context.Context, bsPath string) (ServiceTransport, error) {
	var ret ServiceTransport

	path := bsPath + impl.serviceManifest
	serviceManifest, _, err := impl.dataAccessor.Get(path, nil)
	if err != nil {
		return ret, ErrConsul{error: err, Op: OpGet, Path: path}
	}
	if serviceManifest == nil {
		return ret, ErrConsul{error: ErrNotFound}
	}
	if err = json.Unmarshal(serviceManifest.Value, &ret); err != nil {
		return ret, errors.Wrapf(ErrDataNotFit2Model, "unmarshall 'ServiceTransport': %v", err)
	}

	path = bsPath + impl.appServersPath
	applicationServersPaths, _, err := impl.dataAccessor.Keys(path, "/", nil)
	if err != nil {
		return ret, ErrConsul{error: err, Op: OpKeys, Path: path}
	}

	for _, applicationServersPath := range applicationServersPaths {
		if applicationServersPath == path {
			continue
		}
		var applicationServerPair *consulAPI.KVPair
		applicationServerPair, _, err = impl.dataAccessor.Get(applicationServersPath, nil)
		if err != nil {
			return ret, ErrConsul{error: err, Op: OpKeys, Path: applicationServersPath}
		}
		if applicationServerPair == nil {
			return ret, ErrConsul{error: ErrNotFound, Op: OpKeys, Path: applicationServersPath}
		}
		var applicationServerTransport ApplicationServerTransport
		if err = json.Unmarshal(applicationServerPair.Value, &applicationServerTransport); err != nil {
			return ret, errors.Wrapf(ErrDataNotFit2Model, "unmarshall 'ApplicationServerTransport': %v", err)
		}
		ret.ApplicationServersTransport =
			append(ret.ApplicationServersTransport, applicationServerTransport)
	}

	return ret, nil
}
