package portadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-healthcheck"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const healthcheckCheckerName = "healthcheck-checker"

type HealthcheckChecker struct {
	client  transport.HealthcheckWorkerClient
	Conn    *grpc.ClientConn
	logging *logrus.Logger
}

func NewHealthcheckChecker(address string, grpcTimeout time.Duration, logging *logrus.Logger) (*HealthcheckChecker, error) {
	withContextDialer := makeDialer(address, 20*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		return nil, fmt.Errorf("can't not connect to HC %v, dial fail: %v", address, err)
	}
	client := transport.NewHealthcheckWorkerClient(conn)

	return &HealthcheckChecker{
		client:  client,
		Conn:    conn,
		logging: logging,
	}, nil
}

func (healthcheckChecker *HealthcheckChecker) IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbTcpData := &transport.TcpData{
		HealthcheckAddress: healthcheckAddress,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}
	isTcpCheckOk, err := healthcheckChecker.client.IsTcpCheckOk(sendCtx, pbTcpData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isTcpCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsHttpCheckOk(healthcheckAddress string,
	uri string,
	validResponseCodes []int64,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbHttpData := &transport.HttpData{
		HealthcheckAddress: healthcheckAddress,
		Uri:                uri,
		ValidResponseCodes: validResponseCodes,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}

	isHttpCheckOk, err := healthcheckChecker.client.IsHttpCheckOk(sendCtx, pbHttpData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isHttpCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsHttpsCheckOk(healthcheckAddress string,
	uri string,
	validResponseCodes []int64,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbHttpsData := &transport.HttpsData{
		HealthcheckAddress: healthcheckAddress,
		Uri:                uri,
		ValidResponseCodes: validResponseCodes,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}

	isHttpsCheckOk, err := healthcheckChecker.client.IsHttpsCheckOk(sendCtx, pbHttpsData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isHttpsCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsHttpAdvancedCheckOk(healthcheckType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbHttpAdvandecData := &transport.HttpAdvancedData{
		HealthcheckType:    healthcheckType,
		HealthcheckAddress: healthcheckAddress,
		NearFieldsMode:     nearFieldsMode,
		UserDefinedData:    userDefinedData,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}

	isHttpAdvancedCheckOk, err := healthcheckChecker.client.IsHttpAdvancedCheckOk(sendCtx, pbHttpAdvandecData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isHttpAdvancedCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsIcmpCheckOk(ipS string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbIcmp := &transport.IcmpData{
		IpS:     ipS,
		Timeout: ptypes.DurationProto(timeout),
		Fwmark:  int64(fwmark),
		Id:      id,
	}
	isIcmpCheckOk, err := healthcheckChecker.client.IsIcmpCheckOk(sendCtx, pbIcmp)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isIcmpCheckOk.GetIsOk()
}
