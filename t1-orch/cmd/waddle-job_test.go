package run

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	appJobs "github.com/gradusp/go-platform/app/jobs"
	"github.com/gradusp/go-platform/logger"
	"github.com/gradusp/go-platform/pkg/backoff"
	pkgNet "github.com/gradusp/go-platform/pkg/net"
	"github.com/gradusp/go-platform/pkg/patterns/observer"
	"github.com/gradusp/go-platform/pkg/scheduler"
	"github.com/gradusp/go-platform/pkg/tm"
	"github.com/gradusp/go-platform/server"
	waddleService "github.com/gradusp/protos/pkg/waddle"
	"github.com/khannz/crispy-palm-tree/t1-orch/application/jobs/consumers"
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	//waddleServerMock mock-waddle сервис
	waddleServerMock struct {
		mock.Mock
		waddleService.UnimplementedNodeInfoServer
	}

	//facadeMock mock-фасад Orch
	facadeMock struct {
		mock.Mock
	}
)

var _ server.APIService = (*waddleServerMock)(nil)

var payload = `
{
  "services": [
    {
      "reals": [
        {
          "id": "111",
          "address": "2.2.2.2",
          "port": 443
        },
        {
          "id": "222",
          "address": "3.3.3.3",
          "port": 443
        }
      ],
      "healthchecks": [
        {
          "id": "dc74f217-d59a-457b-be0f-96d8d609e296",
          "hello_timer": 1000,
          "response_timer": 800,
          "alive_threshold": 5,
          "dead_threshold": 3,
          "quorum": 1,
          "hysteresis": 5,
          "address": "1.1.1.1",
          "icmp": {
            "packet_size": 512
          }
        },
        {
          "id": "dc74f217-d59a-457b-be0f-96d8d609e297",
          "hello_timer": 1000,
          "response_timer": 800,
          "alive_threshold": 5,
          "dead_threshold": 3,
          "quorum": 1,
          "hysteresis": 5,
          "address": "8.8.8.8",
          "tcp": {
            "port": 443
          }
        },
        {
          "id": "dc74f217-d59a-457b-be0f-96d8d609e298",
          "hello_timer": 1000,
          "response_timer": 800,
          "alive_threshold": 5,
          "dead_threshold": 3,
          "quorum": 1,
          "hysteresis": 5,
          "address": "1.1.2.2",
          "http": {
            "response_codes": [
              200,
              201,
              302
            ],
            "port": 443,
            "uri": "/healthz"
          }
        },
        {
          "id": "dc74f217-d59a-457b-be0f-96d8d609e298",
          "hello_timer": 1000,
          "response_timer": 800,
          "alive_threshold": 5,
          "dead_threshold": 3,
          "quorum": 1,
          "hysteresis": 5,
          "address": "1.1.2.2",
          "https": {
            "response_codes": [
              200,
              201,
              302
            ],
            "port": 443,
            "uri": "/healthz",
            "ssl_verify": false
          }
        }
      ],
      "id": "057cb217-ce30-49ed-a72a-b2466a64de52",
      "protocol": "TCP",
      "address": "10.46.220.129",
      "port": 443,
      "routing_type": "TUN_GRE",
      "balancing_type": "MAGLEV_HASH_PORT"
    }
  ]
}
`

var _ consumers.FacadeInterface = (*facadeMock)(nil)

//ApplyNewConfig impl consumers.FacadeInterface
func (fm *facadeMock) ApplyNewConfig(a0 domain.ServiceInfoConf) error {
	ret := fm.Called(a0)
	type f = func(domain.ServiceInfoConf) error
	if rf, ok := ret.Get(0).(f); ok {
		return rf(a0)
	}
	return nil
}

//RemoveAllConfigs impl consumers.FacadeInterface
func (fm *facadeMock) RemoveAllConfigs() error {
	ret := fm.Called()
	type f = func() error
	if rf, ok := ret.Get(0).(f); ok {
		return rf()
	}
	return nil
}

//GetServices ...
func (srv *waddleServerMock) GetServices(_a0 context.Context, _a1 *waddleService.Node) (*waddleService.ServicesResponse, error) {
	ret := srv.Called(_a0, _a1)
	var (
		r0 *waddleService.ServicesResponse
		r1 error
	)
	type f = func(context.Context, *waddleService.Node) (*waddleService.ServicesResponse, error)
	if rf, ok := ret.Get(0).(f); ok {
		r0, r1 = rf(_a0, _a1)
	}
	return r0, r1
}

func (srv *waddleServerMock) Description() grpc.ServiceDesc {
	return waddleService.NodeInfo_ServiceDesc
}

func (srv *waddleServerMock) RegisterGRPC(_ context.Context, grpcServer *grpc.Server) error {
	waddleService.RegisterNodeInfoServer(grpcServer, srv)
	return nil
}

func setupWaddleServer() (*server.APIServer, error) {
	s := new(waddleServerMock)
	s.On("GetServices", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *waddleService.Node) (*waddleService.ServicesResponse, error) {
			res := new(waddleService.ServicesResponse)
			err := jsonpb.UnmarshalString(payload, res)
			return res, err
		})
	return server.NewAPIServer(server.WithServices(s))
}

/*//
тестируем получения данных из нового источника "WADDLE"
1 -- запускаем CronJob которая каждую минуту получент данные из waddle
2 -- ловим данные из CronJon=b  и перенаплавляем в фасад Orch
TODO: все что нижу нужно еще сделать
3 -- транслируем данные из waddle во внутреннюю модель фасада Orch
*/
//Test_WaddleJobWorksGood ...
func Test_WaddleJobWorksGood(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)
	err := os.Setenv("WADDLE_ADDRESS", "tcp://127.0.0.1:7001")
	if !assert.NoError(t, err) {
		return
	}
	err = os.Setenv("WADDLE_NODE_IP", "1.1.1.1")
	if !assert.NoError(t, err) {
		return
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	var endPoint *pkgNet.Endpoint
	endPoint, err = pkgNet.ParseEndpoint(viper.GetString("waddle-address"))
	if !assert.NoError(t, err) {
		return
	}

	//отводим на тест не более 30-ти секунд
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//cron-job работает с 3-х секундным интервалом
	jobRunsInConstInterval := scheduler.NewConstIntervalScheduler(3 * time.Second)
	jc := &jobConstructor{
		ctx:        ctx,
		taskManger: tm.NewTaskManager(),
		scheduler:  jobRunsInConstInterval,
		backoff: backoff.ExponentialBackoffBuilder().
			WithInitialInterval(500 * time.Millisecond).
			WithMaxInterval(5 * time.Minute).
			Build(),
	}

	jc.observers = append(jc.observers,
		observer.NewObserver(
			func(evt observer.EventType) { //подключаем observer для логирования из cron-job
				switch t := evt.(type) {
				case appJobs.OnJobLog:
					logger.Infof(ctx, "%s", t)
				}
			}, false, appJobs.OnJobLog{},
		),
	)

	var job appJobs.JobScheduler
	job, err = jc.constructWaddleJob() //создаем cron-job  для получения данных из waddle
	if !assert.NoError(t, err) {
		return
	}
	defer job.Close()

	facade := new(facadeMock)
	goodResult := make(chan struct{}, 1)
	goodResultCount := 0
	facade.On("ApplyNewConfig", mock.Anything).
		Return(func(_ domain.ServiceInfoConf) error {
			goodResultCount++
			if goodResultCount == 3 {
				select {
				case goodResult <- struct{}{}: //после 3-х кратк=ного полученя наннхв фасадом - решаем что все хорошо - тест успешный
				default:
				}
			}
			return nil
		})

	//создаем consumer для того чтобы данные из cron-job попадали в фасад Orch
	facadeConsumerCloser := consumers.NewFacadeConsumer(job, facade, nil)
	defer facadeConsumerCloser.Close()

	var serv *server.APIServer
	serv, err = setupWaddleServer()
	if !assert.NoError(t, err) {
		return
	}

	job.Schedule() //запускаем cron-job
	job.Enable(true)
	go func() {
		_ = serv.Run(ctx, endPoint) //запускаем waddle mock-сервер
	}()
	select {
	case <-goodResult:
		err = nil
	case <-ctx.Done():
		err = ctx.Err()
	}
	assert.NoError(t, err)
}
