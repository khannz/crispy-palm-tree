package portadapter

import (
	context "context"
	"fmt"
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/grpc-transport"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

type HeathcheckEntity struct {
	address        string
	grpcTimeout    time.Duration // TODO: somehow use tickers?
	conn           *grpc.ClientConn
	hcGetClient    transport.HCGetClient
	hcNewClient    transport.HCNewClient
	hcUpdateClient transport.HCUpdateClient
	logging        *logrus.Logger
}

func NewHeathcheckEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *HeathcheckEntity {
	return &HeathcheckEntity{address: address, logging: logging}
}

func (hc *HeathcheckEntity) initGRPC() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	hc.conn, err = grpc.DialContext(ctx, hc.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	hc.hcGetClient = transport.NewHCGetClient(hc.conn)
	hc.hcNewClient = transport.NewHCNewClient(hc.conn)
	hc.hcUpdateClient = transport.NewHCUpdateClient(hc.conn)

	return nil
}

func (hc *HeathcheckEntity) StartHealthchecksForCurrentServices(servicesInfo []*domain.ServiceInfo, id string) error {
	for _, serviceInfo := range servicesInfo {
		if err := hc.NewServiceToHealtchecks(serviceInfo, id); err != nil {
			return fmt.Errorf("can't start hc for service %v", serviceInfo.Address)
		}
	}
	return nil
}

func (hc *HeathcheckEntity) NewServiceToHealtchecks(serviceInfo *domain.ServiceInfo, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbServiceInfo.Id = id
	if _, err := hc.hcNewClient.HCNewPbService(ctx, pbServiceInfo); err != nil {
		return fmt.Errorf("can't add new service to healtchecks: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(serviceInfo *domain.ServiceInfo, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbServiceInfo.Id = id
	if _, err := hc.hcUpdateClient.HCRemovePbService(ctx, pbServiceInfo); err != nil {
		return fmt.Errorf("can't remove service from healtchecks: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(serviceInfo *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbServiceInfo.Id = id
	pbUpdatedServiceInfo, err := hc.hcUpdateClient.HCUpdatePbService(ctx, pbServiceInfo)
	if err != nil {
		return nil, fmt.Errorf("can't update service in healtchecks: %v", err)
	}
	domainUpdatedServiceInfo := pbServiceToDomainServiceInfo(pbUpdatedServiceInfo)
	return domainUpdatedServiceInfo, nil
}

func (hc *HeathcheckEntity) GetServiceState(serviceInfo *domain.ServiceInfo, id string) (*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbServiceInfo.Id = id
	pbUpdatedServiceInfo, err := hc.hcGetClient.HCGetPbService(ctx, pbServiceInfo)
	if err != nil {
		return nil, fmt.Errorf("can't get service from healtchecks: %v", err)
	}
	domainUpdatedServiceInfo := pbServiceToDomainServiceInfo(pbUpdatedServiceInfo)
	return domainUpdatedServiceInfo, nil
}

func (hc *HeathcheckEntity) GetServicesState(id string) ([]*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pbUpdatedServicesInfo, err := hc.hcGetClient.HCGetPbServiceS(ctx, &transport.EmptyHcData{Id: id})
	if err != nil {
		return nil, fmt.Errorf("can't get services from healtchecks: %v", err)
	}
	domainUpdatedServicesInfo := []*domain.ServiceInfo{}
	for _, pbUpdatedServiceInfo := range pbUpdatedServicesInfo.Services {
		domainUpdatedServicesInfo = append(domainUpdatedServicesInfo, pbServiceToDomainServiceInfo(pbUpdatedServiceInfo))
	}
	return domainUpdatedServicesInfo, nil
}

func (hc *HeathcheckEntity) ConnectToHealtchecks() error {
	return hc.initGRPC()
}
func (hc *HeathcheckEntity) DisconnectFromHealtchecks() {
	if err := hc.conn.Close(); err != nil {
		hc.logging.Errorf("close grpc connection to hc error: %v", err)
	}
}
