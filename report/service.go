package report

import (
	"TL-Gateway/config"
	"TL-Gateway/model"
	"TL-Gateway/proto/gateway"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	gcache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalFusionErrorCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_fusion_error_coutn",
		Help: "The total count of fusion error from server",
	})

	totalRedisErrorCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_redis_error_count",
		Help: "The total count of redis error from server",
	})

	totalCollectedCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_collected_count",
		Help: "The total count of messages collected by grpc server",
	})

	totalRefusedCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_refused_count",
		Help: "The total count of messages refused by fusion auth",
	})

	totalTimeoutCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_timeout_count",
		Help: "The total count of messages writeing timeout",
	})
)

func init() {
	prometheus.MustRegister(totalFusionErrorCount)
	prometheus.MustRegister(totalRedisErrorCount)
	prometheus.MustRegister(totalCollectedCount)
	prometheus.MustRegister(totalRefusedCount)
	prometheus.MustRegister(totalTimeoutCount)
}

const (
	// DefaultFustionTimeout defines the default timeout for http request every time
	DefaultFustionTimeout = 5
	// DefaultStreamBuffer - queue size for goroutines which is responsible for sending kafka(default 1M)
	DefaultStreamBuffer = 1024 * 1024
	// DefaultWriteTimeout - the timeout for writing the messages to queue
	DefaultWriteTimeout = 2
	// DefaultCacheExpireTime - the default expire time for local cache
	DefaultCacheExpireTime = 30
	// DefaultCacheCleanTime - the default clean time for local cache
	DefaultCacheCleanTime = 60
)

// Service represents the interfaces for gateway
type Service struct {
	gateway.UnimplementedReportServiceServer

	cache        *gcache.Cache                // local cache for token
	fusionClient *fusionauth.FusionAuthClient // fusion client for validate the token
	settings     *config.Config               // settings for the gateway
	stream       chan model.Carrier           // stream for buffering messages
}

// NewService create the report service
func NewService(settings *config.Config) *Service {
	if settings.Cache.Buffer == 0 {
		settings.Cache.Buffer = DefaultStreamBuffer
	}

	s := &Service{settings: settings}

	// init the local cache
	s.cache = gcache.New(DefaultCacheExpireTime, DefaultCacheCleanTime)

	// init the fusion client
	s.fusionClient = s.newFusionClient()
	// init the buffer queue
	s.stream = make(chan model.Carrier, s.settings.Cache.Buffer)

	return s
}

// new a fusion client
func (s *Service) newFusionClient() *fusionauth.FusionAuthClient {
	client := &http.Client{
		Timeout: s.settings.Fusion.Timeout * time.Second,
	}

	// parse the host from url
	host, err := url.Parse(s.settings.Fusion.URL)
	if err != nil {
		panic(err)
	}

	return fusionauth.NewClient(client, host, s.settings.Fusion.APIKey)
}

// md5 the token
func (s *Service) md5Sum(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

// validate the token for every requests
func (s *Service) validate(token string) (bool, error) {
	// md5 the token
	md5Token := s.md5Sum(token)

	// check if the md5 token is existed in local cache
	if _, existed := s.cache.Get(md5Token); existed {
		// if the token is existed
		return true, nil
	}

	// if the key is not existed, check the token from fusion auth server
	verify, err := s.fusionClient.ValidateJWT(token)
	if err != nil {
		return true, err
	}
	// access token is not valid
	if verify.StatusCode == 401 {
		return false, nil
	}
	if verify.StatusCode == 500 {
		// update the metric: fusion error count
		totalFusionErrorCount.Add(1)
		return true, nil
	}

	expire := time.Duration(verify.Jwt.Exp - time.Now().Unix() + 1)
	// set the token to local cache
	s.cache.Set(md5Token, "ok", expire*time.Second)

	return true, nil
}

// Report implements gateway.ReportService
func (s *Service) Report(ctx context.Context, request *gateway.ReportRequest) (*gateway.ReportReply, error) {
	// validate the token
	ok, err := s.validate(request.Token)
	if err != nil {
		// just log the error
		fmt.Printf("validate token: %v\n", err)
	}
	if !ok {
		// update the refused metric
		totalRefusedCount.Add(1)
		return &gateway.ReportReply{Status: gateway.Status_Refused}, nil
	}

	timeout := s.settings.Server.Timeout
	// send data to buffer queue
	select {
	case s.stream <- request.Data:
		// update the collected metric
		totalCollectedCount.Add(1)

	case <-time.After(time.Second * timeout):
		// update the timeout metric
		totalTimeoutCount.Add(1)

		return &gateway.ReportReply{Status: gateway.Status_Timeout}, nil
	}

	return &gateway.ReportReply{Status: gateway.Status_Success}, nil
}

// ReadMessages returns the messages from channel
func (s *Service) ReadMessages() <-chan model.Carrier {
	return s.stream
}
