package portadapter

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	transport "github.com/khannz/crispy-palm-tree/t1-orch/grpc-healthcheck"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const healthcheckCheckerName = "healthcheck-checker"

type HealthcheckChecker struct {
	address     string
	grpcTimeout time.Duration // TODO: somehow use tickers?
	logging     *logrus.Logger
}

func NewHealthcheckChecker(address string, grpcTimeout time.Duration, logging *logrus.Logger) *HealthcheckChecker {
	return &HealthcheckChecker{
		address:     address,
		grpcTimeout: grpcTimeout,
		logging:     logging,
	}
}

func (healthcheckChecker *HealthcheckChecker) IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	withContextDialer := makeDialer(healthcheckChecker.address, 20*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckChecker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't connect grpc uds: %v", err)
		return false
	}
	defer conn.Close()

	dummyClient := transport.NewHealthcheckWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbTcpData := &transport.TcpData{
		HealthcheckAddress: healthcheckAddress,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}
	isTcpCheckOk, err := dummyClient.IsTcpCheckOk(sendCtx, pbTcpData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isTcpCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	withContextDialer := makeDialer(healthcheckChecker.address, 20*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckChecker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't connect grpc uds: %v", err)
		return false
	}
	defer conn.Close()

	dummyClient := transport.NewHealthcheckWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbHttpsData := &transport.HttpsData{
		HealthcheckAddress: healthcheckAddress,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}

	isHttpsCheckOk, err := dummyClient.IsHttpsCheckOk(sendCtx, pbHttpsData)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isHttpsCheckOk.GetIsOk()
}

func (healthcheckChecker *HealthcheckChecker) IsHttpAdvancedCheckOk(hcType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	withContextDialer := makeDialer(healthcheckChecker.address, 20*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckChecker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't connect grpc uds: %v", err)
		return false
	}
	defer conn.Close()

	dummyClient := transport.NewHealthcheckWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbHttpAdvandecData := &transport.HttpAdvancedData{
		HcType:             hcType,
		HealthcheckAddress: healthcheckAddress,
		NearFieldsMode:     nearFieldsMode,
		UserDefinedData:    userDefinedData,
		Timeout:            ptypes.DurationProto(timeout),
		Fwmark:             int64(fwmark),
		Id:                 id,
	}

	isHttpAdvancedCheckOk, err := dummyClient.IsHttpAdvancedCheckOk(sendCtx, pbHttpAdvandecData)
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
	withContextDialer := makeDialer(healthcheckChecker.address, 20*time.Second)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer dialCancel()
	conn, err := grpc.DialContext(dialCtx, healthcheckChecker.address, grpc.WithInsecure(), grpc.WithContextDialer(withContextDialer))
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't connect grpc uds: %v", err)
		return false
	}
	defer conn.Close()

	dummyClient := transport.NewHealthcheckWorkerClient(conn)
	sendCtx, sendCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer sendCancel()

	pbIcmp := &transport.IcmpData{
		IpS:     ipS,
		Timeout: ptypes.DurationProto(timeout),
		Fwmark:  int64(fwmark),
		Id:      id,
	}
	isIcmpCheckOk, err := dummyClient.IsIcmpCheckOk(sendCtx, pbIcmp)
	if err != nil {
		healthcheckChecker.logging.WithFields(logrus.Fields{
			"entity":   healthcheckCheckerName,
			"event id": id,
		}).Errorf("can't got error when get response from grpc uds: %v", err)
	}
	return isIcmpCheckOk.GetIsOk()
}
