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
	gateway.UnimplementedServiceServer

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

// validate the token with the local cache
func (s *Service) validateByCache(k string, v string) bool {
	// check if the key is existed in local cache
	if item, existed := s.cache.Get(k); existed {
		// check if the value is valid
		if item == v {
			return true
		}
	}
	return false
}

// validate the token by fusion
func (s *Service) validateByFusion(token string) (int64, error) {
	// validate the token from fusion auth server
	verify, err := s.fusionClient.ValidateJWT(token)
	if err != nil {
		return 0, err
	}
	if verify.StatusCode != 200 {
		return 0, fmt.Errorf("http status code: %d", verify.StatusCode)
	}

	return verify.Jwt.Exp, nil
}

// update the new token into the cache
func (s *Service) updateCache(k string, v string, expire int64) {
	secs := time.Duration(expire - time.Now().Unix() + 1)
	// set the token to local cache
	s.cache.Set(k, v, secs*time.Second)

}

// login with the request fields
func (s *Service) login(request *gateway.LoginRequest) (*fusionauth.LoginResponse, error) {
	// prepare the login request
	var loginRequest fusionauth.LoginRequest
	loginRequest.LoginId = request.LoginId
	loginRequest.Password = request.Password
	loginRequest.IpAddress = request.IpAddress
	loginRequest.ApplicationId = request.ApplicationId

	// login by fusion client
	loginReply, _, err := s.fusionClient.Login(loginRequest)
	if err != nil {
		return nil, err
	}
	// login failed
	if loginReply.StatusCode != 200 {
		return nil, fmt.Errorf("http status code: %v", loginReply.StatusCode)
	}
	return loginReply, nil
}

// Login implements gateway.Service
func (s *Service) Login(ctx context.Context, request *gateway.LoginRequest) (*gateway.LoginReply, error) {
	// login with fusion client
	response, err := s.login(request)
	if err != nil {
		return &gateway.LoginReply{Status: gateway.Status_Refused, Message: err.Error()}, nil
	}

	// login success, validate the token, and retrieve the expire time of the token
	expire, err := s.validateByFusion(response.Token)
	if err != nil {
		return &gateway.LoginReply{Status: gateway.Status_Refused, Message: err.Error()}, nil
	}

	// md5 the login + application as key
	k := s.md5Sum(request.LoginId + request.ApplicationId)
	// md5 the token as value
	v := s.md5Sum(response.Token)

	// update the token into local cache
	s.updateCache(k, v, expire)

	return &gateway.LoginReply{
		Status: gateway.Status_Success,
		Token:  response.Token,
		Expire: expire,
	}, nil
}

// Report implements gateway.Service
func (s *Service) Report(ctx context.Context, request *gateway.ReportRequest) (*gateway.ReportReply, error) {
	// md5 the login + application as key
	k := s.md5Sum(request.LoginId + request.ApplicationId)
	// md5 the token as value
	v := s.md5Sum(request.Token)

	// validate the token by cache
	if ok := s.validateByCache(k, v); ok {
		return &gateway.ReportReply{Status: gateway.Status_Success}, nil
	}
	// validate the token by fusion
	expire, err := s.validateByFusion(request.Token)
	if err != nil {
		// update the metric: fusion refused count
		totalRefusedCount.Add(1)

		return &gateway.ReportReply{Status: gateway.Status_Refused, Message: err.Error()}, nil
	}
	// update the local cache for the token
	s.updateCache(k, v, expire)

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
