package portadapter

import (
	context "context"
	"fmt"
	"net"
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	transport "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/grpc-transport"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

const protocolUdsName = "unix"

type TunnelEntity struct {
	sockAddr        string
	grpcTimeout     time.Duration // TODO: somehow use tickers?
	conn            *grpc.ClientConn
	tunCreateClient transport.TunnelsCreateClient
	tunRemoveClient transport.TunnelsRemoveClient
	logging         *logrus.Logger
}

func NewTunnelEntity(sockAddr string, grpcTimeout time.Duration, logging *logrus.Logger) *TunnelEntity {
	return &TunnelEntity{sockAddr: sockAddr, logging: logging}
}

func (tun *TunnelEntity) initGRPC() error {
	dialer := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial(protocolUdsName, addr)
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tun.conn, err = grpc.DialContext(ctx, tun.sockAddr, grpc.WithInsecure(), grpc.WithDialer(dialer)) // TODO: use context dialer
	if err != nil {
		return fmt.Errorf("did not connect to grpc server: %v", err)
	}
	tun.tunCreateClient = transport.NewTunnelsCreateClient(tun.conn)
	tun.tunRemoveClient = transport.NewTunnelsRemoveClient(tun.conn)

	return nil
}

func (tun *TunnelEntity) ConnectToTunnel() error { // FIXME: do not connect for UDS
	return tun.initGRPC()
}

func (tun *TunnelEntity) DisconnectFromTunnel() {
	if err := tun.conn.Close(); err != nil {
		tun.logging.Errorf("close grpc connection to tun error: %v", err)
	}
}

func (tun *TunnelEntity) CreateTunnels(tunnelInfo []*domain.TunnelForApplicationServer, id string) ([]*domain.TunnelForApplicationServer, error) {
	outPbTunInfo := convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelInfo, id)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	updatedPbTunnelInfo, err := tun.tunCreateClient.CreateTunnels(ctx, outPbTunInfo)
	if err != nil {
		return tunnelInfo, fmt.Errorf("can't create tunnels, tun service error: %v", err)
	}

	updatedTunnelInfo := convertIncomePbTunnelsInfoToDomainTunnelsInfo(updatedPbTunnelInfo)
	return updatedTunnelInfo, nil
}

func (tun *TunnelEntity) RemoveTunnels(tunnelInfo []*domain.TunnelForApplicationServer, id string) ([]*domain.TunnelForApplicationServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	outPbTunInfo := convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelInfo, id)
	updatedPbTunnelInfo, err := tun.tunRemoveClient.RemoveTunnels(ctx, outPbTunInfo)
	if err != nil {
		return tunnelInfo, fmt.Errorf("can't remove tunnels, tun service error: %v", err)
	}

	updatedTunnelInfo := convertIncomePbTunnelsInfoToDomainTunnelsInfo(updatedPbTunnelInfo)
	return updatedTunnelInfo, nil
}

func (tun *TunnelEntity) RemoveAllTunnels(tunnelInfo []*domain.TunnelForApplicationServer, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	outPbTunInfo := convertDomainTunnelsInfoToIncomePbTunnelsInfo(tunnelInfo, id)
	_, err := tun.tunRemoveClient.RemoveTunnels(ctx, outPbTunInfo)

	if err != nil {
		return fmt.Errorf("can't remove all tunnels, tun service error: %v", err)
	}

	return nil
}
