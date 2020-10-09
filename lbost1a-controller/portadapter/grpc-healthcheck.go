package portadapter

import (
	context "context"
	"fmt"
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-controller/domain"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

type HeathcheckEntity struct {
	address        string
	grpcTimeout    time.Duration
	conn           *grpc.ClientConn
	hcGetClient    HCGetClient
	hcNewClient    HCNewClient
	hcUpdateClient HCUpdateClient
	logging        *logrus.Logger
}

func NewHeathcheckEntity(address string, grpcTimeout time.Duration, logging *logrus.Logger) *HeathcheckEntity {
	return &HeathcheckEntity{address: address, logging: logging}
}

func (hc *HeathcheckEntity) initGRPC() error {
	var err error
	hc.conn, err = grpc.Dial(hc.address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	hc.hcGetClient = NewHCGetClient(hc.conn)
	hc.hcNewClient = NewHCNewClient(hc.conn)
	hc.hcUpdateClient = NewHCUpdateClient(hc.conn)

	return nil
}

func (hc *HeathcheckEntity) StartHealthchecksForCurrentServices(servicesInfo []*domain.ServiceInfo) error {
	for _, serviceInfo := range servicesInfo {
		if err := hc.NewServiceToHealtchecks(serviceInfo); err != nil {
			return fmt.Errorf("can't start hc for service %v", serviceInfo.Address)
		}
	}
	return nil
}

func (hc *HeathcheckEntity) NewServiceToHealtchecks(serviceInfo *domain.ServiceInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), hc.grpcTimeout)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	if _, err := hc.hcNewClient.HCNewPbService(ctx, pbServiceInfo); err != nil {
		return fmt.Errorf("can't add new service to healtchecks: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) RemoveServiceFromHealtchecks(serviceInfo *domain.ServiceInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), hc.grpcTimeout)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	if _, err := hc.hcUpdateClient.HCRemovePbService(ctx, pbServiceInfo); err != nil {
		return fmt.Errorf("can't remove service from healtchecks: %v", err)
	}
	return nil
}

func (hc *HeathcheckEntity) UpdateServiceAtHealtchecks(serviceInfo *domain.ServiceInfo) (*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.grpcTimeout)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbUpdatedServiceInfo, err := hc.hcUpdateClient.HCUpdatePbService(ctx, pbServiceInfo)
	if err != nil {
		return nil, fmt.Errorf("can't update service in healtchecks: %v", err)
	}
	domainUpdatedServiceInfo := pbServiceToDomainServiceInfo(pbUpdatedServiceInfo)
	return domainUpdatedServiceInfo, nil
}

func (hc *HeathcheckEntity) GetServiceState(serviceInfo *domain.ServiceInfo) (*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.grpcTimeout)
	defer cancel()
	pbServiceInfo := domainServiceInfoToPbService(serviceInfo)
	pbUpdatedServiceInfo, err := hc.hcGetClient.HCGetPbService(ctx, pbServiceInfo)
	if err != nil {
		return nil, fmt.Errorf("can't get service from healtchecks: %v", err)
	}
	domainUpdatedServiceInfo := pbServiceToDomainServiceInfo(pbUpdatedServiceInfo)
	return domainUpdatedServiceInfo, nil
}

func (hc *HeathcheckEntity) GetServicesState() ([]*domain.ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.grpcTimeout)
	defer cancel()
	pbUpdatedServicesInfo, err := hc.hcGetClient.HCGetPbServiceS(ctx, &EmptyPbService{})
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
