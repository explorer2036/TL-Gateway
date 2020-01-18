package report

import (
	"TL-Gateway/config"
	"TL-Gateway/log"
	"TL-Gateway/model"
	"TL-Gateway/proto/gateway"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/peer"
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

	redisClient  *redis.Client                // redis for token cache
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

	// init the redis connection for token cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     settings.Redis.Addr,
		Password: settings.Redis.Passwd, // no password
		DB:       settings.Redis.DB,     // use default DB
	})
	// try to ping the redis
	if _, err := redisClient.Ping().Result(); err != nil {
		panic(err)
	}
	s.redisClient = redisClient

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

// validate the token with the redis cache
func (s *Service) validateByCache(k string, v string) bool {
	// check if the key is existed in redis cache
	token, err := s.redisClient.Get(k).Result()
	if err != nil {
		log.Errorf("redis get %s: %v", k, err)
		return false
	}
	// check if the value is equal
	if token == v {
		return true
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
		return 0, fmt.Errorf("status code: %d", verify.StatusCode)
	}

	return verify.Jwt.Exp, nil
}

// update the new token into the redis cache
func (s *Service) updateCache(k string, v string, expire int64) {
	secs := expire - time.Now().Unix() - 60
	if secs > 0 {
		// set the token to redis cache
		result := s.redisClient.Set(k, v, time.Duration(secs)*time.Second)
		if result.Err() != nil {
			log.Errorf("update the token %s: %v", v, result.Err())
		}
	}
}

// retrieve the peer address
func (s *Service) peerAddr(ctx context.Context) string {
	v, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	if v.Addr == net.Addr(nil) {
		return ""
	}
	addr := v.Addr.String()

	return strings.Split(addr, ":")[0]
}

// login with the request fields
func (s *Service) login(ctx context.Context, request *gateway.LoginRequest) (*fusionauth.LoginResponse, error) {
	// prepare the login request
	var loginRequest fusionauth.LoginRequest
	loginRequest.LoginId = request.LoginId
	loginRequest.Password = request.Password
	loginRequest.IpAddress = s.peerAddr(ctx)
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
	response, err := s.login(ctx, request)
	if err != nil {
		return &gateway.LoginReply{Status: gateway.Status_Refused, Message: err.Error()}, nil
	}

	return &gateway.LoginReply{
		Status: gateway.Status_Success,
		Token:  response.Token,
	}, nil
}

// validate the token with login and application id
func (s *Service) validate(request *gateway.ReportRequest) error {
	// the login + application as key
	k := request.LoginId + request.ApplicationId
	// the token as value
	v := request.Token

	// validate the token by cache
	if ok := s.validateByCache(k, v); ok {
		return nil
	}

	// validate the token by fusion
	expire, err := s.validateByFusion(v)
	if err != nil {
		return err
	}

	// update the local cache for the token
	s.updateCache(k, v, expire)

	return nil
}

// Report implements gateway.Service
func (s *Service) Report(ctx context.Context, request *gateway.ReportRequest) (*gateway.ReportReply, error) {
	// validate the token with login and application id
	if err := s.validate(request); err != nil {
		// update the metric: fusion refused count
		totalRefusedCount.Add(1)

		return &gateway.ReportReply{Status: gateway.Status_Refused, Message: err.Error()}, nil
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
