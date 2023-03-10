package internal

import (
	"TL-Gateway/cache"
	"TL-Gateway/config"
	"TL-Gateway/model"
	"TL-Gateway/proto/gateway"
	"TL-Gateway/proto/id"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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
	// GRPCClientDialTimeout - the timeout for client to dial the server
	GRPCClientDialTimeout = 5
	// GRPCClientKeepaliveTime - After a duration of this time if the client doesn't see any activity it
	// pings the server to see if the transport is still alive.
	GRPCClientKeepaliveTime = 15
	// GRPCClientKeepaliveTimeout - After having pinged for keepalive check, the client waits for a duration
	// of Timeout and if no activity is seen even after that the connection is closed.
	GRPCClientKeepaliveTimeout = 5
)

// Service represents the interfaces for gateway
type Service struct {
	gateway.UnimplementedServiceServer

	tokenCache      cache.Cache                  // cache for the token
	idServiceConn   *grpc.ClientConn             // grpc connection for id service
	idServiceClient id.ServiceClient             // id service client
	fusionClient    *fusionauth.FusionAuthClient // fusion client for validate the token
	settings        *config.Config               // settings for the gateway
	stream          chan model.Carrier           // stream for buffering messages
}

// NewService create the internal service
func NewService(settings *config.Config) *Service {
	if settings.Cache.Buffer == 0 {
		settings.Cache.Buffer = DefaultStreamBuffer
	}

	s := &Service{settings: settings}

	// init the cache for token
	if s.settings.Server.Cache == "redis" {
		s.tokenCache = cache.NewRedisCache(settings)
	} else {
		s.tokenCache = cache.NewLocalCache()
	}

	// init the fusion client
	s.newFusionClient()
	// init the id service client
	s.initIDServiceClient()

	// init the buffer queue
	s.stream = make(chan model.Carrier, s.settings.Cache.Buffer)

	return s
}

// init the id service client
func (s *Service) initIDServiceClient() {
	// prepare the dial options for grpc client
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    GRPCClientKeepaliveTime * time.Second,
		Timeout: GRPCClientKeepaliveTimeout * time.Second,
	}))
	opts = append(opts, grpc.WithInsecure())

	// set up a connection to the id server.
	conn, err := grpc.Dial(s.settings.Server.IDService, opts...)
	if err != nil {
		panic(fmt.Sprintf("grpc conn: %v", err))
	}
	// update the connection
	s.idServiceConn = conn

	// new id service client
	s.idServiceClient = id.NewServiceClient(conn)
}

// new a fusion client
func (s *Service) newFusionClient() {
	client := &http.Client{
		Timeout: s.settings.Fusion.Timeout * time.Second,
	}

	// parse the host from url
	host, err := url.Parse(s.settings.Fusion.URL)
	if err != nil {
		panic(err)
	}
	s.fusionClient = fusionauth.NewClient(client, host, s.settings.Fusion.APIKey)
}

// md5 the token
func (s *Service) md5Sum(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
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

	// fetch the source from id service
	result, err := s.idServiceClient.GetSource(ctx, &id.GetSourceRequest{Id: uuid2id(response.User.Id)})
	if err != nil {
		return &gateway.LoginReply{Status: gateway.Status_Error, Message: err.Error()}, nil
	}

	return &gateway.LoginReply{
		Status: gateway.Status_Success,
		Token:  response.Token,
		UserID: response.User.Id,
		Source: result.Source,
	}, nil
}

// dbd62208-d0ea-4a5a-9066-64d41e7dcedd
// 8+4+4+4+12
// for the uuid format, i will put the user id with 10 bytes replace the part "d0ea-4a5a-90"
func uuid2id(uuid string) int32 {
	if len(uuid) != 36 {
		panic(uuid)
	}

	out := strings.Replace(uuid[9:21], "-", "", -1)
	id, err := strconv.Atoi(out)
	if err != nil {
		panic(err)
	}
	return int32(id)
}

// dbd62208-d0ea-4a5a-9066-64d41e7dcedd
// 8+4+4+4+12
// for the uuid format, i 'll put the user id with 10 bytes replace the part "d0ea-4a5a-90"
func id2uuid(id int32) string {
	s := fmt.Sprintf("%010v", id)
	// return fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	uuid := fmt.Sprintf("%08v-%s-%s-%s%02v-%012v",
		rand.Int31n(100000000),
		s[0:4],
		s[4:8],
		s[8:10],
		rand.Int31n(100),
		rand.Int63n(1000000000000),
	)
	return uuid
}

// register the user from fusion auth server
func (s *Service) register(ctx context.Context, request *gateway.RegisterRequest, uuid string) error {
	// prepare the register request
	registrationRequest := fusionauth.RegistrationRequest{
		GenerateAuthenticationToken:  false,
		SendSetPasswordEmail:         false,
		SkipRegistrationVerification: true,
		Registration: fusionauth.UserRegistration{
			ApplicationId: request.ApplicationId,
		},
		User: fusionauth.User{
			SecureIdentity: fusionauth.SecureIdentity{
				Password: request.Password,
			},
			Email: request.LoginId,
		},
	}
	// make the registration with user information
	registrationReply, _, err := s.fusionClient.Register(uuid, registrationRequest)
	if err != nil {
		return err
	}
	// register failed
	if registrationReply.StatusCode != 200 {
		return fmt.Errorf("http status code: %v", registrationReply.StatusCode)
	}
	return nil
}

// Register implements gateway.Service
func (s *Service) Register(ctx context.Context, request *gateway.RegisterRequest) (*gateway.RegisterReply, error) {
	// fetch the user id from id service
	idReply, err := s.idServiceClient.Generate32Bit(ctx, &id.Generate32BitRequest{})
	if err != nil {
		return &gateway.RegisterReply{Status: gateway.Status_Error, Message: err.Error()}, nil
	}

	// format the uuid with an integer
	uuid := id2uuid(idReply.Id)

	// call the register api with user information
	if err := s.register(ctx, request, uuid); err != nil {
		return &gateway.RegisterReply{Status: gateway.Status_Error, Message: err.Error()}, nil
	}

	return &gateway.RegisterReply{Status: gateway.Status_Success}, nil
}

// validate the token with login and application id
func (s *Service) validate(request *gateway.ReportRequest) error {
	// the login + application as key
	k := request.LoginId + request.ApplicationId
	// the token as value
	v := request.Token

	// validate the token by cache
	if ok := s.tokenCache.Check(k, v); ok {
		return nil
	}

	// validate the token by fusion
	expire, err := s.validateByFusion(v)
	if err != nil {
		return err
	}

	// update the cache for the token
	s.tokenCache.Set(k, v, expire)

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
